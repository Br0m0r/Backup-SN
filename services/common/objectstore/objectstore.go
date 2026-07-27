// Package objectstore provides infrastructure-only access to S3-compatible
// media storage. Domain services remain responsible for key ownership rules.
package objectstore

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Store is the object storage behavior consumed by upload handlers.
type Store interface {
	Put(context.Context, string, io.Reader, int64, string) error
	Delete(context.Context, string) error
	URL(string) string
	KeyFromURL(string) (string, error)
}

// Config contains S3-compatible connection and public URL settings.
type Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseTLS        bool
	PublicBaseURL string
}

// FromEnvironment loads object storage configuration from the environment.
func FromEnvironment() (Config, error) {
	return FromLookup(os.LookupEnv)
}

// FromLookup builds Config using an environment-like lookup function.
func FromLookup(lookup func(string) (string, bool)) (Config, error) {
	required := func(name string) (string, error) {
		value, ok := lookup(name)
		value = strings.TrimSpace(value)
		if !ok || value == "" {
			return "", fmt.Errorf("%s is required", name)
		}
		return value, nil
	}

	endpoint, err := required("OBJECT_STORAGE_ENDPOINT")
	if err != nil {
		return Config{}, err
	}
	if strings.Contains(endpoint, "://") || strings.ContainsAny(endpoint, "/?#") {
		return Config{}, fmt.Errorf("OBJECT_STORAGE_ENDPOINT must be host[:port] without a URL scheme or path")
	}
	accessKey, err := required("OBJECT_STORAGE_ACCESS_KEY")
	if err != nil {
		return Config{}, err
	}
	secretKey, err := required("OBJECT_STORAGE_SECRET_KEY")
	if err != nil {
		return Config{}, err
	}
	bucket, err := required("OBJECT_STORAGE_BUCKET")
	if err != nil {
		return Config{}, err
	}
	publicBaseURL, err := required("OBJECT_STORAGE_PUBLIC_BASE_URL")
	if err != nil {
		return Config{}, err
	}
	publicBaseURL = strings.TrimRight(publicBaseURL, "/")
	if err := validatePublicBaseURL(publicBaseURL); err != nil {
		return Config{}, err
	}

	useTLS := true
	if raw, ok := lookup("OBJECT_STORAGE_USE_TLS"); ok && strings.TrimSpace(raw) != "" {
		useTLS, err = strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("OBJECT_STORAGE_USE_TLS must be true or false")
		}
	}

	return Config{
		Endpoint:      endpoint,
		AccessKey:     accessKey,
		SecretKey:     secretKey,
		Bucket:        bucket,
		UseTLS:        useTLS,
		PublicBaseURL: publicBaseURL,
	}, nil
}

// Open creates a store and verifies that its bucket is reachable.
func Open(ctx context.Context, config Config) (*S3Store, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("create object storage client: %w", err)
	}

	exists, err := client.BucketExists(ctx, config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check object storage bucket %q: %w", config.Bucket, err)
	}
	if !exists {
		return nil, fmt.Errorf("object storage bucket %q does not exist", config.Bucket)
	}
	return &S3Store{client: client, bucket: config.Bucket, publicBaseURL: config.PublicBaseURL}, nil
}

// S3Store stores media in one S3-compatible bucket.
type S3Store struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

var _ Store = (*S3Store)(nil)

// Put uploads an object with its validated media content type.
func (store *S3Store) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	_, err := store.client.PutObject(ctx, store.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

// Delete removes an object. S3 deletion is intentionally idempotent.
func (store *S3Store) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := store.client.RemoveObject(ctx, store.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

// URL returns the browser-facing URL for an object key.
func (store *S3Store) URL(key string) string {
	if validateKey(key) != nil {
		return ""
	}
	segments := strings.Split(key, "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	return store.publicBaseURL + "/" + strings.Join(segments, "/")
}

// KeyFromURL converts a URL previously returned by URL back into an object key.
func (store *S3Store) KeyFromURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid object URL: %w", err)
	}
	base, err := url.Parse(store.publicBaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid configured public base URL: %w", err)
	}
	if base.IsAbs() {
		if parsed.Scheme != base.Scheme || parsed.Host != base.Host {
			return "", fmt.Errorf("object URL is outside the configured public origin")
		}
	} else if parsed.IsAbs() || parsed.Host != "" {
		return "", fmt.Errorf("object URL must be root-relative")
	}
	prefix := strings.TrimRight(base.Path, "/") + "/"
	if !strings.HasPrefix(parsed.Path, prefix) {
		return "", fmt.Errorf("object URL is outside the configured public base URL")
	}
	key, err := url.PathUnescape(strings.TrimPrefix(parsed.Path, prefix))
	if err != nil {
		return "", fmt.Errorf("decode object key: %w", err)
	}
	if err := validateKey(key); err != nil {
		return "", err
	}
	return key, nil
}

func validateKey(key string) error {
	if key == "" || strings.HasPrefix(key, "/") || strings.Contains(key, "\\") || path.Clean(key) != key {
		return fmt.Errorf("invalid object key")
	}
	return nil
}

func validatePublicBaseURL(rawURL string) error {
	if strings.HasPrefix(rawURL, "/") && !strings.HasPrefix(rawURL, "//") {
		if parsed, err := url.Parse(rawURL); err != nil || parsed.RawQuery != "" || parsed.Fragment != "" || path.Clean(parsed.Path) != parsed.Path {
			return fmt.Errorf("OBJECT_STORAGE_PUBLIC_BASE_URL contains an invalid path")
		}
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("OBJECT_STORAGE_PUBLIC_BASE_URL must be an absolute HTTP(S) URL or a root-relative path")
	}
	return nil
}

package utils

import "testing"

func TestValidateImagePathAcceptsLocalObjectStorageAndLegacyPaths(t *testing.T) {
	paths := []*string{
		nil,
		stringPointer(""),
		stringPointer("uploads/posts/legacy.png"),
		stringPointer("/media/social-network-media/posts/users/2/image.png"),
	}
	for _, imagePath := range paths {
		if err := ValidateImagePath(imagePath); err != nil {
			t.Fatalf("expected path %v to be valid: %v", imagePath, err)
		}
	}
}

func TestValidateImagePathRejectsExternalAbsoluteAndTraversalPaths(t *testing.T) {
	paths := []string{
		"https://example.com/image.png",
		"//example.com/image.png",
		"/etc/passwd",
		"../image.png",
		"/media/social-network-media/%2e%2e/private.png",
		`C:\images\image.png`,
		"/media/",
		"/media/bucket/image.png?download=true",
	}
	for _, imagePath := range paths {
		if err := ValidateImagePath(&imagePath); err == nil {
			t.Fatalf("expected path %q to be rejected", imagePath)
		}
	}
}

func stringPointer(value string) *string {
	return &value
}

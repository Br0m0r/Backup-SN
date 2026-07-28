ARG GO_VERSION=1.24.13
ARG ALPINE_VERSION=3.22
ARG ALPINE_RUNTIME_VERSION=3.22.2

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY services/ ./services/
COPY db/migrations/ ./db/migrations/

ARG SERVICE
RUN case "${SERVICE}" in \
      auth|users|posts|groups|chat|notifications) ;; \
      *) echo "Unsupported backend service: ${SERVICE}" >&2; exit 1 ;; \
    esac && \
    CGO_ENABLED=1 go build \
      -tags sqlite_omit_load_extension \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/service ./services/${SERVICE} && \
    CGO_ENABLED=1 go build \
      -tags sqlite_omit_load_extension \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/migrate-sqlite ./services/common/cmd/migrate-sqlite && \
    if [ "${SERVICE}" = "notifications" ]; then \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/migrate ./services/notifications/cmd/migrate && \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/copy-sqlite ./services/notifications/cmd/copy-sqlite; \
    elif [ "${SERVICE}" = "chat" ]; then \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/migrate ./services/chat/cmd/migrate && \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/copy-sqlite ./services/chat/cmd/copy-sqlite && \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/copy-chat-media ./services/chat/cmd/copy-media; \
    elif [ "${SERVICE}" = "posts" ]; then \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/migrate ./services/posts/cmd/migrate && \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/copy-sqlite ./services/posts/cmd/copy-sqlite && \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/copy-post-media ./services/posts/cmd/copy-media; \
    elif [ "${SERVICE}" = "groups" ]; then \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/migrate ./services/groups/cmd/migrate && \
      CGO_ENABLED=1 go build \
        -tags sqlite_omit_load_extension \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/copy-sqlite ./services/groups/cmd/copy-sqlite; \
    fi

FROM alpine:${ALPINE_RUNTIME_VERSION}

RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 10001 app && \
    adduser -S -D -H -u 10001 -G app app && \
    mkdir -p /app/uploads /data && \
    chown -R app:app /app /data

WORKDIR /app

COPY --from=builder --chown=app:app /out/ ./
COPY --from=builder --chown=app:app /build/db/migrations/ ./db/migrations/

ARG PORT
ENV PORT=${PORT}

USER 10001:10001

EXPOSE ${PORT}
STOPSIGNAL SIGTERM

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1

ENTRYPOINT ["./service"]

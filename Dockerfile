# Stage 1: build SvelteKit frontend
FROM oven/bun:1 AS frontend
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

# Stage 2: build Go binary
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Replace placeholder with real frontend build
COPY --from=frontend /app/web/build ./internal/ui/build
RUN go build -o bin/openpadel ./cmd/server

# Final image
FROM alpine:latest
COPY --from=litestream/litestream:latest /usr/local/bin/litestream /usr/local/bin/litestream
RUN apk add --no-cache ca-certificates tzdata sqlite
WORKDIR /app
COPY --from=backend /app/bin/openpadel ./openpadel
COPY litestream.yml ./litestream.yml
COPY scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh && mkdir -p /data
EXPOSE 8080
CMD ["./entrypoint.sh"]

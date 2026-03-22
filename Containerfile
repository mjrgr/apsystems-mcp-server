FROM docker.io/golang:1.26-alpine AS builder

RUN apk add --no-cache \
    musl-dev \
    openssl-dev \
    openssl-libs-static \
    ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /apsystems-mcp-server ./cmd/server

FROM scratch

LABEL org.opencontainers.image.base.name="scratch"
LABEL org.opencontainers.image.description="A production-ready [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server written in Go that wraps the APsystems OpenAPI, giving AI assistants like Claude direct access to your solar monitoring data. Includes an optional web dashboard for visual monitoring."
LABEL org.opencontainers.image.ref.name="apsystems-mcp-server"
LABEL org.opencontainers.image.authors="Mehdi Jr-Gr"
LABEL org.opencontainers.image.title="Apsystems MCP Server"
LABEL org.opencontainers.image.vendor="Mehdi Jr-Gr"
LABEL org.opencontainers.image.source="https://github.com/mjrgr/apsystems-mcp-server"
LABEL org.opencontainers.image.licenses="Apache-2.0"

# Copy essential files for networking and TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

USER 1000:1000

COPY --from=builder --chmod=0755 /apsystems-mcp-server /apsystems-mcp-server

EXPOSE 8080 8888

ENTRYPOINT ["/apsystems-mcp-server"]

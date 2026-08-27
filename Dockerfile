FROM golang:1.26 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/udb-mysql-mcp-server .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -u 65532 mcp
COPY --from=builder /out/udb-mysql-mcp-server /app/udb-mysql-mcp-server
WORKDIR /app
USER 65532
EXPOSE 9000
ENTRYPOINT ["/app/udb-mysql-mcp-server"]
# 默认回环监听；Docker 端口映射场景需显式指定 --listen 0.0.0.0:9000（见 README）
CMD ["--transport", "sse", "--listen", "127.0.0.1:9000"]

FROM golang:1.15-alpine
COPY . /app
ENV NODE_ENV=production
ENV CGO_ENABLED=0
RUN cd /app && \
        test -e static.go || (echo "Please run go generate first"; exit 1) && \
        go vet ./... && \
        go test ./... && \
        go build -o /tmp/patrol ./cmd/patrol

FROM alpine:3.9
LABEL org.opencontainers.image.source https://github.com/karimsa/patrol
RUN apk add --no-cache \
        curl
COPY --from=0 /tmp/patrol /usr/local/bin/patrol
RUN addgroup -S patrol && \
        adduser -S patrol -G patrol && \
        mkdir /data && \
        chown -R patrol /data && \
        chmod 0755 /data
WORKDIR /data
USER patrol
ENTRYPOINT ["patrol"]

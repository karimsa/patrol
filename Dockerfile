FROM golang:1.16-alpine
COPY . /app
ENV NODE_ENV=production
ENV CGO_ENABLED=0
RUN cd /app && \
        go vet ./... && \
        go test ./... && \
        go build -o /tmp/patrol ./cmd/patrol

FROM alpine:3.9
LABEL org.opencontainers.image.source https://github.com/karimsa/patrol
RUN apk add --no-cache \
        curl \
        iputils
COPY --from=0 /tmp/patrol /usr/local/bin/patrol
RUN addgroup -S patrol && \
        adduser -S patrol -G patrol && \
        mkdir /data && \
        chown -R patrol /data && \
        chmod 0755 /data
WORKDIR /data
USER patrol
ENTRYPOINT ["patrol"]

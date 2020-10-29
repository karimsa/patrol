FROM golang:1.15-alpine
COPY . /app
ENV NODE_ENV=production
ENV CGO_ENABLED=0
RUN cd /app && \
        test -e static.go || (echo "Please run go generate first"; exit 1) && \
        go vet ./... && \
        go test ./... && \
        go build -o /tmp/patrol ./cmd/patrol

FROM alpine
LABEL org.opencontainers.image.source https://github.com/karimsa/patrol
COPY --from=0 /tmp/patrol /usr/local/bin/patrol
ENTRYPOINT ["patrol"]
CMD ["run"]

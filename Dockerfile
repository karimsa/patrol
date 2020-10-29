FROM golang:1.15-alpine
COPY . /app
RUN cd /app && \
        go generate && \
        go vet ./... && \
        go test ./... && \
        go build -o /tmp/patrol ./cmd/patrol

FROM alpine
LABEL org.opencontainers.image.source https://github.com/karimsa/patrol
COPY --from=0 /tmp/patrol /usr/local/bin/patrol
ENTRYPOINT ["patrol"]
CMD ["run"]

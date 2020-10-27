FROM golang:1.15-alpine
COPY . /app
RUN cd /app && \
        go vet ./... && \
        go test ./... && \
        go build -o /tmp/patrol ./cmd/patrol

FROM alpine
COPY --from=0 /tmp/patrol /usr/local/bin/patrol
ENTRYPOINT ["patrol"]
CMD ["run"]

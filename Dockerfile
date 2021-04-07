FROM node:14
WORKDIR /app
COPY package.json .
COPY package-lock.json .
COPY scripts scripts
COPY index.html .
COPY tailwind.config.js .
COPY postcss.config.js .
RUN npm install --silent && ./scripts/build-css.sh

FROM golang:1.16-alpine
COPY . /app
ENV NODE_ENV=production
ENV CGO_ENABLED=0
COPY --from=0 /app/dist dist
RUN cd /app && \
        go vet ./... && \
        go test ./... && \
        go build -o /tmp/patrol ./cmd/patrol

FROM alpine:3.9
LABEL org.opencontainers.image.source https://github.com/karimsa/patrol
RUN apk add --no-cache \
        curl \
        iputils
COPY --from=1 /tmp/patrol /usr/local/bin/patrol
RUN addgroup -S patrol && \
        adduser -S patrol -G patrol && \
        mkdir /data && \
        chown -R patrol /data && \
        chmod 0755 /data
WORKDIR /data
USER patrol
ENTRYPOINT ["patrol"]

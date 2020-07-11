FROM node:10

WORKDIR /app

ENV NODE_ENV=production

RUN apt-get update -yq \
		&& apt-get install -yq jq

ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

COPY packages/api/package.json packages/api/package.json
COPY packages/api/package-lock.json packages/api/package-lock.json
RUN cd packages/api && npm install --only=production

COPY packages/web/dist packages/web/dist
COPY packages/api/patrol.dist.js packages/api/patrol.dist.js
RUN chmod +x packages/api/patrol.dist.js

ENTRYPOINT ["/tini", "--", "/app/packages/api/patrol.dist.js"]

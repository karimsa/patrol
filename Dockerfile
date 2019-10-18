FROM node:10

WORKDIR /app

ENV NODE_ENV=production

COPY packages/api/package.json /app/packages/api/package.json
COPY packages/api/package-lock.json /app/packages/api/package-lock.json
COPY packages/api/patrol.dist.js /app/packages/api/patrol.dist.js
COPY packages/web/dist /app/packages/web/dist

RUN cd packages/api && npm install --only=production

ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--", "/app/packages/api/patrol.dist.js"]

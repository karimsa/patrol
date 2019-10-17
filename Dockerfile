FROM node:10

COPY . /app
WORKDIR /app

RUN npm install --silent
RUN npx @karimsa/mono run build

FROM node:10

COPY --from=0 /app/packages/web/dist
COPY --from=0 /app/packages/api/patrol.dist.js
WORKDIR /app

RUN npm install --only=production

ENTRYPOINT node packages/api/patrol.dist.js

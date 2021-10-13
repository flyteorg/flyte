FROM node:14 as builder
LABEL org.opencontainers.image.source https://github.com/lyft/flyteconsole

WORKDIR /code/flyteconsole
COPY package*.json yarn.lock ./
RUN : \
  # install production dependencies
  && yarn install --production \
  # move the production dependencies to the /app folder
  && mkdir /app \
  && mv node_modules /app \
  # install development dependencies so we can build
  && yarn install

COPY . .
RUN : \
  # build
  && make build_prod \
  # place the runtime application in /app
  && mv dist corsProxy.js index.js env.js plugins.js /app

FROM gcr.io/distroless/nodejs
LABEL org.opencontainers.image.source https://github.com/lyft/flyteconsole

COPY --from=builder /app app
WORKDIR /app
ENV NODE_ENV=production PORT=8080
EXPOSE 8080

USER 1000

CMD ["index.js"]

FROM golang:alpine AS backend-builder
RUN apk update && apk upgrade && apk add --no-cache git
WORKDIR /src
RUN mkdir /output
COPY app/ .
RUN ls -l
RUN GOOS=linux go build -o /output/server github.com/ricdeau/gitlab-extension/app/cmd/server

FROM node:12-alpine as frontend-builder
WORKDIR /front
COPY web .
ENV PATH /front/node_modules/.bin:$PATH
RUN yarn
RUN yarn run build

FROM alpine:latest
WORKDIR /app
ENV GIN_MODE=release
COPY --from=backend-builder /output .
COPY --from=frontend-builder /front/build ./www
RUN apk update \
    && apk upgrade \
    && apk add ca-certificates \
    && update-ca-certificates 2>/dev/null || true
RUN ls -l
CMD ["./boot"]
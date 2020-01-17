FROM golang:alpine AS backend-builder
RUN apk update \
    && apk upgrade \
    && apk add --no-cache git \
    && apk add --no-cache glide
WORKDIR $GOPATH/src/server
RUN mkdir /output
COPY app .
RUN glide install
RUN GOOS=linux go build -o /output/srv

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
CMD ["./srv"]
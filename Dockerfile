FROM --platform=$BUILDPLATFORM node:alpine AS front-builder
WORKDIR /app
COPY frontend/ ./
RUN npm install && npm run build

FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
ARG TARGETARCH
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
ENV GOARCH=$TARGETARCH
COPY --from=front-builder /app/dist /app/web/html

RUN apk update && apk add --no-cache \
    gcc \
    musl-dev \
    libc-dev \
    make \
    git \
    wget \
    unzip \
    bash

ENV CC=gcc

COPY . .


RUN go build -ldflags="-w -s" \
    -tags "with_quic,with_grpc,with_utls,with_acme,with_gvisor" \
    -o sui main.go

FROM --platform=$TARGETPLATFORM alpine
LABEL org.opencontainers.image.authors="alireza7@gmail.com"
ENV TZ=Asia/Tehran
ENV SUI_DB_TYPE=mysql
ENV SUI_DB_HOST=172.16.238.2
ENV SUI_DB_PORT=30409
ENV SUI_DB_USER=root
ENV SUI_DB_PASSWORD=asd123456
ENV SUI_DB_NAME=sui

WORKDIR /app
RUN apk add --no-cache --update ca-certificates tzdata
COPY --from=backend-builder /app/sui /app/
COPY entrypoint.sh /app/
ENTRYPOINT [ "./entrypoint.sh" ]
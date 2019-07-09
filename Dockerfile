FROM golang:1.12 as builder

USER root

RUN mkdir /build \
    && apt-get update -q \
    && apt-get install -q -y --no-install-recommends git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ADD . /build/

WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

FROM artifacts.msap.io/mulesoft/core-paas-base-image-ubuntu:v3.0.191

COPY --from=builder /build/main /app/

WORKDIR /app

EXPOSE 443
EXPOSE 22

CMD ["./main"]
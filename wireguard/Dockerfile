FROM --platform=$TARGETPLATFORM golang:1.21 as builder

COPY --from=project . /wireguard

WORKDIR /wireguard

RUN go mod download -x

ARG VERSION

ENV CGO_ENABLED=0

RUN go build -o ./wireguard main.go



FROM --platform=$TARGETPLATFORM alpine:latest

COPY --from=builder ./wireguard /wireguard

CMD ["/wireguard"]


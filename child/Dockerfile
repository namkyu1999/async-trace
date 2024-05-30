FROM golang:1.22.0 AS builder

ARG TARGETOS=linux
ARG TARGETARCH

COPY . /child-process
WORKDIR /child-process

ENV GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH}

RUN go env
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /output/main -v

## DEPLOY STAGE
FROM alpine:latest
WORKDIR /deploy
COPY --from=builder /output/main /deploy
CMD ["/deploy/main"]
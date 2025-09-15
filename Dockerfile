ARG BIN_NAME=ghfeed
ARG BIN_VERSION=<unknown>

FROM golang:1-alpine AS builder
ARG BIN_NAME
ARG BIN_VERSION

RUN apk --no-cache add ca-certificates

WORKDIR /src/ghfeed
COPY . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME} .

FROM scratch
ARG BIN_NAME
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/ghfeed/out/${BIN_NAME} /usr/bin/ghfeed
ENTRYPOINT ["/usr/bin/ghfeed"]

LABEL license="GPL3"
LABEL maintainer="Chris Dzombak <https://www.dzombak.com>"
LABEL org.opencontainers.image.authors="Chris Dzombak <https://www.dzombak.com>"
LABEL org.opencontainers.image.url="https://github.com/cdzombak/ghfeed"
LABEL org.opencontainers.image.documentation="https://github.com/cdzombak/ghfeed/blob/main/README.md"
LABEL org.opencontainers.image.source="https://github.com/cdzombak/ghfeed.git"
LABEL org.opencontainers.image.version="${BIN_VERSION}"
LABEL org.opencontainers.image.licenses="GPL3"
LABEL org.opencontainers.image.title="${BIN_NAME}"
LABEL org.opencontainers.image.description="GitHub activity feed consolidator"

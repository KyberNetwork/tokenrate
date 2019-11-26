FROM golang:1.12-stretch AS build-env

ENV GO111MODULE=on
COPY . /tokenrate

WORKDIR /tokenrate/usdrate/cmd/usdrate-api
RUN go build -v -o /tokenrate/bin/usdrate-api

FROM debian:stretch
COPY --from=build-env /tokenrate/bin/ /

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

CMD ["/usdrate-api", "--help"]

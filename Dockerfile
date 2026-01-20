FROM golang:1.21.13-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends wget unzip git \
    build-essential \
	&& rm -rf /var/lib/apt/lists/*

ARG GH_CI_USER=$GH_CI_USER
ARG GH_CI_TOKEN=$GH_CI_TOKEN
ARG GL_CI_USER=$GL_CI_USER
ARG GL_CI_TOKEN=$GL_CI_TOKEN

COPY . /build
WORKDIR /build
RUN make build

FROM registry.tespkg.in/library/debian:bookworm-slim

RUN apt-get update && \
      apt-get install -y --no-install-recommends ca-certificates libaio1 && \
      update-ca-certificates && \
      rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/bin/bytes-be /app/bytes-be
COPY contrib/server.yaml /app/contrib/server.yaml

WORKDIR /app
CMD [ "/app/bytes-be", "staff", "-c", "/app/contrib/server.yaml" ]

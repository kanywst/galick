FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk --no-cache add git make bash

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Get build date as an argument
ARG BUILD_DATE

# Build using make with explicit version info
# Set fixed values for version info in Docker builds
ENV VERSION="docker"
ENV COMMIT="docker"
ENV DATE=${BUILD_DATE}

# Run the build using make
RUN make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl bash

ARG VEGETA_VERSION=v12.12.0
RUN VEGETA_VERSION_CLEAN=${VEGETA_VERSION#v} && \
    wget https://github.com/tsenart/vegeta/releases/download/${VEGETA_VERSION}/vegeta_${VEGETA_VERSION_CLEAN}_linux_amd64.tar.gz && \
    tar xzf vegeta_${VEGETA_VERSION_CLEAN}_linux_amd64.tar.gz && \
    mv vegeta /usr/local/bin/ && \
    rm vegeta_${VEGETA_VERSION_CLEAN}_linux_amd64.tar.gz

COPY --from=builder /app/bin/galick /usr/local/bin/galick

COPY scripts/*.sh /scripts/
RUN chmod +x /scripts/*.sh

# Copy configuration file
COPY loadtest.yaml /loadtest.yaml

# Set working directory
WORKDIR /data

# Create additional copy in working directory
RUN cp /loadtest.yaml /data/

ENTRYPOINT ["galick"]
CMD ["run"]

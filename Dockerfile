# syntax=docker/dockerfile:1.6
ARG TF_VERSION=1.6.6

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/bin/pltf ./main.go

FROM ghcr.io/yindia/terraform-cli:latest AS release
WORKDIR /app
COPY --from=builder /app/bin/pltf /usr/local/bin/pltf
ENTRYPOINT ["pltf"]
CMD ["--help"]

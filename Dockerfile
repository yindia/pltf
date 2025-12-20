# syntax=docker/dockerfile:1.6
ARG TF_VERSION=1.5.7

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/bin/pltf ./main.go

FROM hashicorp/terraform:$TF_VERSION AS release
WORKDIR /app
COPY --from=builder /app/bin/pltf /usr/local/bin/pltf
ENTRYPOINT ["pltf"]
CMD ["--help"]

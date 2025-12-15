# syntax=docker/dockerfile:1.6

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/bin/pltf ./main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/bin/pltf /usr/local/bin/pltf
ENTRYPOINT ["pltf"]
CMD ["--help"]

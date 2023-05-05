FROM philipslabs/siderite:v0.12.2 AS siderite

FROM golang:1.19.3-alpine3.15 as builder
WORKDIR /app
RUN apk add git
COPY go.mod .
COPY go.sum .
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Build
COPY . .
RUN go build -o hsdp-docker-cp .

FROM alpine:3.17.3
COPY --from=builder /app/hsdp-docker-cp /usr/bin/hsdp-docker-cp
COPY --from=siderite /app/siderite /usr/bin/siderite

CMD ["siderite", "task"]

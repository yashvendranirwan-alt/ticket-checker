FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ticket-system ./cmd

FROM alpine:3.20
RUN adduser -D -H appuser
COPY --from=build /ticket-system /ticket-system
USER appuser
EXPOSE 8080
ENTRYPOINT ["/ticket-system"]
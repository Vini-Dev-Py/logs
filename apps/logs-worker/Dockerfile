FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
WORKDIR /src/apps/logs-worker
RUN go mod tidy && go build -o /out/app ./cmd/worker
FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/app /app/app
CMD ["/app/app"]

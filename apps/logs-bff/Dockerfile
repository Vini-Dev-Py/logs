FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
WORKDIR /src/apps/logs-bff
RUN go mod tidy && go build -o /out/app ./cmd/api && go build -o /out/seed ./cmd/seed
FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/app /app/app
COPY --from=build /out/seed /app/seed
CMD ["/app/app"]

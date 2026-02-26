FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go mod tidy && go build -o /out/app ./cmd/api
FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/app /app/app
CMD ["/app/app"]

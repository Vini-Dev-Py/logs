FROM golang:1.25-alpine AS build
WORKDIR /src
COPY . .
WORKDIR /src/apps/logs-simulator
RUN go mod tidy && go build -o /out/simulator .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/simulator /app/simulator
ENTRYPOINT ["/app/simulator"]

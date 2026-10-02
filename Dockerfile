FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod main.go main_test.go ./
RUN go test ./... && CGO_ENABLED=0 go build -o /out/notification-api .
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/notification-api /notification-api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/notification-api"]

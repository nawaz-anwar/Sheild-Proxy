FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o /bin/proxy ./main.go

FROM alpine:3.20
COPY --from=build /bin/proxy /bin/proxy
EXPOSE 8080 9090
ENTRYPOINT ["/bin/proxy"]

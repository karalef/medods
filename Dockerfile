FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app main.go

FROM gcr.io/distroless/base-debian11

COPY --from=builder /app/app /app/app

EXPOSE 8080

ENTRYPOINT ["/app/app"]

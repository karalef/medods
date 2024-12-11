FROM golang:1.23 as builder

WORKDIR /go/src/medods

COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/medods ./cmd/main.go

FROM gcr.io/distroless/base-debian11

COPY --from=builder /go/bin/medods /

EXPOSE 8080

ENTRYPOINT ["/medods"]

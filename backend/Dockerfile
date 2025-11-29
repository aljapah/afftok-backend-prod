FROM golang:1.24-alpine
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/api
FROM alpine:latest
COPY --from=0 /server /server
CMD ["/server"]

FROM golang:alpine as builder
COPY . /src
WORKDIR /src
ENV GOOS=linux
ENV CGO_ENABLED=0
RUN go build

FROM alpine
COPY --from=builder /src/trainingpeaks_bot /app/
RUN chmod +x /app/trainingpeaks_bot
WORKDIR /app
ENTRYPOINT ["/app/trainingpeaks_bot"]
FROM golang:alpine

WORKDIR /app
COPY main.go go.mod go.sum api/ ./
RUN go mod download &&\
    go build -trimpath -ldflags "-s -w" &&\
    chmod +x archiv-discord-bot
CMD ["./archiv-discord-bot"]

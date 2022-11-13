FROM golang:alpine as build

WORKDIR /app
COPY main.go go.mod go.sum ./
COPY api api
RUN apk update &&\
    apk add --no-cache build-base &&\
    go mod download &&\
    go build -trimpath -ldflags "-s -w" &&\
    chmod +x archiv-discord-bot

FROM alpine
WORKDIR /app
COPY --from=build /app/archiv-discord-bot .
CMD ["./archiv-discord-bot"]

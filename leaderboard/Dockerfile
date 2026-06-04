FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /leaderboard .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
ENV HOME_LAT=0 HOME_LON=0 \
    DUMP1090_URL=http://ultrafeeder/data/aircraft.json \
    DB_PATH=/data/leaderboard.db \
    POLL_INTERVAL=5 \
    CLEANUP_HOURS=48 \
    PORT=5000
EXPOSE 5000
VOLUME /data
COPY --from=build /leaderboard /leaderboard
ENTRYPOINT ["/leaderboard"]

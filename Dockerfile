FROM golang:latest AS build-env

WORKDIR /src

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app cmd/main.go

FROM alpine

WORKDIR /app

COPY /cmd/cert.pem /app/cert.pem

COPY /cmd/key.pem /app/key.pem

COPY --from=build-env /app .

RUN chmod +x ./app

RUN apk update && apk upgrade && apk add bash && apk add git

CMD ["./app"]
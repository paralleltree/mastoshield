FROM golang:1.22-alpine as build
ADD . /src
WORKDIR /src
RUN go build -o mastoshield cmd/proxy/main.go

FROM alpine
COPY --from=build /src/mastoshield /mastoshield/mastoshield
WORKDIR /mastoshield
EXPOSE 3000

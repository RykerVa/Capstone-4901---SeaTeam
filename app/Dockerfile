FROM golang:latest AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go build -a\
    -ldflags '-s -w -extldflags "-static"'\
    -trimpath\
    -o /envoyrouter

FROM gcr.io/distroless/base

COPY --from=build /envoyrouter /envoyrouter

ENTRYPOINT ["/envoyrouter"]

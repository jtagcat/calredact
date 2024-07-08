FROM golang:1.22 AS build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/calredact

FROM gcr.io/distroless/static-debian11
LABEL org.opencontainers.image.source="https://github.com/jtagcat/calredact"

COPY --from=build /go/bin/calredact /

VOLUME /secrets
CMD ["/calredact"]
EXPOSE 8080

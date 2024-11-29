FROM golang:bookworm AS build
LABEL authors="minz1"

WORKDIR /src

COPY go.mod go.sum ./
COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY assets assets

RUN GOOS=linux go build -ldflags '-linkmode external -extldflags "-static"' -o /bin/snitchbe ./cmd/snitchbe

FROM debian
RUN apt-get update
RUN apt-get -y install ca-certificates
COPY --from=build /bin/snitchbe /bin/snitchbe
CMD ["/bin/snitchbe"]

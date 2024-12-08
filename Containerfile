FROM golang:bookworm AS build
LABEL authors="minz1"

WORKDIR /src

COPY go.mod go.sum ./
COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY assets assets

RUN GOOS=linux go build -ldflags '-linkmode external -extldflags "-static"' -o /bin/snitchbe ./cmd/snitchbe

FROM scratch
COPY --from=build /bin/snitchbe /bin/snitchbe
CMD ["/bin/snitchbe"]

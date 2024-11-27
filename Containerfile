FROM golang:bookworm AS build
LABEL authors="minz1"

WORKDIR /src

COPY go.mod go.sum ./
COPY cmd cmd
COPY internal internal
COPY pkg pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/snitchbe ./cmd/snitchbe

FROM scratch
COPY --from=build /bin/snitchbe /bin/snitchbe
CMD ["/bin/snitchbe"]

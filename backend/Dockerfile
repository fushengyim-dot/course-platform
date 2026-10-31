FROM golang:1.27.3-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go test ./... && CGO_ENABLED=0 go build -trimpath -o /server ./cmd/server
FROM scratch
COPY --from=build /server /server
USER 65532:65532
ENTRYPOINT ["/server"]

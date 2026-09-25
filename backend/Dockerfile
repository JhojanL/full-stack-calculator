# syntax=docker/dockerfile:1

# Keep Go aligned with go.mod and ../mise.toml.
FROM golang:1.26.8-alpine AS build
WORKDIR /app
ENV CGO_ENABLED=0 GOTOOLCHAIN=local

COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY cmd ./cmd
COPY internal ./internal
RUN go build -mod=readonly -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM scratch AS runtime
COPY --from=build /out/api /api
USER 65532:65532
EXPOSE 4000
ENTRYPOINT ["/api"]
CMD ["-port=4000", "-env=production"]

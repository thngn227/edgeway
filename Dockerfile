FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
ARG COMMIT=none
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X edgeway/internal/version.Version=${VERSION} -X edgeway/internal/version.Commit=${COMMIT}" \
    -o /out/edgeway ./cmd/edgeway

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/edgeway /edgeway
COPY configs ./configs
EXPOSE 8000 8081
USER nonroot:nonroot
ENTRYPOINT ["/edgeway"]
CMD ["-config", "configs/edgeway.yaml"]

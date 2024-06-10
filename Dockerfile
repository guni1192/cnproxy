# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.22.3 AS build
WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/cnproxy .

FROM gcr.io/distroless/static-debian12:nonroot AS final

COPY --from=build /bin/cnproxy /bin/

EXPOSE 8080

ENTRYPOINT [ "/bin/cnproxy" ]

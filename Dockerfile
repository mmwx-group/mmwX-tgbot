# mmwX-tgbot — 独立 TG bot,无前端,无 PRO gate,CGO 关。
# 多阶段:golang builder → distroless static(小镜像 + 无 shell)。

FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=docker

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -trimpath \
    -ldflags="-s -w -X 'main.version=${VERSION}'" \
    -o /out/mmwx-tgbot \
    ./cmd/mmwx-tgbot

# Final: distroless static(自带 ca-certs + tzdata,无 shell)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/mmwx-tgbot /app/mmwx-tgbot

USER nonroot:nonroot

# 配置走环境变量(MMWX_TGBOT_*)或挂载 /app/config.yaml
ENV MMWX_TGBOT_HTTP_TIMEOUT_SECONDS=8

VOLUME ["/app/config"]

ENTRYPOINT ["/app/mmwx-tgbot"]
CMD ["-c", "/app/config/config.yaml"]

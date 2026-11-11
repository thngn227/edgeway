# edgeway

Edge gateway / reverse proxy viết bằng Go, lấy cảm hứng từ [Traefik](https://github.com/traefik/traefik).
Mục tiêu: hiểu sâu networking, security, observability bằng cách tự xây từng phần, và phát triển dần
từ một binary đơn lẻ lên kiến trúc Control Plane / Data Plane.

> **Trạng thái:** Stage 0 — nền móng. Xem [ROADMAP.md](ROADMAP.md).

## Bắt đầu nhanh

Yêu cầu: Go 1.26+, Docker (cho backend mẫu).

```bash
make up      # bật 2 backend whoami ở :9001, :9002
make run     # build + chạy edgeway với configs/edgeway.yaml

curl localhost:8081/ping          # OK
curl localhost:8081/api/version   # {"version":"dev",...}
curl localhost:8000/              # 404 — routing có ở v0.1

make check   # fmt + vet + test + build (giống CI)
make help    # xem mọi lệnh
```

## Cấu trúc

```
cmd/edgeway/              binary Data Plane (mỗi thư mục trong cmd/ = 1 service tương lai)
internal/
  configmodel/            ★ hợp đồng config (static + dynamic), load + validate
  dataplane/
    server/               entry point: listener, timeouts, graceful shutdown
    admin/                admin API read-only (/ping, /api/version)
  logger/                 slog JSON/text
  version/                version/commit gắn lúc build
configs/                  config mẫu: edgeway.yaml (static), dynamic.yaml (routing)
deploy/                   docker-compose cho môi trường dev
docs/
  learning/               khái niệm nền + bài tập (bắt đầu từ đây nếu mới với infra)
  architecture/           tổng quan + ảnh chụp kiến trúc từng stage
  adr/                    Architecture Decision Records — vì sao kiến trúc như vậy
```

Mới với mảng này? Đọc [docs/learning/concepts.md](docs/learning/concepts.md) rồi làm
[bài tập Stage 0](docs/learning/stage-0.md).

Package mới được thêm **khi stage cần đến**, không tạo sẵn thư mục rỗng.
Kiến trúc đích theo từng stage: [docs/architecture/overview.md](docs/architecture/overview.md).

## Hai loại config

| | Static (`configs/edgeway.yaml`) | Dynamic (`configs/dynamic.yaml`) |
|---|---|---|
| Chứa | entry points, provider, log, admin API | routers, services (sau: middlewares) |
| Áp dụng | lúc khởi động | runtime (hot reload từ v0.2) |

Key gõ sai sẽ bị từ chối ngay lúc load, không âm thầm bỏ qua.

## Observability (SigNoz) — dùng từ v0.4

edgeway chỉ phát dữ liệu chuẩn OpenTelemetry (OTLP); SigNoz là nơi nhận và hiển thị.
Cài SigNoz self-host theo hướng dẫn chính thức: <https://signoz.io/docs/install/docker/>.

## Đóng góp / quy trình

Xem [CONTRIBUTING.md](CONTRIBUTING.md): branch, commit, version, tag, cách chuyển stage.

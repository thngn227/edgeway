# 0001 — Go + modular monolith

- Trạng thái: Accepted
- Ngày: 2026-11-11
- Stage: 0

## Bối cảnh

Dự án học networking/security, chưa có kinh nghiệm mảng này. Về sau sẽ tách thành nhiều service
(Control Plane / Data Plane, rồi các service SaaS). Cần ngôn ngữ để tập trung học domain,
và cấu trúc repo cho phép tách service mà không viết lại.

## Quyết định

- Ngôn ngữ: **Go**.
- Cấu trúc: **một repo, một Go module, modular monolith** — mỗi thư mục `cmd/<name>` là một binary,
  ranh giới giữa các phần đi qua package trong `internal/`.

## Phương án đã cân nhắc

| Phương án | Ưu | Nhược |
|---|---|---|
| Go | Stdlib đủ cho proxy/TLS; hệ sinh thái cloud-native (Traefik, Caddy); học nhanh; OTel SDK trưởng thành | Có GC → p99 kém ổn định hơn Rust |
| Rust | Không GC, p99 ổn định; Pingora | Học ngôn ngữ và domain cùng lúc; ra kết quả chậm |
| Microservices ngay từ đầu | Ranh giới rõ | Vận hành phức tạp khi chưa có lý do tách |

## Hệ quả

- Có thể viết lại Data Plane bằng Rust sau này nếu cần hiệu năng; Control Plane và hợp đồng config giữ nguyên.
- Tách service = thêm `cmd/<service>` + Dockerfile, không cần đổi repo.

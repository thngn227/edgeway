# Stage 5 — Hệ sinh thái mở rộng · `v1.x → v2.0`

> Trạng thái: kiến trúc đích, chưa bắt đầu.
> Học trước: [docs/learning/stage-5.md](../learning/stage-5.md)

## Mục tiêu

Mở rộng khả năng: người dùng tự viết middleware, bảo vệ tầng ứng dụng bằng WAF, hỗ trợ TCP/UDP và HTTP/3,
tích hợp sâu với Kubernetes. Tính năng bổ sung ra trong `v1.x`; thay đổi phá vỡ hợp đồng gom vào `v2.0`.

## Sơ đồ

```
                  Provider: file · docker · controlplane · ★ kubernetes (Gateway API)
                                        │
Client ─► EntryPoint ──┬─ HTTP/1.1, HTTP/2, ★ HTTP/3 (QUIC)
                       │      └─► Router ─► Middlewares ─────────────► Service ─► Backend
                       │                     ├─ built-in
                       │                     ├─ ★ WAF (Coraza + OWASP CRS)
                       │                     └─ ★ Plugin WASM (sandbox)
                       └─ ★ TCP/UDP router (SNI, TLS passthrough) ─► TCP/UDP service
```

## Thay đổi so với Stage 4

| | |
|---|---|
| Thêm | Plugin runtime WASM · middleware WAF · router/service TCP & UDP · HTTP/3 · Kubernetes provider + operator |
| Đổi | `configmodel` tách `http` / `tcp` / `udp`; hợp đồng plugin → `apiVersion: edgeway/v2` |
| Giữ nguyên | Kiến trúc CP/DP, phân phối snapshot, multi-tenant |

Điểm đáng chú ý: Kubernetes Gateway API **lại chỉ là thêm một Provider** — controller đọc `Gateway`/`HTTPRoute`
rồi chuyển thành `configmodel.Dynamic`. Thiết kế Provider từ Stage 1 tiếp tục phát huy.

## Hạng mục

| Hạng mục | Nội dung | Version dự kiến |
|---|---|---|
| WAF | Middleware dùng Coraza + OWASP Core Rule Set; chế độ chỉ ghi nhận trước, chặn sau; chọn mức paranoia | v1.8 |
| TCP/UDP | Router `HostSNI(...)`, TLS passthrough, load balance TCP; UDP cơ bản | v1.9 |
| HTTP/3 | Entry point QUIC (vd thư viện `quic-go`), header `Alt-Svc` | v1.10 |
| Kubernetes | Provider Gateway API; operator quản lý vòng đời CP/DP; chạy conformance test của Gateway API | v1.11 |
| Plugin | Runtime WASM (vd `wazero`, thuần Go), ABI cho middleware, giới hạn bộ nhớ/CPU/thời gian, registry plugin | v2.0 |
| Config v2 | Tách giao thức trong `configmodel`; `edgewayctl migrate` chuyển config v1 → v2 | v2.0 |

## Exit criteria

- Plugin viết bằng ngôn ngữ biên dịch ra WASM chạy được, plugin lỗi/vòng lặp vô hạn không làm sập gateway.
- WAF chặn được bộ test tấn công OWASP cơ bản với tỉ lệ chặn nhầm chấp nhận được.
- Qua conformance test Gateway API (mức core).
- Config v1 chuyển sang v2 bằng công cụ, không sửa tay.

## Rủi ro / điểm khó

- Plugin = chạy code của người khác trong hot path → sandbox, giới hạn tài nguyên, đo overhead.
- WAF chặn nhầm request hợp lệ → bắt đầu ở chế độ ghi nhận, tinh chỉnh rule theo dữ liệu thật.
- HTTP/3 chạy trên UDP → firewall/LB nhiều nơi chưa hỗ trợ tốt.

## ADR cần viết khi mở stage

- Công nghệ plugin (WASM + wazero vs interpreter Go như Yaegi vs proxy-wasm ABI).
- Chiến lược phiên bản config v2 và đường nâng cấp.
- Phạm vi hỗ trợ Gateway API.

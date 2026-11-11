# Kiến trúc tổng quan

## Hai mặt phẳng

| | Data Plane (DP) | Control Plane (CP) |
|---|---|---|
| Làm gì | Xử lý **từng request** người dùng | Quản lý **config**: route, chặn ai, cert nào |
| Lưu lượng | Rất lớn | Rất nhỏ |
| Yêu cầu | Nhanh, không được chết | Đúng, an toàn |
| Số lượng | Nhiều node | 1 – vài instance |
| Binary | `edgeway` | `edgeway-controlplane` (Stage 2) |

## Các khái niệm của Data Plane

```
Client ─► EntryPoint ─► Router ─► Middlewares ─► Service (load balancer) ─► Backend
          listener,     match       auth, rate     chọn backend
          TLS, SNI      rule        limit, ...     còn sống
                ▲
         Provider (file / docker / controlplane) cung cấp dynamic config
```

- **EntryPoint**: một địa chỉ lắng nghe (`:80`, `:443`), có timeouts, TLS.
- **Router**: rule (`Host`, `PathPrefix`...) quyết định request thuộc về đâu.
- **Middleware**: biến đổi hoặc chặn request trước khi tới service.
- **Service**: nhóm backend + cách cân bằng tải + health check.
- **Provider**: nguồn dynamic config. Control Plane ở Stage 2 cũng chỉ là một provider.

## Tiến hoá theo stage

```
Stage 1   [ edgeway ]                                            1 binary
Stage 2   [ controlplane ] ──gRPC──► [ edgeway ×N ]              2 service
Stage 3   [ controlplane HA ] ─mTLS─► [ edgeway ×N ] + Redis + K8s
Stage 4   [ tenant | config | billing | analytics ] ──► DP       microservices
Stage 5   + plugins · WAF · L4 · K8s operator
```

Mỗi thư mục trong `cmd/` là một service độc lập về sau (modular monolith → microservices),
nên "tách service" chỉ là đóng gói binary riêng, không phải viết lại.

## Nguyên tắc bất biến

1. **DP không phụ thuộc sự sống của CP.** CP chết, DP chạy tiếp bằng config cuối cùng (cache ra đĩa).
2. **Hot path không gọi DB hay mạng ngoài.** Mọi thứ cần để route nằm sẵn trong RAM.
3. **`configmodel` là hợp đồng trung tâm.** File, CP, dashboard, CLI đều nói cùng một model.
4. **DP stateless.** Node chết thì thay node khác; trạng thái dùng chung (rate limit) nằm ở Redis.
5. **Bảo vệ mặt quản lý.** Admin API, CP API, kênh CP↔DP đều phải có auth/mTLS khi ra khỏi máy dev —
   chiếm được CP là điều khiển được toàn bộ traffic.
6. **Phát telemetry theo chuẩn OTel**, không gắn với một backend cụ thể (xem ADR 0003).

## Kiến trúc từng stage

| Stage | Kiến trúc | Học trước |
|---|---|---|
| 1 — Gateway một binary | [stage-1.md](stage-1.md) | [learning/stage-1.md](../learning/stage-1.md) |
| 2 — Tách Control/Data Plane | [stage-2.md](stage-2.md) | [learning/stage-2.md](../learning/stage-2.md) |
| 3 — Production-ready | [stage-3.md](stage-3.md) | [learning/stage-3.md](../learning/stage-3.md) |
| 4 — SaaS multi-tenant | [stage-4.md](stage-4.md) | [learning/stage-4.md](../learning/stage-4.md) |
| 5 — Hệ sinh thái mở rộng | [stage-5.md](stage-5.md) | [learning/stage-5.md](../learning/stage-5.md) |

Stage chưa bắt đầu là kiến trúc đích; khi mở stage sẽ xem lại, viết ADR, rồi file trở thành ảnh chụp khi stage đóng.

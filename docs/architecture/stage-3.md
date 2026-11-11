# Stage 3 — Production-ready · `v1.0`

> Trạng thái: kiến trúc đích, chưa bắt đầu.
> Học trước: [docs/learning/stage-3.md](../learning/stage-3.md)

## Mục tiêu

Đủ tin cậy để chạy traffic thật: chịu tải đo được, không điểm chết đơn lẻ, bảo mật kênh nội bộ,
triển khai lặp lại được trên Kubernetes. Kết thúc bằng `v1.0.0` — cam kết hợp đồng config ổn định.

## Yêu cầu dẫn tới stage này

- Rate limit đếm riêng trên mỗi node → sai khi có N node.
- CP chỉ có 1 instance → chết là không quản trị được.
- Kênh CP↔DP chỉ có token → cần xác thực hai chiều.
- Chưa biết hệ thống chịu được bao nhiêu tải, chậm ở đâu.

## Sơ đồ

```
                    ┌── controlplane ×2..3 (stateless) ──┐
                    │  leader (Postgres advisory lock)   │──► Postgres (primary + replica)
                    │  chạy job ACME/gia hạn             │
                    └──────────────┬─────────────────────┘
                           mTLS (CA nội bộ, cert node tự xoay vòng)
Internet ─► Cloud LB ─► ┌──────────▼───────────────────────┐
                        │ edgeway ×N  (K8s Deployment, HPA) │──► Backends
                        └───────┬───────────────┬──────────┘
                                │               │ OTLP
                          Redis (rate limit)    ▼
                                        OTel Collector ─► SigNoz (dashboard SLO, alert)
```

## Thay đổi so với Stage 2

| | |
|---|---|
| Thêm | Redis cho trạng thái dùng chung · CA nội bộ + mTLS · Helm chart · load test · fuzz test · pipeline release |
| Đổi | rate limit chuyển từ bộ nhớ sang Redis (fallback local) · CP chạy nhiều bản, job định kỳ cần leader |
| `configmodel` | thêm `apiVersion: edgeway/v1`; từ `v1.0.0` chỉ thêm field, không đổi/xoá |

## Hạng mục

| # | Hạng mục | Nội dung |
|---|---|---|
| R1 | Rate limit phân tán | Thuật toán GCRA/sliding window bằng Lua script trên Redis; Redis chết → fallback rate limit cục bộ (quyết định fail-open hay fail-closed trong ADR) |
| R2 | mTLS CP↔DP | CA nội bộ; node đổi bootstrap token lấy cert ngắn hạn; tự xoay vòng trước khi hết hạn |
| R3 | HA Control Plane | CP stateless chạy nhiều bản sau LB; job ACME/dọn dẹp chạy trên leader (advisory lock) |
| R4 | Kubernetes | Helm chart: Deployment/HPA cho gateway, PodDisruptionBudget, readiness = đã nạp config, liveness = `/ping` |
| R5 | Hiệu năng | Load test (k6 hoặc vegeta), đo p50/p99/p99.9, pprof CPU/heap, tối ưu `Transport` (idle conns), giảm allocation |
| R6 | Độ bền | Retry có ngân sách, circuit breaker, passive health check (outlier detection), timeout theo route |
| R7 | Bảo mật | Fuzz rule parser + config loader (`go test -fuzz`), `govulncheck` trong CI, threat model, admin API tắt mặc định |
| R8 | SLO + alert | Định nghĩa SLI (tỉ lệ thành công, p99), SLO, error budget; alert trên SigNoz |
| R9 | Release | goreleaser: binary đa nền tảng + image lên GHCR, SBOM, checksum; release candidate `v1.0.0-rc.N` |
| R10 | Tài liệu người dùng | Getting started, tham chiếu config, vận hành, nâng cấp |

Có thể đánh `v0.10.x`, `v0.11.x` cho các hạng mục trước khi ra `v1.0.0-rc.1`.

## Exit criteria

- Đạt mục tiêu hiệu năng đặt ra khi mở stage (vd X nghìn req/s mỗi node, p99 overhead < Y ms) — có báo cáo load test.
- Rate limit sai số trong ngưỡng chấp nhận với N node.
- Kill ngẫu nhiên 1 node gateway / 1 bản CP / Redis → không mất request ngoài ngưỡng SLO.
- Cài đặt từ đầu trên cluster mới chỉ bằng Helm + tài liệu.
- Tag `v1.0.0`.

## Rủi ro / điểm khó

- Đo hiệu năng sai (coordinated omission, máy bắn tải là nút thắt) → đọc kỹ trước khi đo.
- Redis thành điểm chết mới → fallback rõ ràng, có metric theo dõi.
- Xoay vòng cert lỗi giữa đêm → test xoay vòng bằng cert vài phút tuổi thọ.

## ADR cần viết khi mở stage

- Thuật toán và hành vi khi Redis chết của rate limit phân tán.
- PKI nội bộ và vòng đời cert node.
- Chiến lược HA của CP (leader election).
- Chính sách tương thích config sau `v1.0`.

## Lý do sang Stage 4

Đã có sản phẩm chạy ổn; muốn phục vụ nhiều khách hàng/nhóm trên cùng hệ thống, có phân quyền, đo lượng dùng và tính tiền.

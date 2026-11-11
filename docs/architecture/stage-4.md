# Stage 4 — SaaS multi-tenant · `v1.x`

> Trạng thái: kiến trúc đích, chưa bắt đầu.
> Học trước: [docs/learning/stage-4.md](../learning/stage-4.md)

## Mục tiêu

Biến edgeway thành sản phẩm SaaS: nhiều tổ chức (tenant) dùng chung Control Plane, mỗi tenant
quản lý gateway của mình, có phân quyền, đo lượng dùng, hạn mức và hoá đơn. Đây là phần gần với kinh nghiệm SaaS nhất.

## Yêu cầu dẫn tới stage này

- Nhiều nhóm/khách hàng muốn dùng, nhưng không được thấy config và traffic của nhau.
- Cần biết ai dùng bao nhiêu để giới hạn và tính tiền.

## Mô hình triển khai: hybrid (khuyến nghị, chốt bằng ADR)

```
             ┌──────────────── edgeway Cloud (mình vận hành) ────────────────┐
Tenant A ───►│ Dashboard · API ─► identity · config · distribution · usage    │
Tenant B ───►│                     billing · analytics                        │
             └───────────────▲──────────────────────────▲─────────────────────┘
                   mTLS outbound      │                   │ OTLP (usage, metrics)
             ┌──────────────┴───┐                  ┌───────┴──────────┐
             │ Hạ tầng tenant A │                  │ Hạ tầng tenant B │
             │ edgeway agent ×N │──► backend A     │ edgeway agent ×N │──► backend B
             └──────────────────┘                  └──────────────────┘
```

Gateway (DP) chạy **trong hạ tầng của tenant**, chủ động kết nối ra CP của mình. Mình không gánh traffic
của khách → chi phí thấp, dữ liệu nhạy cảm không đi qua mình. Phương án còn lại (mình chạy chung một
fleet DP cho mọi tenant, định tuyến theo hostname) ghi trong ADR để so sánh.

## Thay đổi so với Stage 3

| | |
|---|---|
| Dữ liệu | Mọi bảng có `tenant_id`; có thể bật Row-Level Security của Postgres làm lớp bảo vệ thứ hai |
| Định danh | Org → member → role (owner/admin/editor/viewer); SSO (OIDC) cho org lớn |
| DP | Đăng ký agent theo tenant (enrollment token → cert mTLS gắn tenant) |
| Telemetry | Mọi metric/span có `tenant.id`; đo lượng dùng từ đây |
| CP | Có thể tách thành nhiều service (xem dưới) |

## Tách Control Plane — chỉ khi có lý do

| Service | Trách nhiệm | Lý do tách |
|---|---|---|
| `identity` | Org, user, role, token, SSO | Bảo mật cao, thay đổi ít |
| `config` | Router/service/revision theo tenant | Lõi nghiệp vụ |
| `distribution` | Giữ hàng nghìn stream gRPC tới agent | Scale theo số kết nối, khác hẳn phần API |
| `usage` | Nhận sự kiện lượng dùng, tổng hợp theo kỳ | Ghi nhiều, xử lý bất đồng bộ |
| `billing` | Gói cước, hạn mức, hoá đơn, cổng thanh toán | Tích hợp bên ngoài, tách để cô lập lỗi |
| `analytics` | Truy vấn traffic theo tenant (ClickHouse) | Truy vấn nặng, không được làm chậm API |

Giao tiếp: gRPC đồng bộ giữa các service; sự kiện bất đồng bộ (vd `revision.published`, `usage.recorded`)
qua message broker (NATS hoặc Kafka — chốt bằng ADR) với **outbox pattern** để không mất sự kiện.
Mỗi service sở hữu dữ liệu của mình; không đọc chéo bảng.

Tiêu chí để tách một module thành service: cần scale khác biệt, cần cô lập lỗi, hoặc chu kỳ release khác biệt.
Không thoả tiêu chí nào → giữ trong monolith.

## Chi tiết từng version (dự kiến)

| Version | Nội dung |
|---|---|
| v1.1 | Org, thành viên, role; `tenant_id` khắp nơi; test cô lập dữ liệu giữa tenant |
| v1.2 | Agent mode: enrollment token, cert mTLS theo tenant, dashboard quản lý agent |
| v1.3 | Usage metering: agent gửi số request/bytes theo tenant; tổng hợp theo giờ/ngày; idempotent |
| v1.4 | Gói cước + hạn mức (số route, req/tháng); vượt hạn mức → cảnh báo rồi giới hạn |
| v1.5 | Billing: hoá đơn theo kỳ, tích hợp cổng thanh toán (chế độ test) |
| v1.6 | Analytics theo tenant: dashboard traffic, lỗi, độ trễ cho từng khách |
| v1.7 | Tách `distribution` và `usage` thành service riêng (nếu đạt tiêu chí) |

## Exit criteria

- Hai tenant dùng chung hệ thống, không đọc/sửa được dữ liệu của nhau (có test tự động chứng minh).
- Một tenant bắn tải lớn không làm chậm tenant khác (noisy neighbor).
- Hoá đơn khớp với lượng dùng đo được.

## Rủi ro / điểm khó

- Quên lọc `tenant_id` ở một truy vấn = rò dữ liệu → RLS + test cô lập trong CI.
- Đếm lượng dùng trùng/thiếu khi retry → sự kiện có ID, xử lý idempotent.
- Bùng nổ cardinality metric khi gắn `tenant.id` + route → giới hạn nhãn, tổng hợp trước khi lưu.

## ADR cần viết khi mở stage

- Mô hình triển khai (hybrid agent vs fleet dùng chung).
- Cô lập tenant ở tầng dữ liệu (cột `tenant_id` + RLS vs schema riêng vs DB riêng).
- Message broker và outbox pattern.
- Ranh giới các service khi tách CP.

## Lý do sang Stage 5

Nền tảng đã vững; mở rộng khả năng cho người dùng tự tuỳ biến và hỗ trợ thêm giao thức, môi trường.

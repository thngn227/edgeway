# Học cho Stage 4 — SaaS multi-tenant

Kiến trúc: [../architecture/stage-4.md](../architecture/stage-4.md). Nhiều phần ở đây gần với kinh nghiệm SaaS sẵn có —
điểm mới là áp vào hệ thống hạ tầng phân tán.

---

## Multi-tenancy

**Cần hiểu:** ba mô hình cô lập dữ liệu (chung bảng + `tenant_id`, schema riêng, DB riêng) và đánh đổi;
Postgres Row-Level Security; noisy neighbor và hạn mức theo tenant.

**Thực hành:** bật RLS trên một bảng thử, đặt `SET app.tenant_id = '...'` trong session, kiểm tra truy vấn không lọc vẫn chỉ thấy dữ liệu tenant đó.

**Tự kiểm tra:** RLS bảo vệ khỏi lỗi gì trong code, và không bảo vệ khỏi gì?

---

## Định danh và phân quyền

**Cần hiểu:** RBAC (role → permission), org/member/invite; OIDC và SSO; nguyên tắc quyền tối thiểu.

**Thực hành:** viết bảng ma trận role × hành động cho edgeway (ai được publish, ai chỉ xem...).

---

## Agent mode (hybrid)

**Cần hiểu:** agent chủ động kết nối ra ngoài (outbound) — vì sao dễ được firewall của khách chấp nhận;
enrollment token dùng một lần; cert gắn với tenant.

**Thực hành:** vẽ luồng đăng ký agent từ lúc khách bấm "Thêm gateway" tới khi agent nhận snapshot đầu tiên.

---

## Đo lượng dùng và billing

**Cần hiểu:** sự kiện usage có ID để xử lý idempotent; tổng hợp theo cửa sổ thời gian; độ trễ chấp nhận được;
hạn mức mềm (cảnh báo) vs cứng (chặn); làm việc với cổng thanh toán ở chế độ test.

**Thực hành:** gửi cùng một sự kiện usage 2 lần, đảm bảo tổng chỉ tăng 1 lần.

**Tự kiểm tra:** agent mất mạng 1 giờ rồi gửi bù — hệ thống tính tiền thế nào cho đúng?

---

## Tách service và giao tiếp bất đồng bộ

**Cần hiểu:** khi nào nên tách (scale khác biệt, cô lập lỗi, chu kỳ release); database-per-service;
message broker (NATS, Kafka); at-least-once delivery; outbox pattern; tracing xuyên qua message.

**Thực hành:**
1. Chạy NATS hoặc Kafka bằng Docker, publish/consume sự kiện `revision.published`.
2. Cài outbox: ghi sự kiện vào bảng `outbox` trong cùng transaction với dữ liệu, worker đẩy ra broker.
3. Propagate trace context qua header của message, xem trace trên SigNoz nối qua broker.

**Tự kiểm tra:** vì sao không publish sự kiện trực tiếp sau khi commit transaction?

---

## Analytics theo tenant

**Cần hiểu:** ClickHouse (lưu dạng cột, phù hợp truy vấn tổng hợp); cardinality của metric và chi phí lưu trữ.

**Thực hành:** chạy ClickHouse bằng Docker, nạp 1 triệu dòng log giả, so tốc độ đếm theo tenant/route với Postgres.

---

## Đọc thêm

- [AWS — SaaS Tenant Isolation Strategies (whitepaper)](https://docs.aws.amazon.com/whitepapers/latest/saas-tenant-isolation-strategies/saas-tenant-isolation-strategies.html)
- [microservices.io — Transactional outbox](https://microservices.io/patterns/data/transactional-outbox.html)
- [PostgreSQL — Row Security Policies](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)

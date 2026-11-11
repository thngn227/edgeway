# Roadmap

**Hiện tại: Stage 0 — Nền móng** · Version mới nhất: _chưa có tag_

| Stage | Tên                      | Versions    | Trạng thái | Kiến trúc | Học |
|-------|--------------------------|-------------|------------|-----------|-----|
| 0     | Nền móng                 | v0.0.x      | 🚧 Đang làm | [overview](docs/architecture/overview.md) | [stage-0](docs/learning/stage-0.md) |
| 1     | Gateway một binary       | v0.1 – v0.5 | ⏳ Chưa | [stage-1](docs/architecture/stage-1.md) | [stage-1](docs/learning/stage-1.md) |
| 2     | Tách Control/Data Plane  | v0.6 – v0.9 | ⏳ Chưa | [stage-2](docs/architecture/stage-2.md) | [stage-2](docs/learning/stage-2.md) |
| 3     | Production-ready         | v1.0        | ⏳ Chưa | [stage-3](docs/architecture/stage-3.md) | [stage-3](docs/learning/stage-3.md) |
| 4     | SaaS multi-tenant        | v1.x        | ⏳ Chưa | [stage-4](docs/architecture/stage-4.md) | [stage-4](docs/learning/stage-4.md) |
| 5     | Hệ sinh thái mở rộng     | v1.x → v2.0 | ⏳ Chưa | [stage-5](docs/architecture/stage-5.md) | [stage-5](docs/learning/stage-5.md) |

Khái niệm nền: [docs/learning/concepts.md](docs/learning/concepts.md).
Cách chuyển stage: [CONTRIBUTING.md](CONTRIBUTING.md#chuyển-stage).

Kế hoạch cho stage chưa bắt đầu là **kiến trúc đích**, sẽ được xem lại khi mở stage (viết ADR, chỉnh version).

---

## Stage 0 — Nền móng · `v0.0.x`

Mục tiêu: hiểu hành trình một request; có repo sạch, chuẩn, chạy được để mọi stage sau dựng lên trên.

**Học trước** — [docs/learning/stage-0.md](docs/learning/stage-0.md):

- [ ] Bài 1 — DNS (`dig`)
- [ ] Bài 2 — Request HTTP từ A đến Z (`curl -v`)
- [ ] Bài 3 — TLS và chứng chỉ (`openssl s_client`)
- [ ] Bài 4 — Dựng reverse proxy bằng Docker
- [ ] Bài 5 — Cài SigNoz, làm quen giao diện

**Repo:**

- [x] Go module, cấu trúc `cmd/` + `internal/`
- [x] `configmodel`: static + dynamic (bản nháp), load + validate, test
- [x] Entry point có timeouts + graceful shutdown, test
- [x] Admin API: `/ping`, `/api/version`
- [x] Makefile, Dockerfile, docker-compose (backend whoami)
- [x] CI GitHub Actions (`make check`)
- [x] ROADMAP, CHANGELOG, CONTRIBUTING, ADR 0001–0003
- [x] Kiến trúc đích + tài liệu học cho Stage 1 → 5
- [ ] Commit đầu tiên + push lên GitHub, CI xanh
- [ ] Tạo milestones v0.1 → v0.5 + labels trên GitHub
- [ ] Tag `v0.0.1`

**Exit criteria:** xong 5 bài tập · CI xanh trên GitHub · `make up && make run` chạy được · SigNoz chạy local.

---

## Stage 1 — Gateway một binary · `v0.1 → v0.5`

Mục tiêu: proxy thật, tương đương Traefik cơ bản. Chặng học networking.

| Version | Nội dung |
|---------|----------|
| v0.1 | Proxy tĩnh: rule `Host`/`PathPrefix`, priority, round-robin, `X-Forwarded-*` |
| v0.2 | File provider watch + hot reload (atomic swap), health check |
| v0.3 | TLS: SNI, cert tự ký → ACME, redirect HTTP→HTTPS, HSTS |
| v0.4 | OpenTelemetry: traces, metrics, access log → SigNoz · `examples/shop/` (order → inventory, payment) gắn OTel để có **service map** nhiều node, một service cố tình chậm/lỗi để luyện tìm chỗ nghẽn |
| v0.5 | Middlewares: rate limit, IP allowlist, JWT/forward-auth, body limit; Docker provider |

**Exit criteria:** proxy 3 backend qua HTTPS · sửa config không rớt request · service map trên SigNoz hiện gateway + 3 service shop · tìm được service chậm trong 5 phút.
**Lý do sang Stage 2:** cần quản lý nhiều node gateway + UI/API thay vì sửa YAML từng máy.

## Stage 2 — Tách Control Plane / Data Plane · `v0.6 → v0.9`

| Version | Nội dung |
|---------|----------|
| v0.6 | `edgeway-controlplane` API + Postgres + revision (publish/rollback); gateway thêm provider "controlplane" (polling) |
| v0.7 | gRPC streaming đẩy snapshot; gateway cache config ra đĩa; heartbeat node |
| v0.8 | Dashboard (React) + CLI `edgewayctl` + đăng nhập admin, API token |
| v0.9 | Cert manager tập trung, private key mã hoá khi lưu |

**Exit criteria:** tắt CP gateway vẫn phục vụ · publish trên dashboard → mọi node áp dụng < 1s · rollback bằng 1 lệnh.
**Lý do sang Stage 3:** chạy thật cho người dùng → rate limit chính xác đa node, chịu tải, HA, kênh nội bộ an toàn.

## Stage 3 — Production-ready · `v1.0`

Redis rate limit phân tán · mTLS CP↔DP · HA control plane · Helm chart · load test có số liệu ·
retry/circuit breaker · fuzz + govulncheck · SLO + alert · release tự động · tài liệu người dùng.
Có thể đi qua `v0.10.x`, `v0.11.x`, `v1.0.0-rc.N` trước `v1.0.0`.

**Exit criteria:** đạt mục tiêu hiệu năng (định nghĩa khi mở stage) · kill ngẫu nhiên 1 thành phần không vi phạm SLO · cài từ đầu bằng Helm · `v1.0.0` = config ổn định.
**Lý do sang Stage 4:** phục vụ nhiều khách hàng/nhóm, cần phân quyền, đo lượng dùng, tính tiền.

## Stage 4 — SaaS multi-tenant · `v1.x`

| Version | Nội dung |
|---------|----------|
| v1.1 | Org, thành viên, role; `tenant_id` khắp nơi; test cô lập |
| v1.2 | Agent mode (gateway chạy ở hạ tầng tenant, kết nối ra CP) |
| v1.3 | Usage metering idempotent |
| v1.4 | Gói cước + hạn mức |
| v1.5 | Billing (cổng thanh toán chế độ test) |
| v1.6 | Analytics theo tenant |
| v1.7 | Tách `distribution`, `usage` thành service riêng (nếu đạt tiêu chí) |

**Exit criteria:** 2 tenant cô lập dữ liệu (có test) · không noisy neighbor · hoá đơn khớp lượng dùng.
**Lý do sang Stage 5:** nền tảng vững, mở rộng khả năng tuỳ biến và giao thức.

## Stage 5 — Hệ sinh thái mở rộng · `v1.x → v2.0`

| Version | Nội dung |
|---------|----------|
| v1.8 | WAF (Coraza + OWASP CRS) |
| v1.9 | TCP/UDP router, TLS passthrough theo SNI |
| v1.10 | HTTP/3 (QUIC) |
| v1.11 | Kubernetes Gateway API provider + operator |
| v2.0 | Plugin WASM · config v2 + `edgewayctl migrate` |

**Exit criteria:** plugin lỗi không làm sập gateway · WAF qua bộ test OWASP cơ bản · qua conformance Gateway API (core) · migrate v1 → v2 tự động.

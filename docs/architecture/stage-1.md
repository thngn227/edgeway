# Stage 1 — Gateway một binary · `v0.1 → v0.5`

> Trạng thái: kiến trúc đích, chưa bắt đầu. Khi Stage 1 đóng, file này giữ nguyên làm ảnh chụp lịch sử.
> Học trước: [docs/learning/stage-1.md](../learning/stage-1.md)

## Mục tiêu

Một reverse proxy thật, tương đương Traefik bản cơ bản: định tuyến, cân bằng tải, TLS, hot reload,
middleware bảo mật, telemetry. Đây là chặng học networking.

## Sơ đồ

```
            ┌────────────────────── edgeway (1 process) ───────────────────────┐
Client ───► │ EntryPoint ─► TLS ─► Router ─► Middleware chain ─► Service (LB)  │ ───► Backends
            │                         ▲                                         │
            │   File/Docker provider ─► Watcher ─► build runtime ─► atomic swap │
            │                                                                   │
            │   Admin API (:8081)            OTel SDK ──────────────────────────┼──► Collector ─► SigNoz
            └───────────────────────────────────────────────────────────────────┘
```

## Thay đổi so với Stage 0

| | |
|---|---|
| Thêm | router, service (LB + reverse proxy), provider, runtime, tls, telemetry, middleware, `examples/shop` |
| Đổi | entry point thôi trả 404, chuyển request vào runtime hiện hành |
| Giữ nguyên | `configmodel` (chỉ mở rộng field), server, admin, logger |

## Package (thêm dần theo version)

| Package | Version | Trách nhiệm |
|---|---|---|
| `internal/configmodel` | có sẵn | Static + dynamic model, load, validate |
| `internal/dataplane/server` | có sẵn | Listener, timeouts, graceful shutdown |
| `internal/dataplane/admin` | có sẵn | Admin API read-only |
| `internal/dataplane/router` | v0.1 | Parse rule, match theo priority |
| `internal/dataplane/service` | v0.1 | Reverse proxy, round-robin; health check ở v0.2 |
| `internal/dataplane/provider` | v0.2 | Interface `Provider`, file provider có watch; Docker provider ở v0.5 |
| `internal/dataplane/runtime` | v0.2 | Build handler từ `Dynamic`, atomic swap |
| `internal/dataplane/tls` | v0.3 | Cert store, SNI, ACME |
| `internal/telemetry` | v0.4 | Khởi tạo OTel tracer/meter/logger (dùng chung cho gateway và `examples/shop`) |
| `examples/shop/` | v0.4 | 3 service demo (order, inventory, payment) gắn OTel, chạy sau gateway |
| `internal/dataplane/middleware` | v0.5 | ratelimit, ipallow, auth, headers, bodylimit |

## Chi tiết từng version

### v0.1 — Proxy tĩnh
- Rule `Host(...)`, `PathPrefix(...)`, `Path(...)`, kết hợp `&&` `||`. Priority mặc định = độ dài rule (giống Traefik).
- `httputil.ReverseProxy` với `Rewrite`: set `X-Forwarded-For/Proto/Host`, bỏ hop-by-hop headers.
- Round-robin trên danh sách server; `http.Transport` dùng chung theo service (connection pool, timeouts).
- Không router nào khớp → 404; backend lỗi → 502; backend quá thời gian → 504.
- Test: unit cho parser/matcher, integration bằng `httptest.Server`.

### v0.2 — Config động
- `Provider` đẩy `Dynamic` vào channel; file provider dùng `fsnotify`, debounce ~200ms.
- Runtime build mới từ `Dynamic` → `atomic.Pointer[http.Handler]`; config lỗi → log và giữ bản cũ.
- Health check chủ động: `GET path` theo interval; server down bị loại khỏi LB, sống lại thì đưa vào.
- Admin API thêm `GET /api/rawdata` (config đang chạy) và trạng thái health của từng server.

### v0.3 — TLS
- Entry point có `tls`: cert từ file, chọn theo SNI qua `tls.Config.GetCertificate`.
- ACME (Let's Encrypt staging trước) bằng `autocert`, challenge HTTP-01.
- Middleware/entry point redirect HTTP → HTTPS; header HSTS.
- Mặc định an toàn: TLS 1.2+, cipher suites của Go.

### v0.4 — Observability
- `internal/telemetry`: TracerProvider + MeterProvider + log bridge, OTLP gRPC, tắt hẳn nếu không cấu hình endpoint.
- Span cho mỗi request (attribute: router, service, server), propagate `traceparent` xuống backend.
- Metric RED theo router/service; access log JSON có `trace_id`.
- `examples/shop/`: order → inventory, payment; inventory chậm/lỗi ngẫu nhiên có cấu hình được.

### v0.5 — Security middlewares + Docker provider
- Chuỗi middleware khai báo trong dynamic config, gắn vào router.
- `rateLimit` (token bucket theo IP, có `trustedProxies`), `ipAllowList`, `basicAuth`/`forwardAuth`/JWT, `headers`, `bodyLimit`.
- Docker provider: đọc label container (`edgeway.http.routers.<name>.rule=...`).
- Metric số request bị chặn theo từng middleware.

## Service map mục tiêu (v0.4)

```
client ─► edgeway ─► order ─┬─► inventory   (cố tình chậm/lỗi ngẫu nhiên)
                            └─► payment
mọi service truyền `traceparent` → SigNoz nối thành một trace và vẽ service map
```

## Luồng hot reload (v0.2)

```
file đổi → provider gửi Dynamic vào channel → debounce → validate
   ✗ lỗi: log, giữ config cũ
   ✓ ok : build runtime mới → atomic.Pointer.Store → request mới dùng runtime mới,
          request đang chạy hoàn tất trên runtime cũ
```

## Interface chính (dự kiến)

```go
type Provider interface {
    Provide(ctx context.Context, out chan<- configmodel.Dynamic) error
}

type Middleware func(next http.Handler) http.Handler
```

## Exit criteria

- Proxy 3 backend qua HTTPS, cert tự động.
- Sửa dynamic config khi đang bắn tải → không request nào lỗi.
- Service map trên SigNoz hiện gateway + 3 service shop; tìm ra service chậm trong 5 phút.
- Rate limit trả 429 đúng ngưỡng; header `X-Forwarded-For` giả từ client không qua được.

## Rủi ro / điểm khó

- Rò rỉ goroutine khi reload (health checker của runtime cũ phải dừng) → test với `-race` và đếm goroutine.
- Streaming / WebSocket qua proxy: không được đặt `WriteTimeout` cứng, phải flush đúng.
- Rule parser là đầu vào không tin cậy → test kỹ, fuzz ở Stage 3.

## ADR cần viết

- Cú pháp rule (tự viết parser hay dùng thư viện expression).
- Chiến lược hot reload (snapshot toàn bộ + atomic swap).
- Thư viện ACME (`autocert` hay `lego`).

## Lý do sang Stage 2

Nhiều node gateway thì sửa YAML từng máy không còn ổn; cần API + UI để quản lý tập trung và biết node nào đang chạy config nào.

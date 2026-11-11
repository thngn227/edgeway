# Học cho Stage 1 — Gateway một binary

Kiến trúc: [../architecture/stage-1.md](../architecture/stage-1.md). Mỗi phần học ngay **trước** version tương ứng,
không cần học hết từ đầu.

---

## v0.1 — Reverse proxy hoạt động thế nào

**Cần hiểu:**
- `httputil.ReverseProxy`: nhận request, tạo request mới tới backend, copy response về.
- Hop-by-hop headers (`Connection`, `Keep-Alive`, `Transfer-Encoding`, `Upgrade`...) — chỉ có nghĩa trên một chặng, proxy phải bỏ.
- `X-Forwarded-For / Proto / Host` và header chuẩn `Forwarded`.
- Connection pool của `http.Transport`: vì sao tái sử dụng kết nối tới backend quan trọng.
- Mã lỗi của proxy: 404 (không router), 502 (backend lỗi), 504 (backend quá hạn).
- Load balancing: round-robin, weighted, least-connections.

**Thực hành:**
1. Đọc source `net/http/httputil/reverseproxy.go` (chạy `go env GOROOT` để tìm), tìm chỗ xoá hop-by-hop headers.
2. Viết proxy 20 dòng bằng `httputil.NewSingleHostReverseProxy` trỏ vào whoami, so sánh header với Bài 4 Stage 0.
3. Tắt một backend whoami rồi gọi proxy: thấy mã lỗi gì?

**Tự kiểm tra:** vì sao proxy không được chuyển nguyên header `Connection` của client sang backend?

---

## v0.2 — Concurrency và hot reload

**Cần hiểu:**
- Goroutine, channel, `context` để huỷ; `sync.WaitGroup`.
- `atomic.Pointer` — đọc không cần lock trên hot path.
- Data race và `go test -race`.
- Health check chủ động (gọi định kỳ) vs thụ động (quan sát lỗi thật).
- Debounce: gom nhiều sự kiện liên tiếp thành một.

**Thực hành:**
1. Viết chương trình watch một file bằng `fsnotify`, in ra mỗi lần đổi; mở file bằng editor rồi lưu — có bao nhiêu sự kiện?
2. Cố tình tạo data race (2 goroutine ghi cùng map), chạy `go run -race`.
3. Đếm goroutine bằng `runtime.NumGoroutine()` trước/sau khi dừng một worker để phát hiện rò rỉ.

**Tự kiểm tra:** request đang chạy dở khi config đổi sẽ dùng config nào? Vì sao không bị lỗi?

---

## v0.3 — TLS trong thực tế

**Cần hiểu:**
- TLS handshake (1.2 vs 1.3), certificate chain, CA.
- SNI: một IP phục vụ nhiều domain, chọn cert theo tên.
- ACME và Let's Encrypt: challenge HTTP-01 vs DNS-01; môi trường staging để tránh bị giới hạn tần suất.
- HSTS và redirect HTTP → HTTPS.

**Thực hành:**
1. Tạo cert tự ký:
   `openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 30 -subj "/CN=api.local"`
2. Chạy server Go `ListenAndServeTLS` với cert đó, gọi `curl -v --resolve api.local:8443:127.0.0.1 https://api.local:8443` — xem lỗi, rồi thêm `-k` / `--cacert cert.pem`.
3. Tạo 2 cert cho 2 domain, dùng `GetCertificate` chọn theo SNI, kiểm tra bằng `openssl s_client -servername`.

**Tự kiểm tra:** vì sao challenge HTTP-01 không dùng được cho domain wildcard `*.example.com`?

---

## v0.4 — OpenTelemetry

**Cần hiểu:**
- TracerProvider, MeterProvider, Exporter (OTLP), Resource (`service.name`).
- Context propagation và header `traceparent` (W3C Trace Context).
- Semantic conventions cho HTTP (`http.request.method`, `http.response.status_code`, `url.path`...).
- Metric types: counter, histogram; vì sao latency dùng histogram.
- Sampling: giữ bao nhiêu phần trăm trace, ở đâu (head vs tail).

**Thực hành:**
1. Làm theo "Getting started" Go của OpenTelemetry, gửi trace tới SigNoz.
2. Hai chương trình Go nhỏ: A gọi B qua HTTP, có/không propagate context — so sánh trace trên SigNoz.
3. Tìm service map trên SigNoz; tắt propagate ở một service và xem map bị đứt.

**Tự kiểm tra:** trace và log liên kết với nhau bằng gì?

---

## v0.5 — Bảo mật ở gateway

**Cần hiểu:**
- Rate limit: token bucket, leaky bucket, fixed/sliding window — khác nhau ở đâu.
- Trusted proxies: chỉ tin `X-Forwarded-For` từ IP đã biết.
- Xác thực: Basic, JWT (header.payload.signature, kiểm tra `exp`, `aud`, thuật toán), forward-auth.
- Security headers: HSTS, `X-Content-Type-Options`, CSP.
- Docker labels như một nguồn config.

**Thực hành:**
1. Dùng `golang.org/x/time/rate` viết limiter 5 req/s, bắn bằng vòng lặp `curl` hoặc công cụ tải (vd `hey`, `vegeta` — cần cài), đếm số 429.
2. Giải mã một JWT trên jwt.io, đổi một ký tự payload rồi kiểm chữ ký.
3. `docker inspect` một container có label, đọc label bằng Docker SDK cho Go.

**Tự kiểm tra:** rate limit theo IP bị qua mặt thế nào nếu tin mọi `X-Forwarded-For`?

---

## Đọc thêm

- Go — package `net/http`, `net/http/httputil`, `crypto/tls`
- [Traefik docs](https://doc.traefik.io/traefik/) — đối chiếu khái niệm EntryPoint/Router/Middleware/Service
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Let's Encrypt — challenge types](https://letsencrypt.org/docs/challenge-types/)

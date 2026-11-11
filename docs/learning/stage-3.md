# Học cho Stage 3 — Production-ready

Kiến trúc: [../architecture/stage-3.md](../architecture/stage-3.md).

---

## Rate limit phân tán

**Cần hiểu:** vì sao đếm cục bộ sai khi có N node; Redis và Lua script (thao tác nguyên tử);
GCRA, sliding window log/counter; fail-open vs fail-closed khi Redis chết.

**Thực hành:** chạy Redis bằng Docker, viết script Lua tăng bộ đếm có TTL, gọi từ 2 tiến trình Go cùng lúc, so với đếm cục bộ.

**Tự kiểm tra:** Redis chết thì gateway nên cho qua hết hay chặn hết? Phụ thuộc vào điều gì?

---

## PKI nội bộ và mTLS

**Cần hiểu:** CA tự dựng; cert server vs cert client; mTLS hai chiều; cert ngắn hạn và xoay vòng; thu hồi.

**Thực hành:**
1. Bằng `openssl`: tạo CA, ký cert server và cert client.
2. Server Go với `tls.Config{ClientAuth: tls.RequireAndVerifyClientCert}`; gọi bằng `curl --cert client.pem --key client-key.pem --cacert ca.pem`, rồi gọi không có cert client.

**Tự kiểm tra:** vì sao cert ngắn hạn + xoay vòng tự động an toàn hơn cert 1 năm?

---

## Kubernetes cơ bản

**Cần hiểu:** Pod, Deployment, Service, Ingress, ConfigMap/Secret, readiness vs liveness probe,
HPA, PodDisruptionBudget, Helm chart và values.

**Thực hành** (cần cài `kind`, `kubectl`, `helm`):
1. Tạo cluster local bằng `kind`, deploy whoami với 3 replica + Service, gọi thử.
2. Xoá một pod, quan sát Deployment tạo lại.
3. Viết Helm chart tối thiểu cho edgeway từ các manifest đã thử.

**Tự kiểm tra:** gateway nên báo "ready" khi nào — lúc process chạy hay lúc đã nạp config?

---

## Đo hiệu năng

**Cần hiểu:** throughput vs latency; percentile (p50/p99/p99.9) và vì sao trung bình đánh lừa;
coordinated omission; máy bắn tải là nút thắt; pprof CPU/heap/goroutine; benchmark Go (`testing.B`).

**Thực hành** (cần cài `k6` hoặc `vegeta`):
1. Bắn tải thẳng vào whoami, rồi qua edgeway — overhead của gateway là bao nhiêu?
2. Bật `net/http/pprof` trên admin port, lấy CPU profile trong lúc bắn tải, mở bằng `go tool pprof -http`.

**Tự kiểm tra:** p99 = 50ms nghĩa là gì với người dùng gọi 20 request mỗi lần tải trang?

---

## Độ bền và bảo mật

**Cần hiểu:** retry budget (retry có giới hạn để không tự DDoS), circuit breaker, outlier detection;
fuzz testing; quét lỗ hổng phụ thuộc; threat modeling (STRIDE).

**Thực hành:**
1. Viết fuzz test cho rule parser: `go test -fuzz=FuzzParseRule`.
2. Chạy `govulncheck ./...` (cần cài).
3. Liệt kê theo STRIDE các mối đe doạ cho kênh CP↔DP.

---

## SLO

**Cần hiểu:** SLI, SLO, error budget; alert theo tốc độ đốt error budget thay vì theo ngưỡng cứng.

**Thực hành:** định nghĩa SLO "99.9% request thành công trong 30 ngày", tính error budget ra phút, dựng dashboard + alert trên SigNoz.

---

## Đọc thêm

- [Google SRE Book — Service Level Objectives](https://sre.google/sre-book/service-level-objectives/)
- [Kubernetes — Concepts](https://kubernetes.io/docs/concepts/)
- [Go — Fuzzing](https://go.dev/doc/security/fuzz/)
- [Go — Diagnostics (pprof)](https://go.dev/doc/diagnostics)

# Học cho Stage 5 — Hệ sinh thái mở rộng

Kiến trúc: [../architecture/stage-5.md](../architecture/stage-5.md).

---

## WAF

**Cần hiểu:** OWASP Top 10 (SQL injection, XSS, path traversal...); WAF dựa trên rule; OWASP Core Rule Set,
anomaly scoring, mức paranoia; chặn nhầm (false positive) và chế độ chỉ ghi nhận.

**Thực hành:** dựng Coraza làm middleware trước whoami, gửi request chứa `' OR 1=1 --` và `<script>`, đọc log rule nào bắt.

**Tự kiểm tra:** vì sao nên chạy chế độ ghi nhận vài tuần trước khi bật chặn?

---

## TCP/UDP (L4) vs HTTP (L7)

**Cần hiểu:** L4 chỉ thấy kết nối và byte, L7 thấy request/header; TLS passthrough — định tuyến theo SNI
mà không giải mã; giới hạn của L4 (không có path, header, retry theo request).

**Thực hành:** viết TCP proxy bằng `net.Listen` + `io.Copy` hai chiều; đọc SNI từ ClientHello để chọn backend (gợi ý: `tls.Server` với `GetConfigForClient`, hoặc tự parse).

---

## HTTP/3 và QUIC

**Cần hiểu:** QUIC chạy trên UDP; hết head-of-line blocking ở tầng TCP; handshake 0-RTT/1-RTT;
header `Alt-Svc` để trình duyệt nâng cấp lên HTTP/3.

**Thực hành:** chạy server HTTP/3 mẫu của `quic-go`, gọi bằng `curl --http3` (cần bản curl hỗ trợ) và xem trong DevTools trình duyệt.

---

## Kubernetes controller và Gateway API

**Cần hiểu:** reconcile loop (so trạng thái mong muốn với thực tế rồi sửa); informer/watch; CRD;
Gateway API: `GatewayClass`, `Gateway`, `HTTPRoute` và phân vai người vận hành / người phát triển.

**Thực hành:** dùng `kind` + manifest Gateway API mẫu; viết controller nhỏ (vd bằng `controller-runtime`) in ra mỗi khi `HTTPRoute` thay đổi.

**Tự kiểm tra:** vì sao Gateway API chỉ cần thêm một Provider trong edgeway?

---

## Plugin WASM

**Cần hiểu:** WebAssembly ngoài trình duyệt; sandbox và giới hạn tài nguyên; ABI giữa host và plugin;
chi phí gọi qua ranh giới host ↔ WASM; so sánh với interpreter Go (Yaegi) mà Traefik dùng.

**Thực hành:** dùng `wazero` nạp một module WASM (biên dịch từ Go hoặc Rust), gọi hàm của nó từ Go; cho module chạy vòng lặp vô hạn và chặn bằng timeout.

---

## Tiến hoá API và config

**Cần hiểu:** semantic versioning cho hợp đồng; deprecation có thời hạn; công cụ migrate tự động.

**Thực hành:** viết hàm chuyển một file config v1 sang v2 và test bằng toàn bộ ví dụ trong repo.

---

## Đọc thêm

- [OWASP Top 10](https://owasp.org/www-project-top-ten/) · [OWASP Core Rule Set](https://coreruleset.org/)
- [Coraza](https://coraza.io/)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
- [wazero](https://wazero.io/)
- [RFC 9000 — QUIC](https://www.rfc-editor.org/rfc/rfc9000)

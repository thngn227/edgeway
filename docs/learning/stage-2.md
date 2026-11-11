# Học cho Stage 2 — Tách Control Plane / Data Plane

Kiến trúc: [../architecture/stage-2.md](../architecture/stage-2.md).

---

## v0.6 — API, database, phiên bản config

**Cần hiểu:**
- Thiết kế REST cho tài nguyên config; OpenAPI.
- Postgres: transaction, constraint, index; migration chỉ tiến.
- Snapshot bất biến + số revision tăng dần; publish/rollback.
- Conditional request: `ETag` / `If-None-Match` → `304 Not Modified`.

**Thực hành:**
1. Chạy Postgres bằng Docker, viết migration tạo bảng `revisions`, chèn 3 revision, truy vấn bản mới nhất.
2. Viết endpoint trả `ETag`, gọi `curl -H 'If-None-Match: "3"'` và xem `304`.

**Tự kiểm tra:** vì sao lưu snapshot nguyên khối thay vì chỉ lưu từng thay đổi?

---

## v0.7 — gRPC streaming và tính sẵn sàng

**Cần hiểu:**
- Protobuf, gRPC; 4 kiểu RPC (unary, server stream, client stream, bidi).
- Reconnect với exponential backoff + jitter (tránh mọi node kết nối lại cùng lúc).
- Eventual consistency: các node có thể lệch nhau vài trăm ms.
- "Static stability": DP vẫn chạy khi CP chết.

**Thực hành:**
1. Viết service gRPC "đếm giờ": server stream trả số tăng mỗi giây; client in ra. Tắt server giữa chừng, cho client tự reconnect.
2. Gọi thử bằng `grpcurl` (cần cài).
3. Mô phỏng CP chết: gateway khởi động lại có còn chạy bằng cache không?

**Tự kiểm tra:** gửi snapshot đầy đủ vs delta — đánh đổi gì?

---

## v0.8 — Dashboard, CLI, xác thực

**Cần hiểu:**
- Session cookie vs JWT cho dashboard; CSRF; cookie `HttpOnly`, `Secure`, `SameSite`.
- API token có phạm vi (scope) và hạn dùng; chỉ lưu hash của token.
- CLI kiểu "declarative": `apply -f` + `diff` (giống `kubectl`).
- Gọi API từ React (phần quen thuộc), xử lý lỗi và trạng thái tải.

**Thực hành:**
1. Thiết kế UX `edgewayctl diff` trước khi code: in ra gì khi thêm/sửa/xoá một router?
2. Đọc tài liệu OWASP về CSRF và áp vào API của CP.

**Tự kiểm tra:** vì sao không lưu API token dạng nguyên văn trong DB?

---

## v0.9 — Quản lý cert và secret

**Cần hiểu:**
- Vòng đời cert: xin, gia hạn, thu hồi; giới hạn tần suất của Let's Encrypt.
- HTTP-01 khi có nhiều node: node nào trả lời challenge?
- Mã hoá khi lưu (envelope encryption): khoá dữ liệu được mã hoá bằng khoá chính.

**Thực hành:**
1. Dùng Let's Encrypt **staging** xin cert cho một domain thử (cần domain thật trỏ về máy có IP public, hoặc dùng DNS-01).
2. Mã hoá/giải mã một chuỗi bằng AES-GCM trong Go; đổi một byte ciphertext và xem giải mã thất bại.

**Tự kiểm tra:** private key bị lộ thì phải làm những gì?

---

## Đọc thêm

- [gRPC — Core concepts](https://grpc.io/docs/what-is-grpc/core-concepts/)
- [Protocol Buffers — Language guide (proto3)](https://protobuf.dev/programming-guides/proto3/)
- [Envoy xDS protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) — cách Envoy phân phối config (tham khảo, phức tạp hơn edgeway)
- [AWS Builders' Library — Static stability](https://aws.amazon.com/builders-library/static-stability-using-availability-zones/)

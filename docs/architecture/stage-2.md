# Stage 2 — Tách Control Plane / Data Plane · `v0.6 → v0.9`

> Trạng thái: kiến trúc đích, chưa bắt đầu.
> Học trước: [docs/learning/stage-2.md](../learning/stage-2.md)

## Mục tiêu

Quản lý tập trung nhiều node gateway qua API, dashboard và CLI. Lần tách service đầu tiên.

## Yêu cầu dẫn tới stage này

- Chạy từ 2 node gateway trở lên: config phải giống nhau trên mọi node và đổi đồng loạt.
- Cần biết node nào đang sống, đang chạy config phiên bản nào.
- Người quản trị cần UI và API thay vì SSH vào sửa file.

## Sơ đồ

```
                     ┌──────────────── edgeway-controlplane ────────────────┐
Admin ─► Dashboard ─►│ REST API (/v1/...) ─► Service layer ─► Postgres        │
CLI (edgewayctl) ───►│                          │                             │
                     │                          ├─► Revision store            │
                     │                          ├─► Cert manager (ACME)       │
                     │                          └─► Distributor (gRPC)        │
                     └──────────────────────────────────┬─────────────────────┘
                                 stream ConfigSnapshot   │   ▲ Report(node, revision, health)
                     ┌───────────────────────────────────▼───┴──────────────────┐
Client ─────────────►│ edgeway node 1 · node 2 · node N                          │──► Backends
                     │ ControlPlaneProvider ─► runtime (như Stage 1)             │
                     │ cache/last-known-good.json  (dùng khi CP không liên lạc được) │
                     └──────────────────────────────────────────────────────────┘
```

## Thay đổi so với Stage 1

| | |
|---|---|
| Thêm binary | `cmd/edgeway-controlplane`, `cmd/edgewayctl` |
| Thêm thư mục | `internal/controlplane/{api,store,revision,distributor,acme,auth}`, `api/proto/`, `web/dashboard/`, `migrations/` |
| Thêm vào gateway | `provider/controlplane` + cache đĩa — **lõi router/middleware/service giữ nguyên** |
| `configmodel` | dùng chung cho CP (validate trước khi lưu) và DP; thêm `Revision` |
| Giữ nguyên | File/Docker provider vẫn chạy được độc lập (chế độ không CP) |

## Mô hình dữ liệu (Postgres, phác thảo)

| Bảng | Nội dung |
|---|---|
| `routers`, `services`, `servers`, `middlewares` | Config đang soạn (bản nháp có thể sửa) |
| `revisions` | Mỗi lần "publish" = 1 snapshot bất biến của toàn bộ `Dynamic` (JSON) + số tăng dần |
| `nodes` | Node gateway đã đăng ký: id, version, `last_seen_at`, `applied_revision` |
| `users`, `api_tokens` | Quản trị viên và token cho CLI/CI |
| `certificates` | Cert + private key **mã hoá khi lưu**, ngày hết hạn, domain |
| `audit_log` | Ai đổi gì, lúc nào |

Quy ước: timestamp `TIMESTAMPTZ`, actor là email/ID người dùng; migration đánh số, chỉ tiến không lùi.

## Giao tiếp

**REST API** (dashboard, CLI) — JSON, mô tả bằng OpenAPI:
`/v1/routers` · `/v1/services` · `/v1/middlewares` · `/v1/revisions` (publish, rollback) · `/v1/nodes` · `/v1/certificates`

**gRPC CP ↔ DP** (`api/proto/edgeway/v1/distribution.proto`, phác thảo):

```proto
service Distribution {
  // DP mở stream, CP đẩy snapshot mỗi khi có revision mới.
  rpc Watch(WatchRequest) returns (stream ConfigSnapshot);
  // DP báo định kỳ: còn sống, đang chạy revision nào, server nào down.
  rpc Report(NodeReport) returns (ReportAck);
}

message WatchRequest   { string node_id = 1; int64 last_revision = 2; }
message ConfigSnapshot { int64 revision = 1; bytes dynamic_json = 2; repeated Certificate certs = 3; }
```

Gửi **snapshot đầy đủ** thay vì delta: đơn giản, tự sửa sai, đủ nhanh ở quy mô này.

## Chi tiết từng version

### v0.6 — CP API + Postgres + polling
- CRUD router/service/server, validate bằng `configmodel` trước khi lưu.
- Publish → tạo revision bất biến; rollback = publish lại revision cũ.
- Gateway thêm `ControlPlaneProvider` kiểu polling: `GET /v1/config` với `If-None-Match: <revision>` → `304` nếu không đổi.
- Chạy được 2 node cùng nhận một config.

### v0.7 — gRPC streaming + cache đĩa
- Thay polling bằng `Watch` stream; reconnect với backoff có jitter.
- Gateway ghi `last-known-good` ra đĩa; khởi động khi CP chết → dùng cache.
- `Report` heartbeat → bảng `nodes`; dashboard thấy node lệch revision.

### v0.8 — Dashboard + CLI + auth
- Dashboard React + TypeScript (Vite): danh sách router/service, publish/rollback, trạng thái node.
- `edgewayctl`: `get`, `apply -f`, `diff`, `publish`, `rollback` — dùng được trong CI.
- Đăng nhập admin (session cookie, CSRF), API token có phạm vi; node gateway xác thực bằng bootstrap token.

### v0.9 — Cert manager tập trung
- CP xin/gia hạn cert qua ACME; challenge HTTP-01 do gateway trả lời (CP đẩy token challenge xuống) hoặc DNS-01.
- Cert phân phối kèm snapshot; gia hạn trước hạn 30 ngày; cảnh báo khi sắp hết hạn.
- Private key mã hoá khi lưu (khoá chính lấy từ biến môi trường / secret manager).

## Exit criteria

- Tắt CP: mọi node gateway vẫn phục vụ bình thường; khởi động lại gateway khi CP chết vẫn chạy bằng cache.
- Publish trên dashboard → mọi node áp dụng < 1 giây; dashboard hiện node nào lệch.
- Rollback về revision bất kỳ bằng 1 lệnh.

## Rủi ro / điểm khó

- Node nhận config nửa vời → luôn áp dụng nguyên snapshot, validate lại ở DP trước khi swap.
- Thứ tự sự kiện khi reconnect → so sánh số revision, bỏ snapshot cũ hơn bản đang chạy.
- Phân phối private key qua mạng → bắt buộc TLS ngay từ v0.7, mTLS ở Stage 3.
- CP trở thành mục tiêu tấn công giá trị nhất → auth từ ngày đầu, audit log mọi thay đổi.

## ADR cần viết khi mở stage

- Tách CP/DP và giao thức phân phối (snapshot qua gRPC stream).
- Lưu trữ: Postgres, thư viện truy cập (vd `pgx` + `sqlc`), công cụ migration.
- Mô hình xác thực: admin, API token, node.
- Quản lý và mã hoá cert/secret.

## Lý do sang Stage 3

Muốn phục vụ người dùng thật: rate limit phải đúng khi có nhiều node, hệ thống phải chịu tải, có HA, và kênh CP↔DP phải được bảo vệ chặt.

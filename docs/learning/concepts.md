# Khái niệm nền — dành cho người đi từ SaaS sang infra

## 1. Mảng này là gì

Khi làm SaaS, ta đứng **trên** một lớp hạ tầng: deploy là có HTTPS, có domain, có log, có load balancer —
người khác (DevOps, cloud, Vercel, Cloudflare) lo. edgeway là việc tự xây **lớp hạ tầng đó**.

| Trụ cột | Trả lời câu hỏi | Ví dụ |
|---|---|---|
| Networking | Request đi từ người dùng tới app thế nào? | DNS, load balancer, reverse proxy |
| Security | Chặn request xấu, bảo vệ dữ liệu trên đường đi? | TLS, rate limit, WAF, auth |
| Observability | Hệ thống đang ra sao, chậm/lỗi ở đâu? | OpenTelemetry, SigNoz, Grafana |

## 2. Hành trình một request — khung để đặt mọi khái niệm vào

```
Người dùng gõ shop.com
  ① DNS        "shop.com ở IP nào?"                    ← danh bạ điện thoại
  ② TCP        mở kết nối tới IP đó                     ← bấm gọi, bên kia nhấc máy
  ③ TLS        bắt tay mã hoá → ổ khoá HTTPS            ← SECURITY
  ④ Gateway    nhận request, xem đi đâu, chặn kẻ xấu    ← edgeway (Data Plane)
  ⑤ Service    app xử lý (phần SaaS quen thuộc)
  ⑥ Service    gọi service khác (order → inventory)
  ⑦ Database
  └─ Suốt hành trình: mỗi chặng ghi lại "mất bao lâu, lỗi gì" → OpenTelemetry → SigNoz
```

## 3. Đối chiếu với thứ đã biết

| Trong SaaS | Bên hạ tầng |
|---|---|
| Route `/api/orders` trong Express/Gin | **Router** — nhưng điều hướng sang *service khác*, không phải *hàm* |
| Middleware check token, CORS | **Middleware** ở gateway — chạy *trước mọi* service |
| `console.log` | **Logs** có cấu trúc (JSON), gắn `trace_id` |
| "API chậm quá" rồi mò từng chỗ | **Trace** — thấy ngay chậm ở bước nào |
| Biểu đồ Google Analytics | **Metrics** — số request, độ trễ, tỉ lệ lỗi theo thời gian |
| Gọi API bằng axios/fetch | **Service gọi service** — một cạnh trên service map |
| Gắn domain trên Vercel là có HTTPS | **DNS + TLS + ACME** — tự làm |
| Cài Sentry SDK vào app | **OpenTelemetry SDK** — nhưng không gắn với vendor nào |

## 4. Gateway (reverse proxy)

```
Client ─► EntryPoint ─► Router ─► Middlewares ─► Service (load balancer) ─► Backend
```

- **Reverse proxy**: đứng trước các service, client chỉ nói chuyện với nó. Lý do công ty nào cũng có:
  một cửa vào duy nhất cho TLS, auth, rate limit, log; ẩn cấu trúc bên trong; chia tải; đổi backend không ảnh hưởng client.
- **EntryPoint**: địa chỉ lắng nghe (`:80`, `:443`).
- **Router**: rule (`Host`, `PathPrefix`) quyết định request thuộc service nào.
- **Middleware**: chặn hoặc biến đổi request (rate limit, auth, header).
- **Service / Load balancer**: chọn một backend còn sống trong nhóm.
- **Provider**: nguồn cấp config (file, Docker, sau này Control Plane).
- **Static vs dynamic config**: static đọc lúc khởi động (mở port nào); dynamic đổi lúc chạy (route nào).
- **Hot reload**: đổi dynamic config mà không restart, không rớt request đang chạy.

## 5. Control Plane vs Data Plane

| | Data Plane | Control Plane |
|---|---|---|
| Ví von | Bảo vệ đứng cửa | Phòng quản lý ra quy định |
| Xử lý | Từng request | Config |
| Giống SaaS? | Không | **Giống hệt**: API + DB + dashboard |

Chi tiết: [../architecture/overview.md](../architecture/overview.md).

## 6. Observability

- **Trace**: toàn bộ hành trình của *một* request qua các service.
- **Span**: một chặng trong trace ("order xử lý 300ms, có gọi inventory").
- **`trace_id`**: mã chung của mọi span trong một trace — như mã vận đơn.
- **`traceparent`**: header W3C mang `trace_id` khi service này gọi service kia. **Thiếu nó là service map bị đứt.**
- **Metric**: con số theo thời gian (req/s, p99 latency, % lỗi). Bộ ba **RED** = Rate, Errors, Duration.
- **Log**: sự kiện rời rạc; gắn `trace_id` để nhảy từ log sang trace.
- **Service map**: SigNoz gộp hàng triệu span → biết ai gọi ai, bao nhiêu, nhanh/chậm → thấy chỗ nghẽn.
- **OTel Collector**: trạm trung chuyển — nhận, lọc, lấy mẫu, gửi tới một hoặc nhiều backend.

**Ai lo phần nào:**

| | Ai lo |
|---|---|
| Sinh traces/metrics/logs | Code của mình, qua OTel SDK |
| Trung chuyển, lọc | OTel Collector (config) |
| Lưu, xem, vẽ, alert | SigNoz hoặc Grafana (chọn + cấu hình, đổi được bất cứ lúc nào) |

**SigNoz vs ClickHouse vs Grafana:** SigNoz là *ứng dụng* chạy trên *database* ClickHouse; đối thủ thật của SigNoz
là bộ Grafana + Prometheus + Tempo + Loki. Lý do chọn: [ADR 0003](../adr/0003-opentelemetry-signoz.md).

Lưu ý: gateway chỉ là **một node** trên service map. Muốn thấy cả mạng lưới, **mỗi service** phải gắn OTel
(→ `examples/shop/` ở v0.4).

## 7. Security — proxy chính là bề mặt tấn công

| Rủi ro | Phòng thủ |
|---|---|
| Slowloris / client chậm giữ kết nối | Timeouts (`readHeader`, `read`, `idle`), giới hạn header/body |
| Request smuggling (`Content-Length` vs `Transfer-Encoding` lệch) | Parse chặt, dùng HTTP server chuẩn |
| Giả `X-Forwarded-For` để lách rate limit theo IP | Chỉ tin header từ trusted proxies |
| Lộ dashboard / admin API | Mặc định tắt hoặc bắt auth; không mở ra ngoài |
| Lộ private key của cert | Lưu và phân quyền cẩn thận |
| SSRF qua rule routing | Không cho route tuỳ ý vào mạng nội bộ |
| Chiếm Control Plane = điều khiển toàn bộ traffic | Auth + mTLS cho CP API và kênh CP↔DP |

HTTPS bảo vệ dữ liệu **trên đường đi** (không bị đọc/sửa) và xác thực server; **không** bảo vệ khỏi
lỗi logic, client độc hại, hay backend bị hack.

## 8. Tự kiểm tra: đã hiểu chưa?

Tự giải thích được cho một người làm SaaS khác:

- [ ] Request đi từ trình duyệt tới database qua những chặng nào, mỗi chặng hỏng kiểu gì.
- [ ] Vì sao cần reverse proxy thay vì cho client gọi thẳng vào app.
- [ ] HTTPS bảo vệ gì và **không** bảo vệ gì.
- [ ] Một request chậm → mở SigNoz tìm ra thủ phạm trong 5 phút.
- [ ] Service map vẽ từ đâu; vì sao thiếu `traceparent` thì map bị đứt.

## Đọc thêm

- MDN — [An overview of HTTP](https://developer.mozilla.org/en-US/docs/Web/HTTP/Overview)
- [Cloudflare Learning Center](https://www.cloudflare.com/learning/) — các bài DNS, TLS, reverse proxy
- *High Performance Browser Networking* (miễn phí: <https://hpbn.co>) — chương TCP, TLS
- OpenTelemetry — [Observability primer](https://opentelemetry.io/docs/concepts/observability-primer/)

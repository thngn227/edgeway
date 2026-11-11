# Bài tập Stage 0 — hiểu hành trình request bằng terminal

Mỗi bài 15–30 phút, chỉ cần terminal + Docker. Làm theo thứ tự; chạy lệnh, đọc output,
trả lời câu hỏi cuối bài. Kẹt chỗ nào thì dán output cho Claude giải thích từng dòng.
Khái niệm liên quan: [concepts.md](concepts.md).

---

## Bài 1 — DNS: tên miền → IP

```bash
dig google.com
dig google.com +short
dig +trace google.com          # đi từ root server xuống
dig github.com CNAME +short
```

**Quan sát:** mục `ANSWER SECTION`, cột TTL, loại bản ghi (A, AAAA, CNAME).

**Tự trả lời:**
- TTL nghĩa là gì, ảnh hưởng thế nào khi đổi IP server?
- Vì sao một domain có thể trả về nhiều IP?
- `+trace` cho thấy DNS được hỏi qua những tầng nào?

---

## Bài 2 — Một request HTTP từ A đến Z

```bash
curl -v https://example.com -o /dev/null
curl -s -o /dev/null -w "dns=%{time_namelookup} tcp=%{time_connect} tls=%{time_appconnect} ttfb=%{time_starttransfer} total=%{time_total}\n" https://example.com
```

**Quan sát** (trong output `-v`):
- Dòng `Trying <IP>...` → `Connected` (TCP)
- Các dòng `TLS handshake` / `SSL connection using ...` (TLS)
- Dòng `>` là request gửi đi, `<` là response nhận về (HTTP)

**Tự trả lời:**
- Map từng nhóm dòng vào chặng ①②③ trong [hành trình request](concepts.md#2-hành-trình-một-request--khung-để-đặt-mọi-khái-niệm-vào).
- Trong lệnh thứ hai, chặng nào tốn thời gian nhất? Chạy lại lần 2 có nhanh hơn không, vì sao?
- Header nào do client gửi, header nào do server trả?

---

## Bài 3 — TLS và chứng chỉ

```bash
openssl s_client -connect google.com:443 -servername google.com </dev/null 2>/dev/null | head -30
openssl s_client -connect google.com:443 -servername google.com </dev/null 2>/dev/null \
  | openssl x509 -noout -subject -issuer -dates
```

**Quan sát:** `Certificate chain` (cert server → cert trung gian → CA), `subject`, `issuer`, ngày hết hạn.

**Tự trả lời:**
- Vì sao cần chuỗi chứng chỉ thay vì một cert?
- `-servername` (SNI) dùng để làm gì? Gợi ý: một IP có thể chứa nhiều domain.
- Cert hết hạn thì trình duyệt phản ứng thế nào? Ai gia hạn cert khi dùng Let's Encrypt?

---

## Bài 4 — Dựng một reverse proxy

Cần Docker đang chạy.

```bash
cd ~/Desktop/edgeway
make up                                   # whoami-1 :9001, whoami-2 :9002

curl localhost:9001                       # gọi thẳng backend

docker run --rm -p 8080:8080 caddy:2.10.0 \
  caddy reverse-proxy --from :8080 --to host.docker.internal:9001

# terminal khác:
curl localhost:8080                       # gọi qua proxy
```

**Quan sát:** so sánh output hai lần `curl` — đặc biệt `RemoteAddr`, `X-Forwarded-For`, `X-Forwarded-Proto`, `X-Forwarded-Host`.

**Tự trả lời:**
- Backend thấy IP của ai khi đi qua proxy? Làm sao nó biết IP thật của client?
- Nếu client tự gửi `X-Forwarded-For: 1.2.3.4` thì sao? (thử: `curl -H "X-Forwarded-For: 1.2.3.4" localhost:8080`)
- Đây chính là việc edgeway sẽ làm ở v0.1 — liệt kê những gì proxy phải tự làm.

Dọn dẹp: `Ctrl+C` để tắt Caddy, `make down`.

---

## Bài 5 — Cài SigNoz, làm quen giao diện

Theo hướng dẫn chính thức: <https://signoz.io/docs/install/docker/>. Cấp cho Docker Desktop đủ RAM (ClickHouse khá ngốn).

**Quan sát:** mở UI, đi qua các mục Services, Traces, Logs, Dashboards, Service Map (lúc này còn trống).

**Tự trả lời:**
- SigNoz đang chạy bao nhiêu container? Container nào là ClickHouse, container nào là OTel Collector?
- Collector nghe ở port nào để nhận OTLP? (sẽ dùng ở v0.4)

---

Xong 5 bài → tick mục "Học trước" trong [ROADMAP.md](../../ROADMAP.md) và tự kiểm tra
bằng [checklist "đã hiểu chưa"](concepts.md#8-tự-kiểm-tra-đã-hiểu-chưa).

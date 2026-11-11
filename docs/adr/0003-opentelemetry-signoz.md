# 0003 — OpenTelemetry làm chuẩn telemetry, SigNoz làm backend mặc định

- Trạng thái: Accepted
- Ngày: 2026-11-11
- Stage: 0 (áp dụng từ v0.4)

## Bối cảnh

Cần thấy được hệ thống: request chậm ở đâu, service nào gọi service nào (service map), lỗi ở đâu.
Không muốn code bị khoá vào một công cụ quan sát cụ thể.

## Quyết định

- edgeway **chỉ phát OpenTelemetry (OTLP)**: traces, metrics, logs. Không import SDK của vendor nào.
- Dữ liệu đi qua **OTel Collector**; backend mặc định để dev là **SigNoz self-host**.
- Mọi request mang `traceparent` (W3C) xuống backend để trace nối liền các service.

## Phương án đã cân nhắc

| Phương án | Ưu | Nhược |
|---|---|---|
| SigNoz | Một app cho trace + metric + log; service map có sẵn; OTel-native | ClickHouse tốn RAM; cộng đồng nhỏ hơn Grafana |
| Grafana + Prometheus + Tempo + Loki | Chuẩn ngành; kỹ năng phổ biến; từng mảnh nhẹ | 4–5 thành phần; service graph phải cấu hình thêm |
| Chỉ ClickHouse | Toàn quyền | Phải tự xây ingest, UI, service map |

## Hệ quả

- Đổi hoặc thêm backend (vd Grafana) chỉ cần sửa config Collector, không sửa code.
- Đặt tên `service.name` và truyền context đúng là trách nhiệm của code; chất lượng biểu đồ phụ thuộc vào đó.

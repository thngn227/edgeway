# CLAUDE.md — edgeway

> Claude Code tự đọc file này mỗi session trong repo. Luật global ở `~/.claude/CLAUDE.md` vẫn áp dụng.

## Dự án

**edgeway**: edge gateway / reverse proxy bằng Go, lấy cảm hứng từ Traefik, telemetry chuẩn OpenTelemetry → SigNoz.
Dự án **cá nhân để học** networking, security, observability. Độc lập hoàn toàn với mọi công việc khác.

## Người dùng

- Nền tảng: làm SaaS / web app, **mới tiếp cận** mảng infra (networking, TLS, proxy, observability).
- Giải thích bằng tiếng Việt, ví von với thứ quen thuộc trong web app (route, middleware, Sentry, axios...).
- Khái niệm mới: giải thích **vì sao** trước **làm thế nào**; có sơ đồ ASCII khi giúp hiểu luồng.
- Học xen kẽ: hiểu đủ để làm bước tiếp theo, không bắt học hết lý thuyết trước.
- Tài liệu học có sẵn: `docs/learning/concepts.md` (khái niệm, đối chiếu SaaS ↔ infra) và
  `docs/learning/stage-N.md` (N = 0..5: cần hiểu gì, bài thực hành, tự kiểm tra cho từng version).
  Trước khi bắt đầu một version, nhắc user phần học tương ứng. Giải thích nhất quán với các file này;
  khi user dán output bài tập thì giải thích từng dòng. Khái niệm mới quan trọng → đề xuất bổ sung vào `concepts.md`.
- Mục tiêu lớn của user: **service map** thấy cả hệ thống như mạng lưới (node = service, thấy traffic, chỗ nghẽn)
  → cần `examples/shop/` gắn OTel ở v0.4, không chỉ gateway.
- Đưa ra **một khuyến nghị** kèm lý do, không liệt kê lan man.

## Đang ở đâu

- **`ROADMAP.md` là nguồn chân lý** về stage/version hiện tại — đọc trước khi bắt đầu việc gì.
- Kiến trúc: `docs/architecture/overview.md` + `stage-1.md` … `stage-5.md`. Lý do quyết định: `docs/adr/`.
- Stage chưa bắt đầu là **kiến trúc đích**: khi mở stage, xem lại cùng user, viết các ADR liệt kê trong mục
  "ADR cần viết", tạo milestones/issues — theo checklist "Chuyển stage" trong `CONTRIBUTING.md`.
- Quy trình branch / commit / version / tag / chuyển stage: `CONTRIBUTING.md`.

## Kiến trúc cốt lõi (tóm tắt)

- Data Plane (`cmd/edgeway`): EntryPoint → Router → Middlewares → Service (LB) → Backend; Provider cấp dynamic config.
- Control Plane (Stage 2+): chỉ là **thêm một Provider** cho gateway, không viết lại lõi.
- `internal/configmodel` là hợp đồng trung tâm; static config (khởi động) tách khỏi dynamic config (runtime).
- Nguyên tắc: DP không phụ thuộc CP còn sống · hot path không gọi DB/mạng ngoài · DP stateless · telemetry chỉ qua OTel.

## Cách làm việc

- **Không tự commit / push / tag / tạo branch / merge.** Mặc định chỉ sửa file trong working tree rồi dừng, báo lại để review.
  Chỉ làm khi user nói đúng chữ "commit" / "push" / "tag" cho đúng đợt đó.
- **Quy ước riêng của repo này** (thay cho mọi quy ước công việc trong luật global):
  - Commit theo Conventional Commits như `CONTRIBUTING.md`; không thêm footer mã chức năng / mã task, không hỏi mã task.
  - Không dùng hệ thống quản lý task bên ngoài, không gate flow, không quy trình branch theo môi trường.
    Theo dõi tiến độ bằng `ROADMAP.md` + GitHub Milestones/Issues; branch theo trunk-based.
  - Vẫn giữ: không tự commit/push, chạy `make check` trước mọi commit, dòng `Co-Authored-By` theo hướng dẫn hệ thống.
- Làm theo từng version trong roadmap: trình kế hoạch → user chốt → code → `make check` → báo kết quả → dừng.
- Package mới chỉ tạo khi version cần đến; không tạo thư mục rỗng, không thêm dependency khi chưa cần.
- Comment thưa: chỉ giải thích WHY hoặc chỗ đặc biệt. Code Go theo idiom stdlib.
- Trước khi báo xong: chạy `make check` (fmt-check + vet + test -race + build); fail thì nói fail kèm output.
- Thay đổi kiến trúc → đề xuất ADR mới (copy `docs/adr/0000-template.md`).
- Thay đổi người dùng thấy được → cập nhật mục `[Unreleased]` trong `CHANGELOG.md`.

## Lệnh hay dùng

```bash
make check   # giống CI
make run     # chạy gateway với configs/edgeway.yaml (entry :8000, admin :8081)
make up      # backend whoami :9001, :9002
make help
```

## Lưu ý môi trường

- Go 1.26. Module path tạm là `edgeway` (đổi thành `github.com/<user>/edgeway` khi có repo GitHub).
- `golangci-lint` local là v1.64.8 (build bằng Go 1.23) → không lint được Go 1.26; chưa đưa lint vào CI.
- SigNoz cài theo hướng dẫn chính thức: https://signoz.io/docs/install/docker/

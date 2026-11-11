# Quy trình làm việc

## Branch — trunk-based

- `main` luôn build được và chạy được. Không có branch dài hạn theo stage.
- Mỗi việc (1 issue) = 1 branch ngắn, sống vài ngày, merge về `main` qua PR:
  `feat/<mô-tả>`, `fix/<mô-tả>`, `docs/<mô-tả>`, `chore/<mô-tả>`.
- Thay đổi kiến trúc lớn: **thêm cái mới bên cạnh cái cũ** (vd thêm provider mới, bật bằng config)
  thay vì đập bỏ, để `main` không bao giờ vỡ giữa chừng.
- Chỉ tạo `release/X.Y` (từ tag) khi phải vá bản cũ cho người đang dùng.

## Commit — Conventional Commits

```
<type>(<scope>): <mô tả ngắn, imperative>

[body: vì sao]
```

- type: `feat` `fix` `refactor` `perf` `test` `docs` `chore` `ci`
- scope: `config` `server` `router` `middleware` `service` `tls` `otel` `admin` `deploy` ...
- Ví dụ: `feat(router): support PathPrefix rule`

## Version & tag — SemVer

| | Khi nào |
|---|---|
| `v0.x.y` | Trước 1.0: config/API còn được thay đổi phá vỡ |
| minor `0.3 → 0.4` | Xong một version trong roadmap |
| patch `0.3.0 → 0.3.1` | Sửa lỗi |
| major `1.x → 2.0` | Thay đổi phá vỡ hợp đồng config/plugin (sau 1.0) |

Phát hành một version:
1. Mọi issue của milestone đã đóng, `make check` xanh.
2. Chuyển mục `[Unreleased]` trong `CHANGELOG.md` thành `[0.X.0] - YYYY-MM-DD`.
3. Cập nhật dòng "Version mới nhất" trong `ROADMAP.md`, tick checklist.
4. `git tag -a v0.X.0 -m "v0.X.0"` → push tag → tạo GitHub Release từ tag.

## Theo dõi tiến độ (GitHub)

- **Milestone** = 1 version (`v0.1`, `v0.2`...).
- **Issue** dùng template *Task*, gắn label:
  - `stage:0` … `stage:5`
  - `area:config` `area:server` `area:router` `area:middleware` `area:service` `area:tls` `area:otel` `area:docs` `area:ci`
  - `type:feat` `type:fix` `type:chore` `type:docs`
- **Project board** (Todo / In progress / Done) gom mọi issue, lọc theo stage.
- `ROADMAP.md` là nguồn chân lý: mở ra là biết đang ở stage nào, version nào.

## Chuyển stage

Chuyển stage là thay đổi **tài liệu**, không phải thao tác git đặc biệt. Code tiếp tục trên `main`.

1. **Đóng stage cũ:** kiểm exit criteria trong `ROADMAP.md` → tag version cuối → giữ nguyên
   `docs/architecture/stage-N.md` làm ảnh chụp lịch sử.
2. **Mở stage mới:**
   - Viết ADR: yêu cầu mới nào buộc kiến trúc phải đổi, đã cân nhắc phương án nào.
   - Viết `docs/architecture/stage-(N+1).md`: kiến trúc đích + sơ đồ.
   - Cập nhật `ROADMAP.md`: dòng "Hiện tại", bảng trạng thái.
   - Tạo milestones + issues cho các version của stage.
3. Làm việc như bình thường.

## ADR — Architecture Decision Record

Mỗi quyết định kiến trúc quan trọng = 1 file `docs/adr/NNNN-<slug>.md`, copy từ
[`0000-template.md`](docs/adr/0000-template.md). ADR không sửa nội dung sau khi Accepted;
quyết định đổi thì viết ADR mới và đánh dấu ADR cũ là `Superseded by NNNN`.

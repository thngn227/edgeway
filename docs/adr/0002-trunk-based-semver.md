# 0002 — Trunk-based development + SemVer theo stage

- Trạng thái: Accepted
- Ngày: 2026-11-11
- Stage: 0

## Bối cảnh

Dự án phát triển qua nhiều stage, kiến trúc mỗi stage khác nhau. Cần biết đang ở đâu,
quay lại được trạng thái cũ, và không bị conflict lớn khi chuyển stage.

## Quyết định

- **Trunk-based:** mọi việc merge về `main` qua branch ngắn; không có branch dài hạn theo stage.
- **SemVer:** mỗi version trong roadmap là một minor (`v0.1.0`, `v0.2.0`...), đánh dấu bằng git tag.
- **Stage** là khái niệm tài liệu: `ROADMAP.md` + `docs/architecture/stage-N.md` + ADR.
- Theo dõi bằng GitHub Milestones (= version), Issues (label `stage:*`, `area:*`), Project board.

## Phương án đã cân nhắc

| Phương án | Ưu | Nhược |
|---|---|---|
| Trunk-based + tag | `main` luôn mới nhất; không merge lớn | Thay đổi lớn phải làm kiểu "thêm bên cạnh" |
| Branch dài hạn mỗi stage | Tách biệt rõ | Lệch xa `main`, conflict lớn, sửa bug 2 nơi |
| Git-flow (develop/release) | Quen thuộc | Nặng nề cho dự án 1 người |

## Hệ quả

- Xem lại kiến trúc cũ: `git checkout v0.5.0`.
- Vá bản cũ khi đã có người dùng: tạo `release/X.Y` từ tag.

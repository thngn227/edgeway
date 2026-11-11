# Changelog

Định dạng theo [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
version theo [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Khung dự án Stage 0: cấu trúc `cmd/` + `internal/`, Makefile, Dockerfile, CI.
- `configmodel`: static + dynamic config, load YAML từ chối key lạ, validate tham chiếu.
- Entry point HTTP có timeouts chống slow client và graceful shutdown.
- Admin API read-only: `GET /ping`, `GET /api/version`.
- Tài liệu: ROADMAP, CONTRIBUTING, kiến trúc tổng quan + Stage 1, ADR 0001–0003.
- Tài liệu học: khái niệm nền (`docs/learning/concepts.md`) và 5 bài tập Stage 0.
- Kiến trúc đích và tài liệu học cho Stage 1 → 5 (`docs/architecture/stage-N.md`, `docs/learning/stage-N.md`).

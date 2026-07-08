# Progress Deployment — matchnbuild-be

> Catatan lanjutan antar-sesi. Di conversation baru, minta Claude:
> **"Baca PROGRESS.md dan DEPLOYMENT.md untuk melanjutkan deployment."**

Terakhir diperbarui: 2026-07-08

---

## Tujuan
Deploy backend Go (Gin) + PostgreSQL + WebSocket chat ke **VPS** dengan
HTTPS otomatis (Caddy). Domain: **api.revolusi-edukasi.com**.

Frontend deploy terpisah (Vercel). Backend TIDAK bisa di Vercel karena
WebSocket + Hub in-memory butuh proses yang hidup terus (serverless tak cocok).

---

## Status: Tahap A SELESAI ✅ | Tahap B (deploy VPS) BELUM

### Tahap A — persiapan produksi (SELESAI, sudah dites jalan lokal)
Stack terbukti jalan di Docker lokal: build binary → migrasi otomatis →
app serve `/api/v1/*` (200 JSON) → reverse proxy → postgres healthy.

File yang DIBUAT:
- `docker/Dockerfile.prod` — multi-stage, binary statis (CGO_ENABLED=0), non-root
- `docker-compose.prod.yml` — versi nginx (port 80, tanpa HTTPS) — untuk tes lokal
- `docker-compose.caddy.yml` — **versi Caddy (auto-HTTPS) — INI untuk VPS**
- `docker/caddy/Caddyfile` — config domain api.revolusi-edukasi.com
- `docker/nginx/prod.conf` — nginx WebSocket upgrade benar (dipakai versi prod.yml)
- `.dockerignore` — cegah .env & artefak dev masuk image
- `DEPLOYMENT.md` — panduan langkah demi langkah VPS

File yang DIPERBAIKI:
- `middlewares/cors.go` — hindari `*`+credentials; origin dari env `CORS_ALLOWED_ORIGINS`
- `config/database.go` — `godotenv.Load` tak lagi panic bila `.env` tak ada
- `database/migrations/20240613000000_add_client_message_id_to_messages.go`
  — **BUG DIPERBAIKI**: dibuat idempotent (`IF NOT EXISTS`). Sebelumnya selalu
  gagal di DB fresh karena bentrok dengan AutoMigrate yang sudah buat kolom itu.
- `.env` / `.env.example` — tambah `CORS_ALLOWED_ORIGINS`, rapikan `WS_ALLOWED_ORIGIN`

### Tahap B — deploy ke VPS (BELUM dikerjakan)
Ikuti `DEPLOYMENT.md`. Ringkas: sewa VPS → A record → SSH → install Docker →
git clone → buat `.env` produksi → `docker compose -f docker-compose.caddy.yml up -d --build`.

---

## Yang HARUS dilakukan user (blocker Tahap B)
- [ ] Sewa VPS (Ubuntu 24.04, ~2GB RAM, region Singapore/Jakarta). Belum dilakukan.
- [ ] Buat A record `api.revolusi-edukasi.com` → IP VPS. **DNS harus mengarah
      SEBELUM `up`**, kalau tidak penerbitan SSL Caddy gagal (Let's Encrypt rate limit).
- [ ] Siapkan nilai `.env` produksi (checklist Bagian 5 di DEPLOYMENT.md).
- User pakai laptop Windows → SSH via PowerShell (`ssh root@IP`). Sudah dikonfirmasi bisa.

---

## Catatan teknis penting (jangan sampai lupa)
1. **Migrasi otomatis** via `command: sh -c "/app/server --migrate:run && /app/server"`.
   Terverifikasi: `--migrate:run` meng-exit proses setelah selesai (script/command.go:133).
2. **DB_HOST** di-override compose ke `postgres`. `godotenv.Load` tak menimpa env var
   yang sudah ada, jadi `.env` boleh tetap `DB_HOST=localhost` untuk dev.
3. **WebSocket origin**: bila `APP_ENV=production` DAN `WS_ALLOWED_ORIGIN` kosong,
   SEMUA koneksi WS browser DITOLAK (handler.go:127 `return !isProduction`).
   → WAJIB isi `WS_ALLOWED_ORIGIN` = domain frontend di produksi.
4. **CGO_ENABLED=0 aman** walau import driver sqlite mattn (pakai stub; sqlite tak
   dipakai di prod, hanya test). Sudah diverifikasi compile.
5. **Payment (Midtrans) sudah aman secara kode**: verifikasi signature SHA512 +
   constant-time compare, fetch-ulang status (anti-spoof), amount check, row lock
   FOR UPDATE (idempotent), IDOR check. Lihat modules/payment/service/payment_service.go.
6. `.env` TIDAK ter-commit (`.gitignore: *.env`) — aman, tidak ada kebocoran di repo.

---

## Sisa pekerjaan opsional (non-blocker, untuk nanti)
- Rate limiting endpoint `POST /api/v1/payment/notification` (medium).
- Isi `MIDTRANS_NOTIFICATION_URL` + daftarkan di Dashboard Midtrans SETELAH domain HTTPS aktif.
- Rotasi kredensial dev di `.env` (JWT_SECRET, SMTP password, Midtrans key) untuk nilai produksi baru.
- CI/CD otomatis (sekarang deploy manual `git pull && up`).
- Laporan audit lengkap (security-audit-report.md, production-readiness-report.md)
  sesuai CLAUDE.md — BELUM dibuat, user memilih fokus deployment dulu.

---

## Cara verifikasi cepat setelah deploy
```bash
curl -i https://api.revolusi-edukasi.com/api/v1/designers   # harap 200 JSON via HTTPS
docker compose -f docker-compose.caddy.yml logs caddy | grep -i "certificate obtained"
docker compose -f docker-compose.caddy.yml ps
```

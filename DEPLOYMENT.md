# Panduan Deploy ke VPS

Stack: Go (Gin) + PostgreSQL + Nginx, via `docker-compose.prod.yml`.
Sudah dites jalan lokal. Ikuti urut dari atas.

---

## 0. Yang disiapkan dulu (sebelum sentuh VPS)

- [ ] **VPS** ~2GB RAM, Ubuntu 22.04/24.04 (DigitalOcean / Contabo / Hetzner / Vultr / Biznet)
- [ ] **Domain** diarahkan ke IP VPS (A record). Backend biasanya subdomain, mis. `api.matchnbuild.com`
- [ ] **Nilai `.env` produksi** (lihat checklist di Bagian 5)

---

## 1. Login ke VPS & pasang Docker

```bash
ssh root@IP_VPS

# update + Docker (script resmi)
apt update && apt upgrade -y
curl -fsSL https://get.docker.com | sh

# cek
docker --version
docker compose version
```

---

## 2. Ambil kode

```bash
# pasang git bila belum ada
apt install -y git

git clone https://github.com/Rizal-Nurochman/matchnbuild.git
cd matchnbuild
```

---

## 3. Buat file `.env` produksi

`.env` TIDAK ikut di git (sengaja). Buat manual di server:

```bash
cp .env.example .env
nano .env      # isi sesuai checklist Bagian 5
```

> `DB_HOST` boleh dibiarkan apa saja — compose otomatis meng-override ke `postgres`.

---

## 4. Jalankan (dengan Caddy — auto HTTPS)

> Prasyarat: A record `api.revolusi-edukasi.com` sudah mengarah ke IP VPS,
> dan port 80 + 443 terbuka (lihat Bagian 8).

```bash
docker compose -f docker-compose.caddy.yml up -d --build

# pantau: tunggu banner "Caknoo" + "migration completed successfully"
docker compose -f docker-compose.caddy.yml logs -f app

# pantau Caddy menerbitkan sertifikat SSL (cari "certificate obtained successfully")
docker compose -f docker-compose.caddy.yml logs -f caddy
```

Tes:
```bash
curl -i https://api.revolusi-edukasi.com/api/v1/designers   # harap 200 JSON, HTTPS
```

Update kode ke depannya:
```bash
git pull
docker compose -f docker-compose.caddy.yml up -d --build
```

> Caddy mengurus SSL otomatis (terbit + perpanjang) dan meneruskan WebSocket
> tanpa konfigurasi tambahan. Tidak perlu Certbot.

---

## 5. Checklist `.env` produksi

| Variabel | Isi | Catatan |
|----------|-----|---------|
| `APP_ENV` | `production` | WAJIB. Mengaktifkan validasi origin WebSocket |
| `JWT_SECRET` | string acak panjang | **Ganti!** mis. `openssl rand -base64 48` |
| `DB_USER` / `DB_PASS` / `DB_NAME` | kredensial baru | `DB_PASS` jangan pakai yang dev |
| `DB_PORT` | `5432` | |
| `CORS_ALLOWED_ORIGINS` | `https://app.domainanda.com` | domain frontend, pisah koma tanpa spasi |
| `WS_ALLOWED_ORIGIN` | `https://app.domainanda.com` | sama dgn di atas |
| `MIDTRANS_ENVIRONMENT` | `production` | |
| `MIDTRANS_SERVER_KEY` / `CLIENT_KEY` | key produksi Midtrans | |
| `MIDTRANS_NOTIFICATION_URL` | `https://api.domainanda.com/api/v1/payment/notification` | isi SETELAH domain aktif; daftarkan juga di Dashboard Midtrans |
| `MIDTRANS_FINISH_URL` | URL halaman frontend (opsional) | boleh kosong |
| `IMAGEKIT_*` | key ImageKit asli | untuk fitur upload |
| `SMTP_*` | kredensial email | untuk verifikasi/reset password |

Generate JWT secret kuat:
```bash
openssl rand -base64 48
```

---

## 6. HTTPS

Sudah otomatis lewat Caddy (Bagian 4) — tidak ada langkah tambahan.
Caddy menerbitkan & memperpanjang sertifikat Let's Encrypt sendiri, asalkan:
- A record domain mengarah ke IP VPS
- port 80 & 443 terbuka (Bagian 8)

Cek sertifikat berhasil:
```bash
docker compose -f docker-compose.caddy.yml logs caddy | grep -i "certificate obtained"
```

---

## 7. Setelah HTTPS aktif

- [ ] Isi `MIDTRANS_NOTIFICATION_URL=https://api.revolusi-edukasi.com/api/v1/payment/notification`
- [ ] Daftarkan URL yang sama di **Dashboard Midtrans → Settings → Payment Notification URL**
- [ ] `docker compose -f docker-compose.caddy.yml up -d` (restart app agar env baru terbaca)
- [ ] Tes bayar sandbox → cek log app menerima notifikasi

---

## 8. Firewall (disarankan)

```bash
ufw allow OpenSSH
ufw allow 80
ufw allow 443
ufw enable
```
Port DB (5432) TIDAK dibuka — hanya diakses internal oleh app.

---

## Troubleshooting cepat

| Gejala | Sebab / solusi |
|--------|----------------|
| `app` exit / restart terus | `docker compose ... logs app` — biasanya `.env` (JWT/DB) salah |
| Port 80 dipakai | ada web server lain (Apache/nginx host). Matikan atau ubah `NGINX_PORT` |
| WebSocket gagal connect | `APP_ENV=production` + `WS_ALLOWED_ORIGIN` harus cocok origin frontend; HTTPS aktif |
| Midtrans status tak update | `NOTIFICATION_URL` belum diisi / belum didaftarkan di dashboard / bukan HTTPS publik |

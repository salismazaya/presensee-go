## ⚠️ PERINGATAN

_Software ini GRATIS_

- ✅ **Diperbolehkan:** Penggunaan pribadi atau instansi internal.
- ❌ **Dilarang:** Menjual software ini atau menggunakannya untuk tujuan komersil tanpa izin.

---

# **Presensee 📱 (Golang Edition)**

![Presensee Intro](screenshots/intro.jpg)

> **Sistem Absensi Modern dengan Arsitektur Offline-First — cepat, ringan, efisien, dan tetap jalan meskipun tanpa internet.**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-blue?logo=react)](https://react.dev/)
[![Vite](https://img.shields.io/badge/Vite-Bundler-purple?logo=vite)](https://vitejs.dev/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-CSS%204.0-38bdf8?logo=tailwindcss)](https://tailwindcss.com/)
[![DaisyUI](https://img.shields.io/badge/DaisyUI-5.0-1ad1a5?logo=daisyui)](https://daisyui.com/)
[![GORM](https://img.shields.io/badge/GORM-DB_Agnostic-orange)](https://gorm.io/)
[![Bun](https://img.shields.io/badge/Bun-Frontend_Runtime-black?logo=bun)](https://bun.sh/)

**Presensee (Golang Edition)** adalah porting berkinerja tinggi dari sistem absensi siswa sekolah **Presensee** ke bahasa **Go**. Dirancang dengan arsitektur **Offline-First**, aplikasi ini menggabungkan kecepatan native binary Go dengan fleksibilitas database agnostic (SQLite, PostgreSQL, MySQL) dan antarmuka SPA modern.

Dibangun menggunakan stack modern: **Go & Chi & GORM** di backend, **React 19 (Vite + PWA)** di frontend dengan **Tailwind CSS 4 & DaisyUI 5**, serta **Embedded SSR Admin Panel**.

![Preview](screenshots/preview.jpg "Presensee Preview")

---

## 🌟 Fitur Utama

- 📡 **Offline-First:** Absensi tetap bisa di-input tanpa internet (in-memory SQLite snapshot & sync engine). Sinkronisasi dilakukan nanti saat online dengan deteksi konflik otomatis.
- 👥 **Role-Based Access Control (RBAC):**
  - **Sekretaris** → Input absensi harian kelas.
  - **Wali Kelas** → Monitoring, kunci absensi kelas, dan unduh rekap.
  - **Kesiswaan** → Monitoring & rekap absensi seluruh kelas.
  - **Guru Piket** → Melakukan absensi siswa menggunakan QRCode scanner (absen masuk & pulang).
  - **Admin** → Full Power & manajemen data master via Admin Panel.
- 📊 **Rekap Pintar:** Filter otomatis berdasarkan Bulan, Minggu, atau Rentang Tanggal.
- 📄 **Export & Import Excel:**
  - Import data siswa masal dari berkas `.xlsx`.
  - Export rekap matriks absensi bulanan lengkap per tanggal `.xlsx`.
- 🪪 **Cetak Kartu Siswa:** Generate dan cetak kartu pelajar otomatis siap cetak A4.
- ⚡ **Single Binary:** Backend, REST API, Admin Panel, dan asset frontend React terintegrasi dalam satu executable ringkas.

---

## 🛠️ Requirements

1. **[Go 1.23+](https://go.dev/dl/)** – Golang compiler & runtime.
2. **[Bun](https://bun.sh/)** – Runtime & package manager frontend.
3. **Database:**
   - **SQLite** (bawaan, zero-config untuk development).
   - Atau **[PostgreSQL](https://www.postgresql.org/)** / **[MySQL](https://www.mysql.com/)** untuk skala besar.
4. **[Git](https://git-scm.com/)** – Version control.

---

## 🚀 Instalasi

### 1. Clone Repository

```bash
git clone https://github.com/salismazaya/presensee-public.git
cd presensee-public
```

---

### 2. Setup Frontend (React + Vite)

Build asset frontend agar dapat di-serve langsung oleh server Go:

```bash
cd frontend

# Install dependencies frontend
bun install

# Build static assets
bun run build

cd ..
```

---

### 3. Setup Backend (Go)

```bash
# Salin konfigurasi environment
cp .env.example .env

# Build binary server
CGO_ENABLED=1 go build -o presensee cmd/server/main.go
```

---

## ⚡ Menjalankan Aplikasi

Jalankan executable binary:

```bash
./presensee
```

Atau jalankan langsung dengan `go run`:

```bash
go run cmd/server/main.go
```

Akses layanan melalui browser:
- **Aplikasi Absensi & SPA:** [http://127.0.0.1:8000](http://127.0.0.1:8000)
- **Setup Awal (Inisialisasi Admin):** [http://127.0.0.1:8000/setup](http://127.0.0.1:8000/setup)
- **Admin Panel:** [http://127.0.0.1:8000/admin](http://127.0.0.1:8000/admin)

---

## 👥 Kredit

Crafted with ❤️ by **[Salis Mazaya](https://mazaya.is-a.dev)**

---

## 🤖 Catatan Rewrite

Proyek ini adalah hasil **rewrite** dari codebase Python/Django original: **[salismazaya/presensee-public](https://github.com/salismazaya/presensee-public)**.

Seluruh proses porting dari Django ke Go — termasuk analisis codebase, perancangan arsitektur, penulisan kode backend, pembuatan admin panel SSR, porting offline-first sync engine, dan pembuatan test — **dilakukan menggunakan AI** ([Hermes Agent](https://hermes-agent.nousresearch.com/) by Nous Research).

---

# 📋 SiLapor — Backend API

> **Sistem Pengaduan & Tracking Fasilitas Kampus**

SiLapor adalah RESTful API backend untuk sistem pelaporan kerusakan fasilitas kampus. Aplikasi ini memungkinkan mahasiswa melaporkan kerusakan, petugas menangani laporan sesuai kategori tanggung jawabnya, dan admin memantau seluruh aktivitas pengaduan secara real-time.

---

## 🎯 Deskripsi Aplikasi

### Untuk Siapa?

| Role | Kemampuan |
|---|---|
| **Mahasiswa** | Membuat laporan kerusakan, memantau status laporan milik sendiri |
| **Petugas** | Melihat & memperbarui status laporan di kategori yang ditugaskan |
| **Admin** | Akses penuh: kelola user, kategori, laporan, dan pantau dashboard |

### Fitur Utama

- 🔐 **Autentikasi JWT** — Login aman dengan token berbasis peran (role-based)
- 📝 **Manajemen Laporan** — CRUD laporan dengan upload foto bukti ke Supabase Storage
- 🏷️ **Kategori Fasilitas** — Pengelompokan laporan per jenis fasilitas dengan petugas PIC
- ⏱️ **SLA & Eskalasi Otomatis** — Prioritas laporan otomatis naik ke `tinggi` jika melewati batas waktu SLA
- 📊 **Dashboard Admin** — Statistik laporan, prioritas tinggi, dan distribusi per kategori
- 🗂️ **Riwayat Status** — Log perubahan status laporan sebagai timeline audit trail
- 📚 **Swagger UI** — Dokumentasi API interaktif siap pakai

---

## 🛠️ Teknologi yang Digunakan

### Backend Stack

| Komponen | Teknologi | Versi |
|---|---|---|
| Bahasa | Go (Golang) | 1.26.1 |
| Web Framework | [Fiber v2](https://gofiber.io/) | v2.52.13 |
| ORM | [GORM](https://gorm.io/) | v1.31.1 |
| Autentikasi | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | v5.3.1 |
| Password Hashing | bcrypt (`golang.org/x/crypto`) | v0.53.0 |
| Config / Env | [godotenv](https://github.com/joho/godotenv) | v1.5.1 |
| API Docs | [Swaggo](https://github.com/swaggo/swag) + Fiber Swagger | v1.16.6 |

### Database & Storage

| Komponen | Teknologi |
|---|---|
| Database | PostgreSQL (via [Supabase](https://supabase.com/)) |
| DB Driver | GORM PostgreSQL Driver (`gorm.io/driver/postgres`) |
| File Storage | Supabase Storage (REST API) |

---

## 📁 Struktur Folder

```
be_silapor/
├── main.go                        # Entry point: inisialisasi app, DB, router, dan scheduler
│
├── config/
│   ├── database.go                # Koneksi ke Supabase PostgreSQL via GORM
│   ├── cors.go                    # Konfigurasi CORS (whitelist origin frontend)
│   └── middleware/
│       └── jwt.go                 # Middleware autentikasi JWT & otorisasi role
│
├── router/
│   └── router.go                  # Pendaftaran semua route/endpoint API
│
├── handler/
│   ├── auth_handler.go            # Register, Login, Ganti Password
│   ├── laporan_handler.go         # CRUD laporan + upload foto + update status
│   ├── kategori_handler.go        # CRUD kategori fasilitas
│   ├── dashboard_handler.go       # Statistik dashboard untuk admin
│   └── user_handler.go            # Manajemen user oleh admin
│
├── repository/
│   ├── user_repository.go         # Query database untuk tabel users
│   ├── laporan_repository.go      # Query database untuk tabel laporan
│   ├── kategori_repository.go     # Query database untuk tabel kategori_fasilitas
│   └── riwayat_repository.go      # Query database untuk tabel riwayat_status
│
├── model/
│   └── model.go                   # Definisi struct model DB & request/response API
│
├── pkg/
│   ├── scheduler/
│   │   └── escalation.go          # Goroutine scheduler eskalasi SLA (interval 30 menit)
│   └── storage/
│       └── storage.go             # Fungsi upload file ke Supabase Storage via REST API
│
├── docs/
│   ├── docs.go                    # File auto-generated Swagger (jangan diedit manual)
│   ├── swagger.json               # Spec API dalam format JSON
│   └── swagger.yaml               # Spec API dalam format YAML
│
├── .env.example                   # Contoh variabel environment yang dibutuhkan
├── .gitignore                     # File/folder yang diabaikan Git
├── go.mod                         # Definisi module dan dependensi Go
└── go.sum                         # Lock file checksum dependensi
```

---

## 🗄️ Penjelasan Database

Database menggunakan **PostgreSQL** yang di-hosting di [Supabase](https://supabase.com/). Schema dikelola otomatis oleh GORM AutoMigrate saat aplikasi pertama kali dijalankan.

### Tabel & Kolom

#### `users`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | `uint` (PK) | Primary key auto-increment |
| `nama` | `string` | Nama lengkap pengguna |
| `username` | `string` UNIQUE | Username unik untuk login |
| `password` | `string` | Password ter-hash (bcrypt) — tidak dikirim ke response |
| `role` | `string` | Peran: `mahasiswa` / `petugas` / `admin` (default: `mahasiswa`) |
| `created_at` | `timestamp` | Waktu akun dibuat (auto-fill) |

#### `kategori_fasilitas`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | `uint` (PK) | Primary key auto-increment |
| `nama_kategori` | `string` | Nama kategori (misal: Listrik, Air, Meja) |
| `petugas_id` | `uint` FK nullable | ID petugas yang bertanggung jawab (bisa null) |
| `sla_jam` | `int` | Batas waktu penanganan dalam jam (default: `48`) |

#### `laporan`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | `uint` (PK) | Primary key auto-increment |
| `pelapor_id` | `uint` FK | ID mahasiswa pelapor → `users.id` |
| `kategori_id` | `uint` FK | ID kategori fasilitas → `kategori_fasilitas.id` |
| `lokasi` | `string` | Lokasi kerusakan (misal: "Gedung A Lantai 2") |
| `deskripsi` | `string` | Penjelasan detail kerusakan |
| `foto_url` | `string` | URL foto bukti kerusakan (opsional) |
| `status` | `string` | `dilaporkan` / `ditugaskan` / `dikerjakan` / `selesai` |
| `prioritas` | `string` | `normal` (default) / `tinggi` (otomatis naik jika lewat SLA) |
| `tanggal_lapor` | `timestamp` | Waktu laporan dibuat (auto-fill) |
| `tanggal_selesai` | `timestamp` nullable | Diisi saat status berubah ke `selesai` |
| `bukti_selesai` | `string` | URL foto bukti penyelesaian oleh petugas |

#### `riwayat_status`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | `uint` (PK) | Primary key auto-increment |
| `laporan_id` | `uint` FK | ID laporan yang berubah → `laporan.id` |
| `status` | `string` | Status saat perubahan terjadi |
| `keterangan` | `string` | Penjelasan singkat perubahan |
| `waktu` | `timestamp` | Waktu perubahan (auto-fill) |

### Diagram Relasi (ERD)

```
┌─────────┐         ┌──────────────────────┐
│  users  │◄────────┤  kategori_fasilitas  │
│         │         │  (petugas_id → users)│
└────┬────┘         └──────────┬───────────┘
     │ (pelapor_id)            │ (kategori_id)
     └──────────┬──────────────┘
                │
           ┌────▼────┐
           │ laporan │
           └────┬────┘
                │ (laporan_id)
                │
        ┌───────▼────────┐
        │ riwayat_status │
        └────────────────┘
```

**Relasi:**
- `users` → `laporan` : satu user bisa membuat banyak laporan (one-to-many via `pelapor_id`)
- `kategori_fasilitas` → `laporan` : satu kategori bisa memiliki banyak laporan (one-to-many via `kategori_id`)
- `users` → `kategori_fasilitas` : satu petugas bisa ditugaskan ke satu kategori (one-to-one via `petugas_id`)
- `laporan` → `riwayat_status` : satu laporan memiliki banyak entri riwayat (one-to-many via `laporan_id`)

### Setup Database Awal

Schema tabel dibuat **otomatis** oleh GORM saat aplikasi pertama kali dijalankan. Tidak perlu menjalankan SQL migration secara manual. Pastikan:

1. Project Supabase sudah dibuat dan database PostgreSQL tersedia
2. Variabel `SUPABASE_DSN` sudah diisi dengan benar di file `.env`
3. Bucket storage `silapor` sudah dibuat di Supabase Storage

---

## 🚀 Cara Menjalankan Aplikasi

### Prasyarat

- [Go](https://go.dev/dl/) versi 1.21+
- Akun [Supabase](https://supabase.com/) dengan project aktif
- Git

### Langkah-Langkah

**1. Clone Repository**

```bash
git clone https://github.com/<username>/be_silapor.git
cd be_silapor
```

**2. Salin File Environment**

```bash
cp .env.example .env
```

**3. Isi Variabel Environment**

Buka file `.env` dan isi semua nilai yang dibutuhkan (lihat bagian [Environment Variables](#-environment-variables)).

**4. Install Dependensi**

```bash
go mod tidy
```

**5. (Opsional) Generate Ulang Dokumentasi Swagger**

> Lewati langkah ini jika folder `docs/` sudah ada dan tidak ada perubahan pada handler.

```bash
# Install swag CLI (hanya sekali)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs dari komentar di source code
swag init
```

**6. Jalankan Aplikasi**

```bash
go run main.go
```

Aplikasi berjalan di: **`http://localhost:3000`**

Swagger UI tersedia di: **`http://localhost:3000/swagger/index.html`**

### Build Untuk Production (Opsional)

```bash
# Build binary executable
go build -o silapor main.go

# Jalankan binary
./silapor          # Linux/Mac
silapor.exe        # Windows
```

---

## 📡 Daftar Endpoint API

**Base URL:** `http://localhost:3000/api`

> 🔒 Endpoint bertanda **[Auth]** membutuhkan header:
> `Authorization: Bearer <token_dari_login>`

---

### 🔑 Auth

| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| `POST` | `/api/register` | Public | Daftar akun baru (role default: mahasiswa) |
| `POST` | `/api/login` | Public | Login dan dapatkan JWT token |
| `PUT` | `/api/changepassword` | 🔒 Semua | Ganti password sendiri |

<details>
<summary><b>POST /api/register — Daftar akun baru</b></summary>

**Request Body (JSON):**
```json
{
  "nama": "Budi Santoso",
  "username": "budi123",
  "password": "secret123"
}
```

**Response 201 Created:**
```json
{
  "message": "register berhasil",
  "data": {
    "id": 1,
    "nama": "Budi Santoso",
    "username": "budi123",
    "role": "mahasiswa",
    "created_at": "2025-01-15T08:00:00Z"
  }
}
```

**Response 409 Conflict (username sudah dipakai):**
```json
{
  "message": "username sudah digunakan"
}
```
</details>

<details>
<summary><b>POST /api/login — Login dan dapatkan token</b></summary>

**Request Body (JSON):**
```json
{
  "username": "budi123",
  "password": "secret123"
}
```

**Response 200 OK:**
```json
{
  "message": "login berhasil",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "nama": "Budi Santoso",
      "username": "budi123",
      "role": "mahasiswa",
      "created_at": "2025-01-15T08:00:00Z"
    }
  }
}
```

**Response 401 Unauthorized:**
```json
{
  "message": "username atau password salah"
}
```
</details>

<details>
<summary><b>PUT /api/changepassword — Ganti password 🔒</b></summary>

**Request Body (JSON):**
```json
{
  "old_password": "secret123",
  "new_password": "newpassword456"
}
```

**Response 200 OK:**
```json
{
  "message": "password berhasil diubah"
}
```
</details>

---

### 🏷️ Kategori Fasilitas

| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/kategori` | 🔒 Semua | Lihat semua kategori fasilitas |
| `POST` | `/api/kategori` | 🔒 Admin | Tambah kategori baru |
| `PUT` | `/api/kategori/:id` | 🔒 Admin | Update kategori |
| `DELETE` | `/api/kategori/:id` | 🔒 Admin | Hapus kategori |

<details>
<summary><b>GET /api/kategori — Daftar kategori 🔒</b></summary>

**Response 200 OK:**
```json
{
  "message": "berhasil mengambil data kategori",
  "data": [
    {
      "id": 1,
      "nama_kategori": "Kelistrikan",
      "petugas_id": 3,
      "sla_jam": 24,
      "petugas": { "id": 3, "nama": "Pak Joko", "role": "petugas" }
    }
  ]
}
```
</details>

<details>
<summary><b>POST /api/kategori — Tambah kategori 🔒 Admin</b></summary>

**Request Body (JSON):**
```json
{
  "nama_kategori": "Kelistrikan",
  "petugas_id": 3,
  "sla_jam": 24
}
```

> `petugas_id` dan `sla_jam` bersifat opsional.

**Response 201 Created:**
```json
{
  "message": "kategori berhasil dibuat",
  "data": {
    "id": 1,
    "nama_kategori": "Kelistrikan",
    "petugas_id": 3,
    "sla_jam": 24
  }
}
```
</details>

---

### 📝 Laporan

| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/laporan` | 🔒 Semua | Lihat laporan (difilter otomatis per role) |
| `GET` | `/api/laporan/:id` | 🔒 Semua | Lihat detail satu laporan |
| `GET` | `/api/laporan/:id/riwayat` | 🔒 Semua | Timeline riwayat perubahan status |
| `POST` | `/api/laporan` | 🔒 Mahasiswa | Buat laporan baru (multipart/form-data) |
| `PUT` | `/api/laporan/:id/status` | 🔒 Admin, Petugas | Update status laporan |
| `DELETE` | `/api/laporan/:id` | 🔒 Admin | Hapus laporan |

> **Catatan `GET /api/laporan`:** Data yang dikembalikan otomatis difilter berdasarkan role:
> - `mahasiswa` → hanya laporan miliknya sendiri
> - `petugas` → hanya laporan dari kategori yang ditugaskan kepadanya
> - `admin` → semua laporan tanpa filter

<details>
<summary><b>POST /api/laporan — Buat laporan baru 🔒 Mahasiswa</b></summary>

**Content-Type:** `multipart/form-data`

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `kategori_id` | string | ✅ | ID kategori fasilitas |
| `lokasi` | string | ✅ | Lokasi kerusakan |
| `deskripsi` | string | ✅ | Penjelasan detail kerusakan |
| `bukti` | file | ❌ | Foto bukti kerusakan (jpg/png) |

**Response 201 Created:**
```json
{
  "message": "laporan berhasil dibuat",
  "data": {
    "id": 10,
    "pelapor_id": 1,
    "kategori_id": 2,
    "lokasi": "Gedung B Lantai 3",
    "deskripsi": "Lampu mati sejak kemarin",
    "foto_url": "https://xxx.supabase.co/storage/v1/object/public/silapor/...",
    "status": "ditugaskan",
    "prioritas": "normal",
    "tanggal_lapor": "2025-01-15T09:00:00Z",
    "pelapor": { "id": 1, "nama": "Budi Santoso" },
    "kategori": { "id": 2, "nama_kategori": "Kelistrikan" }
  }
}
```

> Status awal otomatis `ditugaskan` jika kategori sudah memiliki petugas, atau `dilaporkan` jika belum.
</details>

<details>
<summary><b>PUT /api/laporan/:id/status — Update status laporan 🔒 Admin/Petugas</b></summary>

**Untuk status selain `selesai` (Content-Type: application/json):**
```json
{
  "status": "dikerjakan"
}
```

**Untuk status `selesai` (Content-Type: multipart/form-data):**

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `status` | string | ✅ | Nilai: `selesai` |
| `bukti_selesai` | file | ✅ | Foto bukti penyelesaian (wajib saat selesai) |

**Nilai status yang valid:** `dilaporkan` / `ditugaskan` / `dikerjakan` / `selesai`

**Response 200 OK:**
```json
{
  "message": "status laporan berhasil diubah",
  "data": { ... }
}
```
</details>

<details>
<summary><b>GET /api/laporan/:id/riwayat — Timeline riwayat status 🔒</b></summary>

**Response 200 OK:**
```json
{
  "message": "berhasil mengambil riwayat status laporan",
  "data": [
    {
      "id": 1,
      "laporan_id": 10,
      "status": "dilaporkan",
      "keterangan": "Laporan baru dibuat",
      "waktu": "2025-01-15T09:00:00Z"
    },
    {
      "id": 2,
      "laporan_id": 10,
      "status": "ditugaskan",
      "keterangan": "Otomatis ditugaskan ke petugas kategori",
      "waktu": "2025-01-15T09:00:01Z"
    }
  ]
}
```
</details>

---

### 📊 Dashboard Admin

| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/dashboard/admin` | 🔒 Admin | Statistik keseluruhan sistem |

<details>
<summary><b>GET /api/dashboard/admin 🔒 Admin</b></summary>

**Response 200 OK:**
```json
{
  "message": "Berhasil mengambil statistik dashboard",
  "data": {
    "total_laporan": 42,
    "belum_selesai": 15,
    "dieskalasi": 3,
    "total_pengguna": 28,
    "priority_reports": [
      {
        "id": 5,
        "lokasi": "Gedung C Lantai 1",
        "prioritas": "tinggi",
        "status": "ditugaskan",
        "pelapor": { "nama": "Andi" },
        "kategori": { "nama_kategori": "Sanitasi" }
      }
    ],
    "category_stats": [
      { "name": "Kelistrikan", "value": 18, "fill": "#3b82f6" },
      { "name": "Sanitasi", "value": 10, "fill": "#10b981" },
      { "name": "Furniture", "value": 14, "fill": "#f59e0b" }
    ]
  }
}
```
</details>

---

### 👥 Manajemen User (Admin)

| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/users` | 🔒 Admin | Lihat semua user |
| `POST` | `/api/users` | 🔒 Admin | Buat user baru (bisa set role apa pun) |
| `DELETE` | `/api/users/:id` | 🔒 Admin | Hapus user |

<details>
<summary><b>POST /api/users — Buat user oleh admin 🔒 Admin</b></summary>

**Request Body (JSON):**
```json
{
  "nama": "Pak Joko",
  "username": "joko_petugas",
  "password": "password123",
  "role": "petugas"
}
```

> Field `role` bisa diisi: `mahasiswa`, `petugas`, atau `admin`.

**Response 201 Created:**
```json
{
  "message": "user berhasil dibuat",
  "data": {
    "id": 5,
    "nama": "Pak Joko",
    "username": "joko_petugas",
    "role": "petugas",
    "created_at": "2025-01-15T10:00:00Z"
  }
}
```
</details>

---

### Format Response Standar

Semua endpoint menggunakan format response yang seragam:

```json
{
  "message": "pesan singkat hasil operasi",
  "data": { },
  "error": "detail error jika ada"
}
```

> Field `data` dan `error` bersifat opsional — tidak dikirim jika nilainya kosong.

---

## ⚙️ Environment Variables

Salin `.env.example` menjadi `.env` lalu isi semua nilai:

```env
# ── Database ──────────────────────────────────────────────────────────────────
# Connection string ke database PostgreSQL Supabase
# Format: postgresql://[user]:[password]@[host]:[port]/[database]?sslmode=require
# Ambil dari: Supabase Dashboard → Settings → Database → Connection String
# PENTING: Gunakan port 6543 (Transaction Pooler), bukan 5432 (Direct Connection)
SUPABASE_DSN=postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres

# ── JWT ───────────────────────────────────────────────────────────────────────
# Secret key untuk menandatangani & memverifikasi JWT token
# Gunakan string acak yang panjang (minimal 32 karakter)
# Generate dengan: openssl rand -hex 32
JWT_SECRET=ganti-dengan-secret-key-yang-kuat-dan-panjang

# ── Server ────────────────────────────────────────────────────────────────────
# Port tempat aplikasi berjalan
# Catatan: Kode saat ini hardcode port 3000 di main.go
PORT=3000

# ── Supabase Storage ──────────────────────────────────────────────────────────
# URL project Supabase (tanpa trailing slash)
# Ambil dari: Supabase Dashboard → Settings → API → Project URL
SUPABASE_URL=https://[project-ref].supabase.co

# Service Role Key untuk akses Supabase Storage
# Ambil dari: Supabase Dashboard → Settings → API → service_role (bukan anon key!)
# RAHASIA: Jangan commit ke Git atau bagikan ke publik!
SUPABASE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Nama bucket Supabase Storage tempat foto laporan disimpan
# Buat bucket di: Supabase Dashboard → Storage → New Bucket
# Pastikan bucket diset sebagai PUBLIC agar URL foto dapat diakses
SUPABASE_BUCKET=silapor
```

### Cara Mendapatkan Nilai Supabase

| Variabel | Lokasi di Supabase Dashboard |
|---|---|
| `SUPABASE_DSN` | Settings → Database → Connection String → **Transaction Pooler** (port 6543) |
| `SUPABASE_URL` | Settings → API → Project URL |
| `SUPABASE_KEY` | Settings → API → `service_role` key (bukan `anon` key) |
| `SUPABASE_BUCKET` | Buat bucket baru di menu **Storage** dengan nama `silapor` |

> ⚠️ **Penting:** Pastikan bucket `silapor` diatur sebagai **Public** di Supabase Storage agar URL foto dapat diakses langsung oleh frontend.

---

## 📚 Dokumentasi API Interaktif (Swagger)

Setelah aplikasi berjalan, buka:

```
http://localhost:3000/swagger/index.html
```

Swagger UI menyediakan dokumentasi interaktif untuk semua endpoint — termasuk kemampuan untuk mengisi parameter, mengirim request, dan melihat response langsung dari browser tanpa alat tambahan.

---

## 🤝 Kontribusi

1. Fork repository ini
2. Buat branch fitur: `git checkout -b feature/nama-fitur`
3. Commit perubahan: `git commit -m 'feat: tambah fitur X'`
4. Push ke branch: `git push origin feature/nama-fitur`
5. Buat Pull Request

---

<div align="center">
  <sub>Dibuat dengan ❤️ untuk pengelolaan fasilitas kampus yang lebih baik</sub>
</div>

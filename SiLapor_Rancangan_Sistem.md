# SiLapor — Dokumen Rancangan Sistem

**SiLapor**
*Sistem Pengaduan & Tracking Fasilitas Kampus*

Dokumen Rancangan Konsep & Arsitektur Sistem
Tugas Besar Pemrograman III (Webservice)

> **Stack:** Golang Fiber · PostgreSQL (Supabase) · React

---

## Daftar Isi

1. [Latar Belakang Masalah](#1-latar-belakang-masalah)
2. [Tujuan Sistem](#2-tujuan-sistem)
3. [Aktor dan Hak Akses](#3-aktor-dan-hak-akses)
4. [Alur Sistem (Business Flow)](#4-alur-sistem-business-flow)
5. [Diagram Alur Status Laporan](#5-diagram-alur-status-laporan)
6. [Rancangan Struktur Database](#6-rancangan-struktur-database)
7. [Rancangan Endpoint API](#7-rancangan-endpoint-api)
8. [Rancangan Tampilan Aplikasi (Frontend)](#8-rancangan-tampilan-aplikasi-frontend)
9. [Fitur Nilai Tambah (Bonus)](#9-fitur-nilai-tambah-bonus)
10. [Struktur Folder Project](#10-struktur-folder-project)
11. [Kesimpulan](#11-kesimpulan)

---

## 1. Latar Belakang Masalah

Di lingkungan kampus, kerusakan fasilitas seperti AC kelas yang mati, toilet bocor, proyektor rusak, atau kursi patah sering terjadi. Selama ini, mekanisme pelaporan masih mengandalkan cara informal: mahasiswa lapor lewat WhatsApp pribadi ke staf, mengisi Google Form yang jarang ditindaklanjuti, atau sekadar memberi tahu petugas kebersihan/keamanan secara lisan.

Masalah utama dari pendekatan ini adalah **kurangnya transparansi dan akuntabilitas**:

- Laporan sering "hilang" — tidak ada catatan resmi siapa yang melapor, kapan, dan ke mana laporan itu pergi.
- Mahasiswa tidak tahu apakah laporannya sedang diproses, diabaikan, atau sudah selesai.
- Tidak ada pembagian tanggung jawab yang jelas — siapa yang harus menangani kerusakan elektrik berbeda dengan siapa yang menangani sanitasi.
- Tidak ada mekanisme yang memastikan laporan lama tidak terus didiamkan.

SiLapor dirancang untuk menjawab masalah ini dengan menyediakan sistem pelaporan terstruktur yang otomatis mengarahkan laporan ke penanggung jawab yang tepat, memberi visibilitas status secara real-time kepada pelapor, dan secara otomatis meningkatkan prioritas laporan yang melewati batas waktu wajar penanganan (SLA).

---

## 2. Tujuan Sistem

- Memberikan kanal pelaporan kerusakan fasilitas yang terstruktur dan terdokumentasi.
- Mengotomatiskan distribusi laporan ke petugas sesuai kategori kerusakan.
- Memberikan transparansi status laporan secara real-time kepada pelapor.
- Mencegah laporan terbengkalai melalui mekanisme eskalasi otomatis berbasis SLA (Service Level Agreement).
- Menyediakan data dan statistik bagi admin untuk evaluasi kinerja penanganan fasilitas kampus.

---

## 3. Aktor dan Hak Akses

| **Role** | **Deskripsi** | **Hak Akses Utama** |
|---|---|---|
| Mahasiswa | Pengguna yang membuat laporan kerusakan | Membuat laporan baru; melihat daftar & status laporan miliknya sendiri; melihat riwayat/timeline status |
| Petugas | Penanggung jawab kategori fasilitas tertentu (mis. elektrik, sanitasi) | Melihat laporan sesuai kategori yang menjadi tanggung jawabnya; mengubah status laporan (ditugaskan → dikerjakan → selesai) |
| Admin | Pengelola sistem secara keseluruhan | Mengelola kategori fasilitas & penugasan petugas; melihat seluruh laporan; melihat dashboard statistik; mengelola akun pengguna |

---

## 4. Alur Sistem (Business Flow)

### 4.1 Alur Pelaporan oleh Mahasiswa

1. Mahasiswa login ke sistem menggunakan akun terdaftar (NIM & password).
2. Mahasiswa membuka form **"Buat Laporan Baru"**, lalu mengisi:
   - Kategori fasilitas (dipilih dari dropdown, mis. Elektrik/Sanitasi/Furnitur/IT)
   - Lokasi (mis. "Ruang B201")
   - Deskripsi kerusakan
   - Foto bukti (opsional)
3. Sistem menyimpan laporan dengan status awal `dilaporkan` dan mencatat timestamp otomatis.
4. Sistem otomatis mengetahui petugas penanggung jawab berdasarkan kategori fasilitas yang dipilih (relasi kategori → petugas sudah ditentukan oleh admin sebelumnya), lalu status berubah menjadi `ditugaskan`.
5. Mahasiswa dapat memantau status laporan kapan saja melalui halaman **"Laporan Saya"**, lengkap dengan timeline riwayat perubahan status.

### 4.2 Alur Penanganan oleh Petugas

1. Petugas login dan melihat daftar laporan yang masuk sesuai kategori tanggung jawabnya.
2. Petugas memilih laporan, mengubah status menjadi `dikerjakan` saat mulai menangani.
3. Setelah selesai, petugas mengubah status menjadi `selesai` dan sistem mencatat `tanggal_selesai` secara otomatis.
4. Setiap perubahan status otomatis tersimpan ke tabel `riwayat_status`, sehingga seluruh proses penanganan dapat ditelusuri.

### 4.3 Alur Eskalasi Otomatis (Inti Nilai Tambah Sistem)

Setiap kategori fasilitas memiliki nilai SLA (mis. 48 jam) yang menandakan batas waktu wajar suatu laporan harus mulai ditangani. Mekanisme eskalasi berjalan melalui proses background (cron job sederhana memakai **goroutine** dan **ticker** di Golang) yang berjalan berkala, dengan logika berikut:

1. Sistem memeriksa seluruh laporan berstatus `dilaporkan` atau `ditugaskan`.
2. Jika selisih waktu sejak `tanggal_lapor` melebihi `sla_jam` dari kategorinya, sistem otomatis mengubah prioritas laporan menjadi `tinggi`.
3. Sistem mencatat kejadian ini ke `riwayat_status` dengan keterangan *"Eskalasi otomatis: melewati batas SLA"*.
4. Laporan berprioritas tinggi ditampilkan menonjol (badge merah) pada dashboard admin dan petugas, sehingga tidak ada laporan yang luput dari perhatian.

### 4.4 Alur Admin

- Admin mengelola data kategori fasilitas dan menentukan petugas penanggung jawab tiap kategori.
- Admin memantau seluruh laporan yang masuk dari semua kategori melalui dashboard terpusat.
- Admin melihat statistik kinerja: jumlah laporan per kategori, rata-rata waktu penyelesaian, dan jumlah laporan yang sempat dieskalasi.

---

## 5. Diagram Alur Status Laporan

Status laporan bergerak melalui alur berikut:

```
DILAPORKAN  →  DITUGASKAN  →  DIKERJAKAN  →  SELESAI
```

> **Catatan:** Jika pada status `DILAPORKAN` atau `DITUGASKAN` laporan melewati batas SLA, prioritas otomatis berubah menjadi `tinggi` tanpa mengubah alur status utama — **prioritas dan status berjalan sebagai dua atribut independen.**

---

## 6. Rancangan Struktur Database

Struktur basis data menggunakan PostgreSQL (Supabase) dengan **4 tabel utama** yang saling berelasi melalui foreign key, dikelola lewat GORM.

### 6.1 Tabel: `users`

| **Kolom** | **Tipe** | **Keterangan** |
|---|---|---|
| id | uint | Primary Key, auto increment |
| nama | varchar | Nama lengkap pengguna |
| username / nim | varchar | Unique, dipakai untuk login |
| password | varchar | Hash bcrypt |
| role | varchar | `admin` / `petugas` / `mahasiswa` |
| created_at | timestamp | Otomatis |

### 6.2 Tabel: `kategori_fasilitas`

| **Kolom** | **Tipe** | **Keterangan** |
|---|---|---|
| id | uint | Primary Key |
| nama_kategori | varchar | Mis. Elektrik, Sanitasi, Furnitur, IT |
| petugas_id | uint | Foreign Key → `users.id` (penanggung jawab kategori) |
| sla_jam | int | Batas waktu wajar penanganan dalam jam, mis. 48 |

### 6.3 Tabel: `laporan`

| **Kolom** | **Tipe** | **Keterangan** |
|---|---|---|
| id | uint | Primary Key |
| pelapor_id | uint | Foreign Key → `users.id` |
| kategori_id | uint | Foreign Key → `kategori_fasilitas.id` |
| lokasi | varchar | Mis. "Ruang B201" |
| deskripsi | text | Penjelasan kerusakan |
| foto_url | varchar | Opsional, link foto bukti |
| status | varchar | `dilaporkan` / `ditugaskan` / `dikerjakan` / `selesai` |
| prioritas | varchar | `normal` / `tinggi` (otomatis berubah via eskalasi) |
| tanggal_lapor | timestamp | Otomatis saat insert |
| tanggal_selesai | timestamp | Nullable, terisi saat status "selesai" |

### 6.4 Tabel: `riwayat_status`

| **Kolom** | **Tipe** | **Keterangan** |
|---|---|---|
| id | uint | Primary Key |
| laporan_id | uint | Foreign Key → `laporan.id` |
| status | varchar | Snapshot status pada waktu tersebut |
| keterangan | varchar | Mis. "Eskalasi otomatis karena melewati SLA" |
| waktu | timestamp | Otomatis |

### 6.5 Diagram Relasi (ERD Sederhana)

```
users (1) ──── (N) kategori_fasilitas   [sebagai penanggung jawab]
users (1) ──── (N) laporan              [sebagai pelapor]
kategori_fasilitas (1) ──── (N) laporan
laporan (1) ──── (N) riwayat_status
```

> Total terdapat **4 relasi foreign key**, melampaui syarat minimal tugas besar (minimal 1 relasi), dan setiap relasi memiliki alasan bisnis yang jelas dan dapat dijelaskan saat presentasi.

---

## 7. Rancangan Endpoint API

### 7.1 Autentikasi

| **Method** | **Endpoint** | **Keterangan** |
|---|---|---|
| POST | `/register` | Mendaftarkan user baru |
| POST | `/login` | Login, menghasilkan JWT |
| PUT | `/changepassword` | Mengubah password |

### 7.2 Kategori Fasilitas *(khusus admin)*

| **Method** | **Endpoint** | **Keterangan** |
|---|---|---|
| GET | `/api/kategori` | Daftar semua kategori fasilitas |
| POST | `/api/kategori` | Tambah kategori baru |
| PUT | `/api/kategori/:id` | Ubah kategori / petugas penanggung jawab |
| DELETE | `/api/kategori/:id` | Hapus kategori |

### 7.3 Laporan

| **Method** | **Endpoint** | **Keterangan** |
|---|---|---|
| GET | `/api/laporan` | Daftar laporan (admin/petugas: semua sesuai akses; mahasiswa: miliknya saja) |
| GET | `/api/laporan/:id` | Detail satu laporan |
| POST | `/api/laporan` | Mahasiswa membuat laporan baru |
| PUT | `/api/laporan/:id/status` | Petugas mengubah status laporan |
| DELETE | `/api/laporan/:id` | Admin menghapus laporan |
| GET | `/api/laporan/:id/riwayat` | Melihat timeline riwayat status |

> Seluruh endpoint yang membutuhkan autentikasi dilindungi **middleware JWT**, dengan otorisasi tambahan berbasis role (`admin` / `petugas` / `mahasiswa`) untuk membedakan hak akses.

---

## 8. Rancangan Tampilan Aplikasi (Frontend)

### 8.1 Halaman Login & Register

Form sederhana dengan input username/NIM dan password. Setelah login berhasil, token JWT disimpan di `localStorage` dan pengguna diarahkan ke dashboard sesuai role-nya.

### 8.2 Dashboard Mahasiswa

- Ringkasan jumlah laporan: total, sedang diproses, selesai.
- Tombol **"+ Buat Laporan Baru"** yang menonjol.
- Daftar laporan milik sendiri dalam bentuk kartu/list, masing-masing menampilkan status terkini dengan badge warna (kuning = diproses, hijau = selesai, merah = prioritas tinggi).

### 8.3 Form Buat Laporan

- Dropdown kategori fasilitas.
- Input lokasi (text).
- Textarea deskripsi kerusakan.
- Upload foto bukti (opsional, nilai tambah).
- Validasi frontend: kategori wajib dipilih, deskripsi minimal beberapa karakter, lokasi tidak boleh kosong.
- Feedback sukses/gagal menggunakan toast atau SweetAlert.

### 8.4 Halaman Detail Laporan & Timeline

Menampilkan informasi lengkap laporan beserta **timeline vertikal** riwayat status (data dari tabel `riwayat_status`), sehingga pelapor dapat melihat secara visual progres penanganan dari waktu ke waktu — termasuk catatan jika laporan sempat dieskalasi.

### 8.5 Dashboard Admin / Petugas

Tabel data utama dengan kolom: ID, Pelapor, Kategori, Lokasi, Status, Prioritas (6 kolom sesuai ketentuan tugas), dilengkapi:

- Fitur pencarian (cari berdasarkan lokasi/pelapor).
- Filter berdasarkan status dan prioritas.
- Tombol detail, ubah status, dan hapus (khusus admin).
- Badge merah menonjol untuk laporan berprioritas tinggi (hasil eskalasi otomatis).
- Loading state saat data sedang diambil dari API.

### 8.6 Halaman Statistik *(Admin)*

- Grafik jumlah laporan per kategori (bar chart).
- Grafik rata-rata waktu penyelesaian laporan.
- Jumlah laporan yang pernah dieskalasi dalam periode tertentu.
- Dibangun menggunakan **Chart.js / Recharts** sebagai nilai tambah (bonus).

### 8.7 Halaman Kelola Kategori *(Admin)*

Form CRUD sederhana untuk menambah/mengubah/menghapus kategori fasilitas beserta penentuan petugas penanggung jawab dan nilai SLA jam.

---

## 9. Fitur Nilai Tambah (Bonus)

- Upload foto bukti kerusakan saat membuat laporan.
- Timeline visual riwayat status laporan.
- Dashboard statistik dengan Chart.js / Recharts.
- **Eskalasi otomatis berbasis background job (goroutine + ticker)** — nilai tambah teknis utama yang membedakan sistem ini dari CRUD biasa.
- Badge notifikasi visual untuk laporan berprioritas tinggi.
- Rating/feedback dari pelapor setelah laporan berstatus selesai.
- Responsive layout dan dark mode.

---

## 10. Struktur Folder Project

### 10.1 Backend (Golang Fiber)

```
backend/
├── config/
│   ├── database.go
│   └── middleware/
├── docs/              # swagger
├── handler/           # controller, request/response
├── model/             # struct GORM: users, kategori, laporan, riwayat
├── repository/        # query ke database
├── router/
├── pkg/
│   └── scheduler/     # cron job eskalasi otomatis
├── .env
├── go.mod
└── main.go
```

### 10.2 Frontend (React + Vite)

```
frontend/
├── src/
│   ├── components/
│   │   ├── atoms/
│   │   ├── molecules/
│   │   ├── organisms/   # Timeline, DataTable, dll
│   │   └── layout/
│   ├── pages/           # Login, Dashboard, BuatLaporan, DetailLaporan, Statistik
│   ├── routes/
│   ├── services/        # api.js, authService.js
│   └── App.jsx
├── .env
├── package.json
└── index.html
```

---

## 11. Kesimpulan

SiLapor dirancang bukan sekadar sebagai aplikasi CRUD pelaporan, melainkan sebagai sistem yang secara aktif menjaga akuntabilitas penanganan fasilitas kampus melalui dua mekanisme kunci: **distribusi otomatis laporan ke penanggung jawab yang tepat**, dan **eskalasi otomatis berbasis SLA** yang mencegah laporan terbengkalai.

Dengan kompleksitas teknis yang proporsional untuk dikerjakan oleh 2 orang dalam satu semester, sistem ini tetap memenuhi seluruh ketentuan teknis tugas besar (Golang Fiber, GORM, PostgreSQL Supabase, JWT, Swagger, role authorization) sekaligus menghadirkan nilai tambah nyata yang relevan dengan permasalahan sehari-hari di lingkungan kampus.

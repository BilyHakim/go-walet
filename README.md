# E-Wallet Backend System

Backend sistem e-wallet menggunakan Go (Gin), PostgreSQL, GORM, JWT, dan RabbitMQ.

## Fitur

- Registrasi pengguna
- Login dengan JWT
- Top-up saldo
- Pembayaran
- Transfer antar pengguna (asinkron dengan RabbitMQ)
- Laporan transaksi
- Update profil pengguna
- Get Pengguna By Phone Number

## Teknologi

- Go dengan framework Gin
- PostgreSQL sebagai database
- GORM sebagai ORM
- JWT untuk autentikasi
- RabbitMQ untuk pemrosesan asinkron transfer antar pengguna
- Docker dan Docker Compose untuk deployment

## Arsitektur

Sistem ini memiliki beberapa komponen utama:

1. **API Layer (Gin)**: Menerima dan memproses permintaan HTTP
2. **Controllers**: Logika bisnis untuk setiap fitur
3. **Models**: Struktur data dan interaksi dengan database
4. **Middleware**: Autentikasi JWT
5. **RabbitMQ Worker**: Memproses transfer secara asinkron

## Transfer Asinkron dengan RabbitMQ

Sistem ini menggunakan RabbitMQ untuk memproses transfer antar pengguna secara asinkron. Berikut adalah alur prosesnya:

1. **Pengguna mengirim permintaan transfer**: API menerima permintaan transfer melalui endpoint `/api/transfers`
2. **Validasi dan pengurangan saldo pengirim**: Sistem memvalidasi bahwa pengirim memiliki saldo cukup, lalu mengurangi saldo pengirim
3. **Membuat catatan transaksi**: Sistem membuat catatan transaksi dengan status "PENDING"
4. **Mengirim pesan ke RabbitMQ**: Data transfer dikirim ke antrian RabbitMQ
5. **Respons cepat ke pengguna**: API mengembalikan respons sukses ke pengguna tanpa menunggu proses transfer selesai
6. **Worker memproses pesan**: Worker RabbitMQ mengambil pesan dari antrian dan memproses transfer:
   - Mencari pengguna tujuan
   - Menambahkan saldo ke akun tujuan
   - Membuat catatan transaksi penerimaan
   - Memperbarui status transaksi pengirim menjadi "SUCCESS" atau "FAILED"

Dengan pendekatan asinkron ini, pengguna mendapatkan respons cepat dan sistem dapat memproses transfer dengan lebih efisien.

## Cara Menjalankan

### Menggunakan Docker Compose

```bash
# Clone repository
git clone https://github.com/BilyHakim/go-walet.git
cd go-walet

# Jalankan dengan Docker Compose
docker-compose up
```

### Menjalankan secara lokal

```bash
# Clone repository
git clone https://github.com/BilyHakim/go-walet.git
cd go-walet

# Siapkan PostgreSQL dan RabbitMQ
# Sesuaikan file .env dengan konfigurasi lokal

# Jalankan aplikasi
go run main.go
```

## Endpoint API

### Endpoint publik

- **POST /api/register**: Mendaftarkan pengguna baru
- **POST /api/login**: Login dan mendapatkan JWT token

### Endpoint yang memerlukan JWT

- **PUT /api/update-profile**: Update profil pengguna
- **POST /api/topup**: Top-up saldo e-wallet
- **POST /api/payments**: Melakukan pembayaran
- **POST /api/transfers**: Transfer ke pengguna lain (asinkron)
- **GET /api/transactions**: Mendapatkan riwayat transaksi
- **POST /api/get-user**: Mendapatkan informasi pengguna berdasarkan nomor telepon

## Menjalankan Test

```bash
# Jalankan integration tests
go test ./tests -v
```

## Struktur Proyek

```
/go-wallet
  ├── config/             # Konfigurasi database dan RabbitMQ
  ├── controllers/        # Handler untuk endpoint API
  ├── middleware/         # Middleware JWT
  ├── models/             # Model database
  ├── routes/             # Konfigurasi routing API
  ├── tests/              # Integration tests
  ├── worker/             # Worker RabbitMQ untuk transfer asinkron
  ├── .env                # File konfigurasi
  ├── docker-compose.yml  # Konfigurasi Docker Compose
  ├── go.mod              # Deklarasi modul Go
  ├── go.sum              # Checksum dependensi
  ├── main.go             # Entry point aplikasi
  └── README.md           # Dokumentasi
```

## Example Test By Developer

Register User
<img width="1920" height="1031" alt="image" src="https://github.com/user-attachments/assets/152f7f21-d123-47f3-a3ff-8132a8bcd94a" />

Login User
<img width="1920" height="1031" alt="image" src="https://github.com/user-attachments/assets/df63b3a7-c4fa-4f41-8da7-83d0a4710e86" />

Update User
<img width="1920" height="1031" alt="image" src="https://github.com/user-attachments/assets/7f919d6e-5ed8-4883-ae48-60950b132aec" />

Top UP Balance
<img width="1920" height="1031" alt="image" src="https://github.com/user-attachments/assets/2c05fb24-2219-4788-8df4-b8019d15aff5" />

Make Payments
<img width="1920" height="1031" alt="image" src="https://github.com/user-attachments/assets/1b81b63d-cee0-4b8a-9ac4-3cd3bafb7c29" />

Transfer Money
<img width="1920" height="1031" alt="image" src="https://github.com/user-attachments/assets/97bb98aa-f8c9-48d9-bacb-1656e2cd2433" />

Transaction History
<img width="1920" height="1031" alt="image" src="https://github.com/user-attachments/assets/63b05314-011a-4e78-a35b-6dd1d49542fe" />

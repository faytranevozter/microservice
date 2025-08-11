### Microservices Todo Example (Go + MySQL + Kong + Vite/Vue)

Contoh aplikasi microservices sederhana:
- **users-service**: service Go untuk manajemen user (MySQL)
- **todo-service**: service Go untuk todo, validasi `user_id` via users-service (MySQL)
- **Kong**: API gateway (DB-less) dengan routing dan CORS
- **Frontend**: Vite + Vue, akses API lewat Kong

### Arsitektur singkat
- Frontend → `http://localhost:8000` (Kong) → route ke service:
  - `/api/users` → `users-service:8082`
  - `/api/todos` → `todo-service:8081`
- `todo-service` memanggil `users-service` di jaringan internal Docker (`http://users-service:8082`).

### Prasyarat
- Docker dan Docker Compose

### Cara jalan cepat
1) Jalankan semua service
```
docker compose up -d
```
2) Akses aplikasi
- Frontend: `http://localhost:5173`
- Kong proxy: `http://localhost:8000`

3) Hentikan
```
docker compose down
```

4) Reset database (hapus data):
```
docker compose down -v
```

### Service & Endpoint
- users-service (port internal 8082)
  - GET `/api/users` → list users
  - POST `/api/users` → buat user `{ email, name }`
  - GET `/api/users/{id}` → ambil user by id

- todo-service (port internal 8081)
  - GET `/api/todos` → list todos
    - Query: `include_user=true` untuk embed data user
  - POST `/api/todos` → buat todo `{ title, user_id? }`
    - Jika `user_id` diisi, akan divalidasi ke users-service
  - PUT `/api/todos/{id}` → update sebagian `{ title?, completed?, user_id? }`
  - DELETE `/api/todos/{id}` → hapus todo

Semua endpoint diakses melalui Kong (`http://localhost:8000`).

### Contoh cURL
```
# Tambah user
curl -X POST http://localhost:8000/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","name":"Alice"}'

# List users
curl http://localhost:8000/api/users

# Buat todo dengan user_id
curl -X POST http://localhost:8000/api/todos \
  -H 'Content-Type: application/json' \
  -d '{"title":"Buy milk","user_id":1}'

# List todos dengan embed user
curl 'http://localhost:8000/api/todos?include_user=true'

# Toggle completed todo id=1
curl -X PUT http://localhost:8000/api/todos/1 \
  -H 'Content-Type: application/json' \
  -d '{"completed":true}'
```

### Struktur direktori
```
frontend/                  # Vite + Vue app
gateway/kong.yml           # Deklarasi Kong (DB-less)
services/
  users-service/           # Go service (users)
    main.go, go.mod, go.sum, Dockerfile
  todo-service/            # Go service (todos, call users-service)
    main.go, go.mod, go.sum, Dockerfile
docker-compose.yml
```

### Environment & Konfigurasi
- Database (MySQL 8): dibuat otomatis dengan kredensial di `docker-compose.yml`
  - DB: `app_db`, user/password: `app/app`
  - Volume: `mysql_data` (tidak expose port 3306 ke host)
- Kong: `gateway/kong.yml` (routes + CORS)
- Frontend: env `VITE_API_BASE` default diset ke `http://localhost:8000`

### Development tips
- Logs
```
docker compose logs -f db
docker compose logs -f users-service
docker compose logs -f todo-service
docker compose logs -f kong
docker compose logs -f frontend
```

- Connect ke MySQL dari host: gunakan `docker exec` (karena port 3306 tidak dipublish)
```
docker exec -it microservices-db-1 mysql -uapp -papp app_db
```

- Rebuild cepat service tertentu
```
docker compose build todo-service && docker compose up -d todo-service
```

### Catatan skema DB
- `todo-service` melakukan migrasi otomatis termasuk penambahan kolom `user_id` bila belum ada.



# Spark S3 Microservice

Go-микросервис для генерации Presigned URLs для Yandex Cloud S3. Используется в качестве Gatekeeper для загрузки и скачивания медиафайлов в Spark Messenger.

## Overview

Этот микросервис не занимается передачей файлов — он только генерирует временные ссылки (Presigned URLs), которые позволяют клиенту напрямую обмениваться данными с Yandex S3.

```
┌─────────────┐     GET /upload-url     ┌──────────────────┐     Generate Presigned URL     ┌───────────────┐
│   Frontend  │ ──────────────────────► │   Go Microservice│ ─────────────────────────────► │  Yandex S3    │
│             │ ◄────────────────────── │                  │ ◄───────────────────────────── │               │
└─────────────┘   {file_id, url}        └──────────────────┘                                └───────────────┘
       │                                                                                              │
       │ PUT file directly (no proxy)                                                                │
       └─────────────────────────────────────────────────────────────────────────────────────────────►
```

## Features

- **Presigned URL Generation** — безопасные временные ссылки для загрузки и скачивания
- **Firebase Auth** — валидация Bearer токенов из Firebase
- **File Type Validation** — проверка допустимых расширений файлов
- **Ownership Verification** — проверка принадлежности файла при удалении
- **Cache-Control Headers** — автоматическое добавление заголовков для бессрочного кэширования

## API Endpoints

### Generate Upload URL
```
GET /api/v1/storage/upload-url?type=avatar&ext=jpg
Authorization: Bearer <firebase_token>

Response:
{
    "file_id": "avatar_user123_1700000000_a1b2c3d4.jpg",
    "upload_url": "https://storage.yandexcloud.net/...",
    "expires_at": "2026-05-08T14:05:00Z"
}
```

### Generate Download URL
```
GET /api/v1/storage/download-url?file_id=avatar_user123_1700000000_a1b2c3d4.jpg
Authorization: Bearer <firebase_token>

Response:
{
    "file_id": "avatar_user123_1700000000_a1b2c3d4.jpg",
    "download_url": "https://storage.yandexcloud.net/...",
    "expires_at": "2026-05-08T14:10:00Z"
}
```

### Delete File
```
DELETE /api/v1/storage/file?file_id=avatar_user123_1700000000_a1b2c3d4.jpg
Authorization: Bearer <firebase_token>

Response:
{
    "file_id": "avatar_user123_1700000000_a1b2c3d4.jpg",
    "deleted": true
}
```

### Health Check
```
GET /health

Response:
{
    "status": "healthy",
    "service": "spark-s3-microservice",
    "version": "1.0.0",
    "time": "2026-05-08T13:00:00Z"
}
```

## File ID Format

```
{type}_{uid}_{timestamp}_{short_uuid}.{ext}
Example: avatar_user123_1700000000_a1b2c3d4.jpg
```

## Project Structure

```
spark-s3-microservice/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/                   # Configuration
│   ├── domain/                   # Business entities
│   ├── handler/                  # HTTP handlers
│   ├── middleware/              # Auth, CORS, logging
│   ├── repository/              # S3 adapter
│   └── service/                 # Business logic
├── pkg/
│   └── firebase/                # Firebase token verification
├── Dockerfile
├── render.yaml
├── go.mod
└── .env.example
```

## Local Development

### Prerequisites

- Go 1.22+
- Docker (optional)

### Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and fill in values
3. Download Firebase service account JSON and place it in the project root
4. Run `go mod tidy`
5. Run `go run ./cmd/server`

### Docker

```bash
# Build image
docker build -t spark-s3-microservice .

# Run container
docker run -p 8080:8080 --env-file .env spark-s3-microservice
```

## Deployment to Render.com

1. Push the code to GitHub
2. Create a new Web Service on Render
3. Connect your GitHub repository
4. Set the following environment variables:
   - `YANDEX_ENDPOINT`
   - `YANDEX_BUCKET`
   - `YANDEX_ACCESS_KEY_ID`
   - `YANDEX_SECRET_ACCESS_KEY`
   - `YANDEX_S3_REGION`
   - `FIREBASE_CREDENTIALS_PATH`
5. Upload the Firebase service account JSON as a Secret File
6. Deploy

Alternatively, use the `render.yaml` blueprint for one-click deployment.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `YANDEX_ENDPOINT` | Yes | storage.yandexcloud.net | Yandex S3 endpoint |
| `YANDEX_BUCKET` | Yes | — | S3 bucket name |
| `YANDEX_ACCESS_KEY_ID` | Yes | — | Access key ID |
| `YANDEX_SECRET_ACCESS_KEY` | Yes | — | Secret access key |
| `YANDEX_S3_REGION` | Yes | ru-central1 | S3 region |
| `FIREBASE_CREDENTIALS_PATH` | Yes | /app/firebase-service-account.json | Path to Firebase JSON |
| `PRESIGNED_URL_EXPIRY_UPLOAD` | No | 300 | Upload URL expiry (seconds) |
| `PRESIGNED_URL_EXPIRY_DOWNLOAD` | No | 900 | Download URL expiry (seconds) |
| `LOG_LEVEL` | No | info | Log level |
| `CORS_ALLOWED_ORIGINS` | No | * | Allowed CORS origins |

## Supported File Types

| Type | Allowed Extensions |
|------|-------------------|
| avatar | jpg, jpeg, png, webp, gif |
| message | jpg, jpeg, png, webp, gif, pdf |
| sticker | png, gif, webp |

## License

MIT
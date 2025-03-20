# SongLibrary

**SongLibrary** is a RESTful service for managing a music library. It enriches song data using an external API ([you can test it by running this code](https://github.com/FIFSAK/SongLibraryExternal)) and
stores it in a PostgreSQL database. Features include filtering, pagination, verse extraction, and Swagger documentation.

---

## Features

- Retrieve a list of songs with filtering by all fields and pagination
- Get song lyrics split into paginated verses
- Add new songs via JSON request (with enrichment from an external API)
- Update and delete existing songs
- Automatic database migration on startup
- Detailed logging with debug and info levels
- Swagger-generated API documentation
- Docker support for easy deployment
- Makefile for simplified development commands

---

## Tech Stack

- **Go** (Gin, GORM)
- **PostgreSQL**
- **Docker / Docker Compose**
- **Swagger** (Swaggo)
- **Logrus** (for logging)

---

## Docker Setup

### 1. Environment Variables

Create a `.env` file in the project root and configure it like [.env.example](.env.example) and external API URL [here](https://github.com/FIFSAK/SongLibrary/blob/master/internal/handlers/song_handlers.go#L17).

```bash:

### 2. Build & Run via Makefile

```bash
# Build Docker images
make build

# Run containers (app + postgres)
make run

# Stop containers
make stop

# Clean containers, images, volumes
make clean
```

> App will be available at: `http://localhost:8080`  
> Swagger UI: `http://localhost:8080/swagger/index.html`

---

## Local Development (No Docker)

### 1. Install dependencies

```bash
go mod tidy
```

### 2. Create `.env` and set DB credentials

### 3. Run the app

```bash
go run cmd/main.go
```

### 4. Generate Swagger docs

```bash
swag init --generalInfo cmd/main.go
```

---

## API Endpoints

### `GET /songs`

Retrieve songs with filtering and pagination  
Query Parameters:

- `id` — Song ID
- `group` — Group name
- `song` — Song name
- `releaseDate` — Release date (`2006-01-02`, `2006.01.02`, or RFC3339)
- `text` — Text fragment
- `page` — Page number (default: 1)
- `limit` — Items per page (default: 10)

---

### `GET /songs/{id}/verses`

Retrieve song lyrics, split by paragraphs (double newline `\n\n` [can change here](https://github.com/FIFSAK/SongLibrary/blob/master/internal/models/song.go#L102))  
Query:

- `page` — Page number (default: 1)
- `limit` — Verses per page (default: 3)

---

### `POST /songs`

Add a song using external API enrichment  
Body:

```json
{
  "group": "Muse",
  "song": "Supermassive Black Hole"
}
```

---

### `PUT /songs/{id}`

Update a song by ID  
Body:

```json
{
  "group_name": "Muse",
  "song_name": "Starlight",
  "release_date": "2006-07-16",
  "text": "Ooh baby, don't you know I suffer...",
  "link": "https://youtube.com/example"
}
```

---

### `DELETE /songs/{id}`

Delete a song by ID

---

## Database Migration

The app uses `gorm.AutoMigrate()` to automatically create the required table.  
Schema (simplified):

```sql
CREATE TABLE IF NOT EXISTS songs
(
    id           SERIAL PRIMARY KEY,
    group_name   TEXT NOT NULL,
    song_name    TEXT NOT NULL,
    release_date DATE NOT NULL,
    text         TEXT NOT NULL,
    link         TEXT NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_song ON songs (group_name, song_name);
```

---

## Logs

- `DEBUG` — Incoming requests, parameters, external API calls
- `INFO` — Successful actions (song added/updated/deleted)
- `ERROR` — Failures in DB operations or external API responses

---

## Swagger Docs

```bash
swag init
```

Accessible at: `http://localhost:8080/swagger/index.html`

---

## Clean Up

To remove all Docker containers, images, and volumes:

```bash
make clean
```

---

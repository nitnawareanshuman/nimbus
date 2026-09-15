# Nimbus URL Shortener

Nimbus is a simple URL-shortening service with a Chrome extension. It lets users shorten links from the browser toolbar or right-click menu and quickly copy the generated URL.

## Features

- Generates secure six-character short links
- Redirects short links to their original URLs
- Stores links permanently in PostgreSQL
- Uses Redis/Valkey as an optional lookup cache
- Includes a Chrome extension with recent-link history
- Supports Docker and Render deployment

## Tech Stack

- Go and Gin
- PostgreSQL
- Redis/Valkey (optional cache)
- Docker
- JavaScript, HTML and CSS

## Run Locally

1. Clone the repository:

   ```bash
   git clone https://github.com/nitnawareanshuman/nimbus.git
   cd nimbus
   git checkout nimbus
   ```

2. Create the environment file:

   ```bash
   cp .env.example .env
   ```

3. Start the application:

   ```bash
   docker compose up --build
   ```

The API will be available at `http://localhost:8080`. Local short links also use `http://localhost:8080` by default.

## Production configuration

Set these environment variables on the API service:

- `DB_URL` — required PostgreSQL connection string (for example, your Neon connection string).
- `BASE_URL` — required for correct public short links, for example `https://nimbus-api-vqsz.onrender.com`.
- `REDIS_URL` — optional. If it is missing or the cache is temporarily unavailable, Nimbus continues using PostgreSQL.
- `PORT` — supplied automatically by Render; defaults to `8080` locally.

Nimbus ensures the `codes` table exists at startup, so a fresh production database is usable even when the Docker migration service is not run. The SQL migration files remain the canonical schema for local/managed migration workflows.

## API

Create a short URL:

```http
POST /shorten
Content-Type: application/json

{
  "url": "https://example.com"
}
```

Health check:

```http
GET /health
```

The health endpoint returns `503` if PostgreSQL is unavailable.

## Chrome Extension

Open `chrome://extensions`, enable **Developer mode**, select **Load unpacked**, and choose the `extension` folder.

The right-click action uses the background service worker. If a free Render web service is asleep and takes too long to wake, Nimbus shows an error notification; retry after a few seconds.

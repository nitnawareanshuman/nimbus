# Nimbus URL Shortener

Nimbus is a simple URL-shortening service with a Chrome extension. It lets users shorten links from the browser toolbar or right-click menu and quickly copy the generated URL.

## Features

- Generates secure six-character short links
- Redirects short links to their original URLs
- Uses Redis for faster URL lookups
- Stores links permanently in PostgreSQL
- Includes a Chrome extension with recent-link history
- Supports Docker and Render deployment

## Tech Stack

- Go and Gin
- PostgreSQL
- Redis
- Docker
- JavaScript, HTML and CSS

## Run Locally

1. Clone the repository:

   ```bash
   git clone https://github.com/nitnawareanshuman/nimbus.git
   cd nimbus
   ```

2. Create the environment file:

   ```bash
   cp .env.example .env
   ```

3. Start the application:

   ```bash
   docker compose up --build
   ```

The API will be available at `http://localhost:8080`.

## API

Create a short URL:

```http
POST /shorten
Content-Type: application/json

{
  "url": "https://example.com"
}
```

## Chrome Extension

Open `chrome://extensions`, enable **Developer mode**, select **Load unpacked**, and choose the `extension` folder.

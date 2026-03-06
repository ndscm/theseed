# Golink

A lightweight short URL redirection service with path passthrough support.

## Overview

Golink is a URL shortening and redirection service that allows users to create
memorable short links that redirect to longer URLs. It supports path
passthrough, meaning additional path segments after the short link key are
appended to the target URL.

## How It Works

1. User sends a GET request to `/{key}` or `/{key}/extra/part`
2. The server grabs the first path segment as the lookup key
3. Keys are normalized: lowercased, underscores become dashes
4. Valid keys contain only letters, numbers, and dashes
5. If found, remaining path segments are appended to the target URL
6. Returns a **307 redirect** to the final destination
7. If not found, falls back to the webapp

### Example

```
Database entry:
  key: "search"
  target: "https://www.google.com/search?q="

Request:  GET /search/theseed
Response: 307 Redirect → https://www.google.com/search?q=theseed
```

## Architecture

```
┌─────────────┐     ┌─────────────────┐     ┌──────────────────────┐
│   Client    │────▶│  Golink Server  │────▶│     SQL Database     │
│  (Browser)  │◀────│    (Go HTTP)    │◀────│  (Postgres / SQLite) │
└─────────────┘     └─────────────────┘     └──────────────────────┘
                            │
                            │ (fallback on miss)
                            ▼
                    ┌─────────────────┐
                    │     Webapp      │
                    └─────────────────┘
```

### Request Flow

```
GET /{key}/{rest...}
        │
        ▼
┌───────────────────┐
│ Parse URL Path    │
│ key = path[0]     │
│ rest = path[1:]   │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Query Database    │
│ SELECT target     │
│ WHERE key = ?     │
└───────────────────┘
        │
        ├── Found ──────────────────┐
        │                           ▼
        │               ┌───────────────────┐
        │               │ Build Final URL   │
        │               │ target + rest     │
        │               └───────────────────┘
        │                           │
        │                           ▼
        │               ┌───────────────────┐
        │               │ 307 Redirect      │
        │               │ Location: {url}   │
        │               └───────────────────┘
        │
        └── Not Found ──────────────┐
                                    ▼
                        ┌───────────────────┐
                        │ Proxy to Webapp   │
                        └───────────────────┘
```

## Project Structure

- **`database/`** - Database schema and migration tools.
  - **`schema.sql`** - Defines the `golink` table with columns: key (primary),
    target, public, owner, hit_count, created_time, updated_time.
  - **`apply.sh`** - Applies schema migrations to the target database.
  - **`format.sh`** - Formats SQL files for consistency.

- **`go/`** - Go packages.
  - **`golinkdb/`** - Database access layer.
    - **`golink_database_client.go`** - Database connection management. Opens
      connections using configuration from environment/flags.
    - **`link.go`** - CRUD operations for the golink table: InsertLink,
      SelectLinkByKey, UpdateLink, DeleteLink, SelectLinks, IncrementHitCount,
      CountLinks.

- **`proto/`** - Protocol buffer definitions and generated code.
  - **`golink.proto`** - Defines the GolinkService RPC interface with
    CreateLink, GetLink, UpdateLink, DeleteLink, ListLinks methods.
  - **`build.sh`** - Generates Go and TypeScript code from proto definitions.

- **`server/`** - Server entry point.
  - **`golink_server.go`** - Main function that initializes the server,
    registers the gRPC service handlers and the redirect handler, then starts
    listening on port 4656.

- **`service/`** - Business logic layer.
  - **`golink_handler.go`** - HTTP handler for redirect requests. Parses the URL
    path, looks up the key in the database, and returns a 307 redirect. Falls
    back to the webapp on cache miss.
  - **`golink_service.go`** - Implements the GolinkService RPC interface.
    Handles authentication, ownership checks, etag validation, and update masks.
  - **`link_common.go`** - Link entity related utilities: key normalization
    (lowercase, underscores to dashes), proto-row conversion, etag generation.

- **`webapp/`** - Frontend single-page application. Served as fallback when a
  short link is not found. Provides UI for creating, viewing, and managing short
  links.

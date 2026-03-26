# hubspot-cli

Go CLI for HubSpot CRM using Kong framework.

## Build & Run

```bash
go build -o hubspot .
./hubspot --version
./hubspot --help
```

## Project Structure

- `main.go` — Entry point, Kong CLI definition
- `cmd/` — Command implementations (tickets, conversations, contacts, companies, deals, owners, pipelines)
- `internal/api/` — HubSpot REST API client (Bearer token auth)
- `internal/config/` — Config loading from env vars and .env files

## Config

Reads from `HUBSPOT_ACCESS_TOKEN` env var or `~/.config/hubspot/.env`.

## Dependencies

- `github.com/alecthomas/kong` — CLI framework
- `github.com/joho/godotenv` — .env file parsing

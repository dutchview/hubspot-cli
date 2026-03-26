# hubspot-cli

A command-line interface for HubSpot CRM, built in Go.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/) (for building from source)
- A HubSpot account with access to create Private Apps

## Getting a HubSpot Access Token

This CLI authenticates using a **Private App access token**. To create one:

1. Log in to your [HubSpot account](https://app.hubspot.com/)
2. Go to **Settings** (gear icon top-right)
3. In the left sidebar, navigate to **Integrations > Private Apps**
4. Click **Create a private app**
5. Give it a name (e.g. "CLI Access")
6. Go to the **Scopes** tab and enable the scopes you need:
   - `crm.objects.contacts.read` — for contacts commands
   - `crm.objects.companies.read` — for companies commands
   - `crm.objects.deals.read` — for deals commands
   - `tickets` — for ticket read/write
   - `crm.objects.owners.read` — for owners commands
   - `conversations.read` — for reading conversations
   - `conversations.write` — for replying/commenting on conversations
7. Click **Create app**, then **Continue creating**
8. Copy the **Access token** shown on the next screen

> **Note:** You only need to enable scopes for the commands you plan to use.

## Installation

### From source

```bash
go install github.com/dutchview/hubspot-cli@latest
```

### Build locally

```bash
git clone https://github.com/dutchview/hubspot-cli.git
cd hubspot-cli
go build -o hubspot .
```

## Configuration

### Option 1: Config file (recommended)

Create the config directory and file:

```bash
mkdir -p ~/.config/hubspot
echo "HUBSPOT_ACCESS_TOKEN=your_token_here" > ~/.config/hubspot/.env
chmod 600 ~/.config/hubspot/.env
```

### Option 2: Environment variable

```bash
export HUBSPOT_ACCESS_TOKEN=your_token_here
```

### Option 3: Custom config file

```bash
hubspot --config /path/to/your/.env tickets list
```

### Config load order

Configuration is loaded from these locations (later values override earlier):

1. `~/.config/hubspot/.env`
2. `.env` in current directory
3. `HUBSPOT_ACCESS_TOKEN` environment variable
4. Custom file via `--config` flag

Run `hubspot configure` for interactive setup help.

## Usage

### Tickets

```bash
# List tickets
hubspot tickets list
hubspot tickets list -n 50

# Search tickets
hubspot tickets search "login issue"
hubspot tickets search --priority HIGH
hubspot tickets search --pipeline 491723220 --stage 748271074

# Get ticket details (includes contacts, companies, conversations)
hubspot tickets get 123456
hubspot tickets get 123456 --json

# Create ticket
hubspot tickets create -s "New issue" -d "Description here"
hubspot tickets create -s "Bug report" --priority HIGH --pipeline 0

# Update ticket
hubspot tickets update 123456 -s "Updated subject"
hubspot tickets update 123456 --priority LOW --stage 748271075

# Delete ticket
hubspot tickets delete 123456
hubspot tickets delete 123456 --force
```

### Conversations

```bash
# List messages in a thread
hubspot conversations messages 12345678

# Add internal comment (not visible to customer)
hubspot conversations comment 12345678 "Internal note here"

# Reply to customer (auto-detects sender/recipient from thread)
hubspot conversations reply 12345678 "Thanks for reaching out!"
hubspot conversations reply 12345678 "Hello" --recipient customer@example.com
```

### Contacts

```bash
hubspot contacts list
hubspot contacts list -n 50
hubspot contacts search "john"
hubspot contacts get 123456
```

### Companies

```bash
hubspot companies list
hubspot companies search "dutchview"
hubspot companies get 123456
```

### Deals

```bash
hubspot deals list
hubspot deals search "project name"
hubspot deals search --stage closedwon
hubspot deals get 123456
```

### Owners

```bash
hubspot owners list
hubspot owners get 123456
```

### Pipelines

```bash
hubspot pipelines list tickets
hubspot pipelines list deals
hubspot pipelines stages tickets 491723220
```

## JSON Output

All commands support `--json` / `-j` for machine-readable JSON output:

```bash
hubspot tickets get 123456 --json
hubspot contacts search "john" -j | jq '.results[].properties.email'
```

## License

MIT

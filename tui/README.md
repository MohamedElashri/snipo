# Snippy 

A rich terminal user interface (TUI) for [Snipo](https://github.com/MohamedElashri/snipo) - your self-hosted snippet manager.

## Quick Start

### 1. Build

```bash
make build
```

Or manually:
```bash
go build -o bin/snippy ./cmd/snippy
```

### 2. Create API Token

1. Open your Snipo web interface (e.g., `http://localhost:8080`)
2. Go to **Settings** → **API Tokens**
3. Click **Create Token**
4. Set permissions to `admin` or `write`
5. Copy the generated API key

### 3. Configure

```bash
./bin/snippy config
```

Enter your server URL and API key when prompted.

### 4. Run

```bash
./bin/snippy
```

## Installation

### System-wide Install

```bash
make install
```

Then run from anywhere:
```bash
snippy
```

### Manual Install

```bash
sudo cp bin/snippy /usr/local/bin/
sudo chmod +x /usr/local/bin/snippy
```

## Configuration

Configuration is stored at `~/.config/snipo/config.json`:

```json
{
  "server_url": "http://localhost:8080",
  "api_key": "your-api-key-here"
}
```

To reconfigure:
```bash
snippy config
```

## Keybindings

### Navigation
| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `←` / `h` | Previous page |
| `→` / `l` | Next page |
| `enter` | View snippet |
| `esc` | Go back |

### Actions
| Key | Action |
|-----|--------|
| `n` | Create new snippet |
| `e` | Edit snippet (detail view) |
| `d` | Delete snippet (detail view) |
| `f` | Toggle favorite |
| `/` | Search |
| `r` | Refresh list |
| `c` | Copy to clipboard (detail view) |

### Other
| Key | Action |
|-----|--------|
| `?` | Toggle help |
| `q` | Quit |
| `ctrl+c` | Force quit |

## Commands

```bash
# Start TUI (default)
snippy

# Configure server and API key
snippy config

# Show version
snippy version

# Show help
snippy help
```

## Development

### Prerequisites

- Go 1.24+
- Running Snipo server
- API token with appropriate permissions (read is currently enough)

### Build & Run

```bash
# Build
make build

# Run without installing
make run

# Clean build artifacts
make clean

# Run tests
make test
```

## Troubleshooting

### "snippy is not configured"

Run the config command:
```bash
snippy config
```

### "API error: Unauthorized"

Your API key may be invalid. Create a new token and reconfigure:
```bash
snippy config
```

### Connection refused

Ensure your `Snipo` server is running:
```bash
curl http://localhost:8080/health
```

## Technical Stack

- **Language**: Go 1.24+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **UI Components**: [Bubbles](https://github.com/charmbracelet/bubbles)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Syntax Highlighting**: [Chroma](https://github.com/alecthomas/chroma)
- **Markdown Rendering**: [Glamour](https://github.com/charmbracelet/glamour)

## License

Affero General Public License v3.0 (AGPLv3) [LICENSE](../LICENSE)
# Snipo TUI

A rich terminal user interface (TUI) for [Snipo](https://github.com/MohamedElashri/snipo) - your self-hosted snippet manager.

## Features

- 🎨 **Rich TUI** - Beautiful terminal interface built with Bubble Tea
- 🔐 **API Key Authentication** - Secure authentication using API tokens
- 📝 **Full CRUD Operations** - Create, read, update, and delete snippets
- 🔍 **Search & Filter** - Search snippets by content, tags, and language
- ⭐ **Favorites** - Mark and filter your favorite snippets
- 🏷️ **Tags & Folders** - Organize snippets with tags and folders
- 📄 **Pagination** - Navigate through large snippet collections
- ⌨️ **Vim-style Keybindings** - Efficient keyboard navigation (h/j/k/l)
- 🌐 **Cross-platform** - Works on Linux, macOS, and Windows
- 🚀 **Fast & Lightweight** - Written in Go with minimal dependencies

## Quick Start

### 1. Build

```bash
make build
```

Or manually:
```bash
go build -o bin/snipo-tui ./cmd/snipo-tui
```

### 2. Create API Token

1. Open your Snipo web interface (e.g., `http://localhost:8080`)
2. Go to **Settings** → **API Tokens**
3. Click **Create Token**
4. Set permissions to `admin` or `write`
5. Copy the generated API key

### 3. Configure

```bash
./bin/snipo-tui config
```

Enter your server URL and API key when prompted.

### 4. Run

```bash
./bin/snipo-tui
```

## Installation

### System-wide Install

```bash
make install
```

Then run from anywhere:
```bash
snipo-tui
```

### Manual Install

```bash
sudo cp bin/snipo-tui /usr/local/bin/
sudo chmod +x /usr/local/bin/snipo-tui
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
snipo-tui config
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
snipo-tui

# Configure server and API key
snipo-tui config

# Show version
snipo-tui version

# Show help
snipo-tui help
```

## Development

### Prerequisites

- Go 1.24+
- Running Snipo server
- API token with appropriate permissions

### Project Structure

```
tui/
├── cmd/snipo-tui/      # CLI entry point
├── internal/
│   ├── api/            # API client
│   ├── app/            # Application logic
│   ├── config/         # Configuration management
│   └── ui/             # TUI components
├── go.mod
├── Makefile
└── README.md
```

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

### "snipo-tui is not configured"

Run the config command:
```bash
snipo-tui config
```

### "API error: Unauthorized"

Your API key may be invalid. Create a new token and reconfigure:
```bash
snipo-tui config
```

### Connection refused

Ensure your Snipo server is running:
```bash
curl http://localhost:8080/health
```

## Technical Stack

- **Language**: Go 1.24+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **UI Components**: [Bubbles](https://github.com/charmbracelet/bubbles)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)

## License

GNU General Public License v3.0 (GPLv3)

## Related Projects

- [Snipo](https://github.com/MohamedElashri/snipo) - Main project with web interface & API

## Support

For issues and questions:
- Open an issue on [GitHub](https://github.com/MohamedElashri/snipo/issues)
- Check the [main documentation](https://github.com/MohamedElashri/snipo)

# Snippy 

A rich terminal user interface (TUI) for [Snipo](https://github.com/MohamedElashri/snipo) - your self-hosted snippet manager.

## Quick Start

### 1. Install

**Homebrew (macOS / Linux):**
```bash
brew install MohamedElashri/snippy/snippy
```

**Pre-compiled Binary:**
Download the binary for your system from the [Releases page](https://github.com/MohamedElashri/snipo/releases) and place it in your path. We also provide `.deb` and `.rpm` packages for Linux users inside the release assets.

*(To build from source, see the [Development & Manual Build](#development--manual-build) section at the bottom of this page).*

### 2. Create API Token

1. Open your Snipo web interface (e.g., `http://localhost:8080`)
2. Go to **Settings** → **API Tokens**
3. Click **Create Token**
4. Set permissions to `admin` or `write`
5. Copy the generated API key

### 3. Configure & Run

```bash
# Start configuration wizard
snippy config
```

Enter your server URL and API key when prompted. Then, run Snippy!

```bash
# Start Snippy
snippy
```

## Commands

```bash
# Start TUI (default)
snippy

# Configure server and API key
snippy config

# Show version
snippy version

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

## Configuration

Configuration is stored at `~/.config/snipo/config.json`:

```json
{
  "server_url": "http://localhost:8080",
  "api_key": "your-api-key-here"
}
```

To reconfigure from the terminal at any time:
```bash
snippy config
```

## Troubleshooting

### "snippy is not configured"

Run the config command:
```bash
snippy config
```

### "API error: Unauthorized"

Your API key may be invalid. Create a new token from Snipo UI and reconfigure:
```bash
snippy config
```

### Connection refused

Ensure your `Snipo` server is actively running and accessible from your terminal:
```bash
curl http://localhost:8080/health
```

## Development & Manual Build

If you wish to contribute to Snippy or build the binary manually from source, follow these steps.

### Prerequisites

- Go 1.25+
- Running Snipo server
- API token with appropriate permissions (read is currently enough)

### Build with `go install`

If you have Go installed, you can automatically download, compile, and install the latest version to your path:
```bash
go install github.com/MohamedElashri/snipo/tui/cmd/snippy@latest
```

### Build with Makefile (Local Clone)

Clone the repository and run the Makefile tasks:
```bash
# Download dependencies & build binary (bin/snippy)
make build

# Install to /usr/local/bin
make install

# Run the binary without installing
make run

# Clean build artifacts
make clean

# Run tests
make test
```

## Technical Stack

- **Language**: Go 1.25+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **UI Components**: [Bubbles](https://github.com/charmbracelet/bubbles)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Syntax Highlighting**: [Chroma](https://github.com/alecthomas/chroma)
- **Markdown Rendering**: [Glamour](https://github.com/charmbracelet/glamour)

## License

Affero General Public License v3.0 (AGPLv3) [LICENSE](../LICENSE)
# Snippy

Snippy is the terminal client for a running Snipo instance.

## Install

With Homebrew:

```bash
brew install MohamedElashri/snippy/snippy
```

Release binaries and Linux packages are available from the
[Snipo releases page](https://github.com/MohamedElashri/snipo/releases).

## Configure

Create a dedicated `write` API token. Snippy does not require `admin`
permission.

Then run:

```bash
snippy config
snippy
```

The configuration is stored at `~/.config/snipo/config.json` with mode `0600`:

```json
{
  "server_url": "https://snipo.example.com",
  "api_key": "your-api-token"
}
```

The token is stored as plaintext in that file. Protect the user account and use
HTTPS for remote servers.

## Keybindings

| Context | Key | Action |
|---|---|---|
| List | `j`/`↓`, `k`/`↑` | Move selection |
| List | `h`/`←`, `l`/`→` | Previous/next page |
| List | `Enter` | Open snippet |
| List | `n`, `e` | Create/edit snippet |
| List | `f` | Toggle favorite |
| List | `d` or `x` | Delete selected snippet |
| List | `/`, `r`, `s` | Search, refresh, settings |
| Detail | `j`/`↓`, `k`/`↑` | Scroll |
| Detail | `h`/`←`, `l`/`→` | Previous/next file |
| Detail | `c`, `e`, `Esc` | Copy, edit, return |
| Form | `Tab`, `Shift+Tab` | Change field |
| Form | `Ctrl+E` | Open content in `$EDITOR` |
| Form | `Ctrl+S`, `Esc` | Save/cancel |
| Global | `?`, `q`, `Ctrl+C` | Help, quit, force quit |

Deletion from the list is immediate; use a token and server backup policy
appropriate for that behavior.

## Build and test

The minimum Go version is declared in `tui/go.mod`.

```bash
cd tui
make build
make test
make lint
```

Or install directly:

```bash
go install github.com/MohamedElashri/snipo/tui/cmd/snippy@latest
```

## Troubleshooting

- **Not configured:** run `snippy config`.
- **Unauthorized:** create a new token and reconfigure.
- **Insufficient permissions:** use `write` for mutations.
- **Connection refused:** verify the URL and `curl https://host/health`.
- **TLS error:** install the private CA in the system trust store; do not bypass
  certificate validation for normal use.
- **Clipboard error:** install or configure the clipboard provider required by
  your desktop/session.

Snippy uses a 30-second HTTP timeout and sends the token as a Bearer credential.

## License

Snippy is licensed under the repository's [AGPLv3 license](../LICENSE).

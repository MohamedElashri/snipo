# Snipo Browser Extension

Browser extension for saving code snippets to your self-hosted Snipo instance.

## Features

- **One-Click Saving**: Save selected code via context menu
- **Interactive Mode**: Edit details before saving
- **Privacy First**: Your data stays on your server
- **Self-Hosted**: Works with your Snipo instance
- **Organization**: Tags and folders support
- **Site Exclusions**: Ignore specific websites
- **Multi-Language**: Automatic language detection

## Installation

### From Browser Stores (Recommended)

**Chrome Web Store:**
```
Coming soon - extension is being prepared for submission
```

**Firefox Add-ons:**
```
Coming soon - extension is being prepared for submission
```

### Manual Installation (Development)

**Chrome:**
1. Clone the repository
2. Run `./build.sh` 
3. Open Chrome and go to `chrome://extensions/`
4. Enable "Developer mode"
5. Click "Load unpacked"
6. Select the `dist/chrome` folder

**Firefox:**
1. Clone the repository
2. Run `./build.sh`
3. Open Firefox and go to `about:debugging#/runtime/this-firefox`
4. Click "Load Temporary Add-on"
5. Navigate to `dist/firefox` and select `manifest.json`

## Configuration

1. Click the extension icon to open settings
2. Enter your Snipo instance URL (e.g., `http://localhost:8080`)
3. Enter your API key (generate in Snipo Settings → API Tokens)
4. Click "Test Connection" to verify
5. Save configuration

## Usage

### Quick Save
1. Select code on any webpage
2. Right-click and choose "Save to Snipo"
3. Snippet is saved instantly

### Interactive Mode
1. Enable "Interactive Mode" in settings
2. Select code and right-click "Save to Snipo"
3. Edit title, description, language, tags, and folder
4. Click "Save" to confirm

### Ignored Sites
1. Go to extension settings → Ignored Sites tab
2. Add domains where you don't want the extension active
3. Example: `stackoverflow.com`, `github.com`

## Requirements

- A running Snipo instance (self-hosted)
- API key with at least `write` permissions

## Development

### Building

This can be run on any Unix-like system with bash.

```bash
./build.sh
```


This creates:
- `dist/snipo-chrome-v{VERSION}.zip` - Chrome package
- `dist/snipo-firefox-v{VERSION}.zip` - Firefox package
- `dist/snipo-source-v{VERSION}.zip` - Source archive

### Testing

1. Build the extension
2. Load unpacked in Chrome or Firefox (see Manual Installation)
3. Configure with a test Snipo instance
4. Test all features:
   - Context menu saving
   - Interactive mode
   - Settings persistence
   - Ignored sites
   - Connection testing

### Debugging

**Chrome:**
- Extension popup: Right-click → Inspect
- Background script: `chrome://extensions/` → Details → Inspect views
- Content script: Page DevTools → Sources → Content scripts

**Firefox:**
- Extension popup: Right-click → Inspect
- Background script: `about:debugging` → Inspect
- Content script: Page DevTools → Debugger → Sources

## Privacy

This extension:
- Stores settings locally in browser storage
- Communicates only with your configured Snipo instance
- Does NOT collect any analytics or telemetry
- Does NOT send data to third parties
- Does NOT track your browsing

## Permissions

- **activeTab**: Read selected text when saving
- **scripting**: Inject content script for UI
- **contextMenus**: Add "Save to Snipo" to right-click menu
- **storage**: Store configuration locally
- **host_permissions**: Communicate with your Snipo instance

## Browser Compatibility

| Browser | Minimum Version | Status |
|---------|----------------|--------|
| Chrome  | 109+           | ✅ Supported / Untested |
| Edge    | 109+           | ✅ Supported / Untested |
| Firefox | 140+           | ✅ Supported / Tested |
| Safari  | -              | ❌ Not yet supported |

## Troubleshooting

### "Configuration missing" error
- Ensure you've configured instance URL and API key in settings
- Click "Test Connection" to verify

### "Connection failed" error
- Check that your Snipo instance is running
- Verify the instance URL is correct (include `http://` or `https://`)
- Ensure API key has correct permissions
- Check CORS settings if using localhost

### Context menu not appearing
- Reload the page after installing the extension
- Check if the site is in your ignored list
- Verify extension is enabled

### Snippets not saving
- Check browser console for errors
- Verify API key has `write` permissions
- Test API endpoint manually with curl
- Check Snipo instance logs

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly in both Chrome and Firefox
5. Submit a pull request

## License

Affero General Public License v3.0 - see [LICENSE](../LICENSE)

## Links

- **Main Repository**: https://github.com/MohamedElashri/snipo
- **Issues**: https://github.com/MohamedElashri/snipo/issues
- **Documentation**: https://github.com/MohamedElashri/snipo/tree/main/docs

## Support

- Open an issue: https://github.com/MohamedElashri/snipo/issues
- Read the docs: https://github.com/MohamedElashri/snipo/tree/main/docs
- Check the FAQ: https://github.com/MohamedElashri/snipo#faq

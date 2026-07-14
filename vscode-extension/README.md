# Snipo VS Code Extension

**Note: This extension is currently under publication and will be available on the VS Code Marketplace soon.**

Snipo is a modern snippet manager designed to boost your productivity. This VS Code extension integrates your Snipo snippets directly into your editor, allowing you to access, search, insert, and save snippets without ever leaving your IDE.

## Installation

While the extension is under publication, you can install it manually from the source:

1. Download or generate the `snipo-*.vsix` file.
2. In VS Code, open the Extensions view (`Ctrl+Shift+X` or `Cmd+Shift+X`).
3. Click the `...` menu in the top right corner of the Extensions view.
4. Select **Install from VSIX...**
5. Locate and select the `snipo-*.vsix` file.

Alternatively, you can install it via the command line if you have `code` in your PATH:

```bash
code --install-extension snipo-*.vsix
```

## Configuration

Once installed, Snipo needs to connect to your backend:
1. Click on the Snipo icon in the Activity Bar.
2. Click the **Configure Snipo** gear icon in the title menu.
3. Enter your Snipo Server URL (e.g. `http://localhost:8080`).
4. Enter your API Token.

You're all set!

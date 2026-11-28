# Snipo VS Code Extension

[![Leave a tip on Ko-fi](https://img.shields.io/badge/Ko--fi-Leave%20a%20tip-913F51?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/melashri)

[![VS Code Marketplace](https://img.shields.io/badge/VS_Code_Marketplace-available-blue.svg?logo=visualstudiocode)](https://marketplace.visualstudio.com/items?itemName=muhammadelashri.snipo)

Snipo is a modern snippet manager designed to boost your productivity. This VS Code extension integrates your Snipo snippets directly into your editor, allowing you to access, search, insert, and save snippets without ever leaving your IDE.

## Installation

You can install the extension directly from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=muhammadelashri.snipo):

1. Open the Extensions view in VS Code (`Ctrl+Shift+X` or `Cmd+Shift+X`).
2. Search for "Snipo".
3. Click **Install**.

Alternatively, you can install it via the command line:

```bash
code --install-extension muhammadelashri.snipo
```

### Manual Installation

If you prefer to install from source:

1. Download or generate the `snipo-*.vsix` file.
2. In VS Code, open the Extensions view.
3. Click the `...` menu in the top right corner and select **Install from VSIX...**.
4. Locate and select the `snipo-*.vsix` file.

## Configuration

Once installed, Snipo needs to connect to your backend:
1. Click on the Snipo icon in the Activity Bar.
2. Click the **Configure Snipo** gear icon in the title menu.
3. Enter your Snipo Server URL (e.g. `http://localhost:8080`).
4. Enter your API Token.

## Notes

Open the **Notes** view to browse saved notes or start one with **Snipo: New Note**.
Choose Markdown or plain text and write in the editor. **Snipo: Preview Markdown Note**
opens a side-by-side preview. Use **Snipo: Save Note** to save a new note or send
changes to an existing note back to your server.

## Templates

Run **Snipo: Use Template** in the Command Palette. Pick a template and fill in
its fields. Review the generated files, open them as untitled tabs, insert a
file into the active editor, or save the whole draft to Snipo. Manage templates
in the Snipo web library.

## Support

Tips are optional and help maintain the project.

**[Leave a tip on Ko-fi](https://ko-fi.com/melashri)**

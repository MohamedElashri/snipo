# Snipo browser extension

The Chrome and Firefox extension saves selected text to a self-hosted Snipo
instance.

## Install

- [Chrome Web Store](https://chromewebstore.google.com/detail/snipo-code-snippet-manage/gagllhkjibjhllhkmiphjpahbgnmgana)
- [Firefox Add-ons](https://addons.mozilla.org/firefox/addon/snipo-code-snippet-manager/)

For local development, follow [Build](#build), then load
`extension/dist/chrome` as an unpacked Chrome extension or
`extension/dist/firefox/manifest.json` as a temporary Firefox add-on.

## Configure

1. In Snipo, create a dedicated API token with `write` permission.
2. Open the extension options.
3. Enter the complete Snipo base URL, including any configured subpath.
4. Enter the API token and test the connection.
5. Save.

Use HTTPS when the browser and Snipo communicate over a network. The token is a
credential with create, update, and delete access; do not use an admin token.

## Use

Select text on a page and choose **Save to Snipo** from the context menu.

- Quick mode immediately creates a private, single-file snippet.
- Interactive mode lets you edit the title, description, filename, language,
  folder, and tags first.
- Ignored sites prevent the extension UI from acting on matching domains.

The page title and URL are used as default snippet metadata.

## Permissions and privacy

The [privacy policy](PRIVACY.md) is the canonical disclosure for data handling,
browser storage, host access, and manifest permissions.

## Build

Use the Node version in `.node-version`, npm, Bash, and `zip`. From the
repository root (or the extracted Firefox source archive):

```bash
npm ci
npm run build -w @snipo/api-client
bash extension/build.sh all
```

Output is written to `extension/dist`. Pass `chrome` or `firefox` instead of
`all` to build one browser package. The source archive includes the shared
client and workspace lockfile needed to reproduce the Firefox bundle.

## Troubleshooting

- **Configuration missing:** save both the base URL and API token.
- **Unauthorized:** replace the token and confirm it has `write` permission.
- **Connection failed:** open the base URL in the same browser, verify HTTPS
  trust, and check the Snipo and extension developer consoles.
- **Context menu or overlay missing:** reload tabs opened before installation
  and confirm the domain is not ignored.
- **Save rejected:** verify that selected content is below Snipo's request and
  file-size limits.

## License

The extension is licensed under the repository's [AGPLv3 license](../LICENSE).

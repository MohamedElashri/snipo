# Snipo browser extension privacy policy

Last updated: July 1, 2026

## Data handled

The extension handles:

- text you explicitly select for saving;
- the current page title and URL, used as default snippet metadata;
- the Snipo server URL and API token you configure;
- interactive-mode and ignored-site preferences;
- titles, descriptions, filenames, languages, folders, and tags entered in the
  save dialog.

## Storage and transmission

Configuration and preferences are stored in the browser's synchronized
extension storage. If browser sync is enabled, the browser vendor may synchronize
that data, including the API token, under the vendor's own privacy and security
terms.

When you save a snippet, its content and metadata are sent directly to the
Snipo URL you configured. Connection tests and folder/tag lookups also contact
that instance.

The extension contains no advertising, analytics, telemetry, or developer-run
backend. It does not intentionally send data to the extension developer or
other third parties. Network intermediaries and the operator of the configured
Snipo instance can handle traffic according to their own policies.

## Permissions

The extension requests `activeTab`, `contextMenus`, `storage`, and host access
to HTTP and HTTPS pages. Broad host access is required because the extension
can run on pages where text is selected and must connect to an arbitrary
self-hosted Snipo URL. The extension does not request the `scripting`
permission.

## Security choices

The API token is not encrypted separately from browser storage. Use a dedicated,
expiring Snipo token with `write` permission and revoke it if a synchronized
browser account or device is compromised.

The extension uses the URL scheme you configure. Use HTTPS except for a
loopback-only development instance. The security and retention of saved
snippets are controlled by the configured Snipo operator.

## Control and deletion

You can:

- change or replace configuration in the extension options;
- revoke the API token in Snipo;
- clear extension storage through the browser;
- remove the extension; and
- delete saved snippets from Snipo.

Removing the extension does not delete snippets already sent to Snipo. Browser
sync may retain synchronized settings according to the browser vendor's
policies.

## Changes and contact

Policy changes are published with the extension source. Questions can be filed
in the [Snipo issue tracker](https://github.com/MohamedElashri/snipo/issues).
Report security vulnerabilities privately as described in
[SECURITY.md](../SECURITY.md).

# Privacy Policy for Snipo Browser Extension

**Last Updated:** January 2, 2026

## Overview

Snipo is a browser extension that allows you to save code snippets to your self-hosted Snipo instance. This extension is designed with privacy as a core principle.

## Data Collection and Usage

### What Data We Collect

The Snipo extension collects and processes the following data **locally on your device**:

1. **Selected Text/Code**: When you select text on a webpage and choose to save it as a snippet
2. **Page Metadata**: The title and URL of the webpage where you save a snippet (used for context)
3. **Configuration Data**: Your self-hosted Snipo instance URL and API key (stored locally in browser storage)
4. **User Preferences**: Settings like Interactive Mode and ignored websites list

### How We Use Your Data

- **Selected text/code** is sent directly to YOUR self-hosted Snipo instance (the server you configure)
- **Configuration data** is stored locally in your browser's sync storage
- **No data is sent to any third-party servers or the extension developers**
- **No analytics, tracking, or telemetry** is collected by this extension

## Data Storage

- All extension settings are stored using Chrome's `chrome.storage.sync` API
- Data is synchronized across your browsers if you're signed into Chrome/Firefox sync
- Your API key is stored securely in browser storage (not accessible to websites)

## Data Transmission

The extension communicates ONLY with:
- **Your self-hosted Snipo instance** (the URL you configure in settings)

The extension does NOT communicate with:
- Extension developers' servers
- Third-party analytics services
- Advertising networks
- Any other external services

## Permissions Explanation

The extension requires the following permissions:

- **activeTab**: To read selected text from the current webpage when you save a snippet
- **scripting**: To inject the content script that enables snippet saving functionality
- **contextMenus**: To add "Save to Snipo" option to the right-click context menu
- **storage**: To save your configuration (instance URL, API key, preferences)
- **host_permissions (http://*/*, https://*/*)**: To allow the extension to work on any website and communicate with your self-hosted instance

## Third-Party Services

This extension does NOT use any third-party services, analytics, or tracking tools.

## Your Self-Hosted Instance

Since you control your own Snipo instance:
- You control where your data is stored
- You control who has access to your snippets
- You control the security and privacy settings of your instance

Please refer to your Snipo instance's privacy practices and security configuration.

## Data Security

- API keys are stored in browser storage and transmitted only to your configured instance
- All communication with your Snipo instance uses the protocol you configure (HTTP or HTTPS)
- We recommend using HTTPS for your Snipo instance in production

## Children's Privacy

This extension is not directed at children under the age of 13. We do not knowingly collect personal information from children.

## Changes to This Privacy Policy

We may update this privacy policy from time to time. Any changes will be reflected in the extension's repository and the "Last Updated" date above.

## Open Source

This extension is open source. You can review the complete source code at:
https://github.com/MohamedElashri/snipo

## Contact

If you have questions about this privacy policy, please open an issue on our GitHub repository:
https://github.com/MohamedElashri/snipo/issues

## Your Rights

Since this extension stores data locally and only communicates with your self-hosted instance:
- You can delete all extension data by removing the extension
- You can clear stored settings through the browser's extension management interface
- You have complete control over your snippet data through your self-hosted instance

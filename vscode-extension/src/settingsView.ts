import * as vscode from 'vscode';
import { api } from './api';

export class SettingsViewProvider implements vscode.WebviewViewProvider {
    public static readonly viewType = 'snipo-settings';
    private _view?: vscode.WebviewView;

    constructor(
        private readonly _extensionUri: vscode.Uri,
        private readonly _context: vscode.ExtensionContext
    ) { }

    public resolveWebviewView(
        webviewView: vscode.WebviewView,
        context: vscode.WebviewViewResolveContext,
        _token: vscode.CancellationToken,
    ) {
        this._view = webviewView;

        webviewView.webview.options = {
            enableScripts: true,
            localResourceRoots: [this._extensionUri]
        };

        webviewView.webview.html = this._getHtmlForWebview(webviewView.webview);

        webviewView.webview.onDidReceiveMessage(async data => {
            switch (data.type) {
                case 'saveSettings': {
                    const { apiUrl, apiToken } = data.value;
                    if (!apiUrl || !apiToken) {
                        vscode.window.showErrorMessage('Server URL and API Token are required.');
                        return;
                    }

                    webviewView.webview.postMessage({ type: 'saving' });

                    const isValid = await vscode.window.withProgress({
                        location: vscode.ProgressLocation.Notification,
                        title: "Verifying Snipo connection...",
                        cancellable: false
                    }, async () => {
                        return await api.verifyConfiguration(apiUrl, apiToken);
                    });

                    if (isValid) {
                        const config = vscode.workspace.getConfiguration('snipo');
                        await config.update('apiUrl', apiUrl, vscode.ConfigurationTarget.Global);
                        await this._context.secrets.store('snipo.apiToken', apiToken);
                        vscode.window.showInformationMessage('Successfully connected to Snipo!');
                        webviewView.webview.postMessage({ type: 'success' });
                        vscode.commands.executeCommand('snipo.refreshTree');
                    } else {
                        vscode.window.showErrorMessage('Connection failed or invalid token. Please check your URL and Token.');
                        webviewView.webview.postMessage({ type: 'error' });
                    }
                    break;
                }
            }
        });
    }

    private _getHtmlForWebview(webview: vscode.Webview) {
        const config = vscode.workspace.getConfiguration('snipo');
        const currentUrl = config.get<string>('apiUrl') || '';

        return `<!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <title>Snipo Configuration</title>
                <style>
                    body {
                        font-family: var(--vscode-font-family);
                        color: var(--vscode-foreground);
                        padding: 10px;
                        display: flex;
                        flex-direction: column;
                        gap: 15px;
                    }
                    .input-group {
                        display: flex;
                        flex-direction: column;
                        gap: 5px;
                    }
                    label {
                        font-size: 13px;
                        font-weight: 600;
                    }
                    input {
                        background-color: var(--vscode-input-background);
                        color: var(--vscode-input-foreground);
                        border: 1px solid var(--vscode-input-border, transparent);
                        padding: 6px;
                        font-family: var(--vscode-font-family);
                        border-radius: 2px;
                        outline: none;
                    }
                    input:focus {
                        border: 1px solid var(--vscode-focusBorder);
                    }
                    button {
                        background-color: var(--vscode-button-background);
                        color: var(--vscode-button-foreground);
                        border: none;
                        padding: 8px;
                        cursor: pointer;
                        font-family: var(--vscode-font-family);
                        font-weight: bold;
                        border-radius: 2px;
                        margin-top: 10px;
                    }
                    button:hover {
                        background-color: var(--vscode-button-hoverBackground);
                    }
                    button:disabled {
                        opacity: 0.6;
                        cursor: not-allowed;
                    }
                    #status {
                        margin-top: 10px;
                        font-size: 13px;
                    }
                    .error { color: var(--vscode-errorForeground); }
                    .success { color: var(--vscode-notificationsInfoIcon-foreground); }
                </style>
            </head>
            <body>
                <div class="input-group">
                    <label for="apiUrl">Server URL</label>
                    <input type="text" id="apiUrl" placeholder="e.g. http://localhost:3000" value="${currentUrl}" />
                </div>
                
                <div class="input-group">
                    <label for="apiToken">API Token</label>
                    <input type="password" id="apiToken" placeholder="Your Snipo API Token" />
                </div>

                <button id="saveBtn">Connect</button>
                <div id="status"></div>

                <script>
                    const vscode = acquireVsCodeApi();
                    const saveBtn = document.getElementById('saveBtn');
                    const apiUrlInput = document.getElementById('apiUrl');
                    const apiTokenInput = document.getElementById('apiToken');
                    const statusDiv = document.getElementById('status');

                    saveBtn.addEventListener('click', () => {
                        const apiUrl = apiUrlInput.value.trim();
                        const apiToken = apiTokenInput.value.trim();
                        
                        if (!apiUrl || !apiToken) {
                            statusDiv.textContent = 'Please fill out both fields.';
                            statusDiv.className = 'error';
                            return;
                        }

                        vscode.postMessage({
                            type: 'saveSettings',
                            value: { apiUrl, apiToken }
                        });
                    });

                    window.addEventListener('message', event => {
                        const message = event.data;
                        switch (message.type) {
                            case 'saving':
                                saveBtn.disabled = true;
                                saveBtn.textContent = 'Connecting...';
                                statusDiv.textContent = '';
                                break;
                            case 'success':
                                saveBtn.disabled = false;
                                saveBtn.textContent = 'Update Connection';
                                statusDiv.textContent = 'Connected successfully!';
                                statusDiv.className = 'success';
                                apiTokenInput.value = '';
                                break;
                            case 'error':
                                saveBtn.disabled = false;
                                saveBtn.textContent = 'Connect';
                                statusDiv.textContent = 'Connection failed.';
                                statusDiv.className = 'error';
                                break;
                        }
                    });
                </script>
            </body>
            </html>`;
    }
}

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

        const config = vscode.workspace.getConfiguration('snipo');
        const currentUrl = config.get<string>('apiUrl') || '';
        
        const sendInit = () => {
            const config = vscode.workspace.getConfiguration('snipo');
            const url = config.get<string>('apiUrl') || '';
            api.isConfigured().then(isConfigured => {
                webviewView.webview.postMessage({ type: 'init', value: { apiUrl: url, isConfigured } });
            });
        };

        // Initialize the webview securely
        sendInit();

        // Re-initialize when the view becomes visible (e.g., after being collapsed)
        webviewView.onDidChangeVisibility(() => {
            if (webviewView.visible) {
                sendInit();
            }
        });

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
                        webviewView.webview.postMessage({ type: 'success', value: { apiUrl } });
                        vscode.commands.executeCommand('snipo.refreshTree');
                    } else {
                        vscode.window.showErrorMessage('Connection failed or invalid token. Please check your URL and Token.');
                        webviewView.webview.postMessage({ type: 'error' });
                    }
                    break;
                }
                case 'disconnect': {
                    const config = vscode.workspace.getConfiguration('snipo');
                    await config.update('apiUrl', undefined, vscode.ConfigurationTarget.Global);
                    await this._context.secrets.delete('snipo.apiToken');
                    vscode.window.showInformationMessage('Disconnected from Snipo.');
                    webviewView.webview.postMessage({ type: 'disconnected' });
                    vscode.commands.executeCommand('snipo.refreshTree');
                    break;
                }
                case 'openDashboard': {
                    const config = vscode.workspace.getConfiguration('snipo');
                    const currentUrl = config.get<string>('apiUrl');
                    if (currentUrl) {
                        vscode.env.openExternal(vscode.Uri.parse(currentUrl));
                    }
                    break;
                }
            }
        });
    }

    private getNonce() {
        let text = '';
        const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        for (let i = 0; i < 32; i++) {
            text += possible.charAt(Math.floor(Math.random() * possible.length));
        }
        return text;
    }

    private _getHtmlForWebview(webview: vscode.Webview) {
        const nonce = this.getNonce();

        return `<!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}';">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <title>Snipo Configuration</title>
                <style>
                    :root {
                        --brand-primary: #6366F1;
                        --brand-hover: #4F46E5;
                        --brand-danger: #ef4444;
                        --brand-danger-hover: #dc2626;
                        --brand-bg: rgba(99, 102, 241, 0.08);
                        --border-radius: 6px;
                        --transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
                    }
                    body {
                        font-family: 'Inter', var(--vscode-font-family);
                        color: var(--vscode-foreground);
                        padding: 16px 20px;
                        display: flex;
                        flex-direction: column;
                        gap: 24px;
                        margin: 0;
                    }
                    
                    /* Brand Header */
                    .brand-header {
                        display: flex;
                        align-items: center;
                        gap: 12px;
                        margin-bottom: 4px;
                    }
                    .brand-logo {
                        width: 32px;
                        height: 32px;
                        background: linear-gradient(135deg, var(--brand-primary), #818cf8);
                        border-radius: 8px;
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        color: white;
                        font-weight: 900;
                        font-size: 18px;
                        box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
                        user-select: none;
                    }
                    .brand-logo svg {
                        width: 18px;
                        height: 18px;
                    }
                    .brand-title {
                        font-size: 18px;
                        font-weight: 600;
                        margin: 0;
                        letter-spacing: -0.3px;
                    }

                    /* Inputs */
                    .input-group {
                        display: flex;
                        flex-direction: column;
                        gap: 6px;
                    }
                    label {
                        font-size: 11px;
                        font-weight: 600;
                        color: var(--vscode-descriptionForeground);
                        text-transform: uppercase;
                        letter-spacing: 0.5px;
                    }
                    input {
                        background-color: var(--vscode-input-background);
                        color: var(--vscode-input-foreground);
                        border: 1px solid var(--vscode-input-border, transparent);
                        padding: 10px 12px;
                        border-radius: var(--border-radius);
                        outline: none;
                        transition: var(--transition);
                        font-size: 13px;
                        font-family: var(--vscode-font-family);
                    }
                    input:focus {
                        border-color: var(--brand-primary);
                        box-shadow: 0 0 0 2px var(--brand-bg);
                    }

                    /* Buttons */
                    .button-group {
                        display: flex;
                        gap: 8px;
                        flex-wrap: wrap;
                        margin-top: 4px;
                    }
                    button {
                        background-color: var(--brand-primary);
                        color: white;
                        border: none;
                        padding: 10px 16px;
                        cursor: pointer;
                        font-weight: 600;
                        font-size: 13px;
                        border-radius: var(--border-radius);
                        transition: var(--transition);
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        gap: 8px;
                        flex: 1;
                        min-width: 120px;
                        font-family: var(--vscode-font-family);
                    }
                    button:hover {
                        background-color: var(--brand-hover);
                        transform: translateY(-1px);
                        box-shadow: 0 4px 12px rgba(99, 102, 241, 0.25);
                    }
                    button:active {
                        transform: translateY(0);
                    }
                    button:disabled {
                        opacity: 0.6;
                        cursor: not-allowed;
                        transform: none;
                        box-shadow: none;
                    }
                    button.secondary {
                        background-color: transparent;
                        color: var(--vscode-foreground);
                        border: 1px solid var(--vscode-widget-border);
                    }
                    button.secondary:hover {
                        background-color: var(--vscode-button-secondaryHoverBackground);
                        box-shadow: none;
                    }
                    button.danger {
                        background-color: transparent;
                        color: var(--brand-danger);
                        border: 1px solid var(--brand-danger);
                    }
                    button.danger:hover {
                        background-color: var(--brand-danger);
                        color: white;
                        box-shadow: 0 4px 12px rgba(239, 68, 68, 0.25);
                    }

                    /* Connected State */
                    #connectedState {
                        display: none;
                        flex-direction: column;
                        gap: 16px;
                        background: var(--brand-bg);
                        border: 1px solid rgba(99, 102, 241, 0.15);
                        border-radius: 8px;
                        padding: 20px;
                        position: relative;
                        overflow: hidden;
                    }
                    #connectedState::before {
                        content: '';
                        position: absolute;
                        top: 0; left: 0; right: 0;
                        height: 3px;
                        background: linear-gradient(90deg, var(--brand-primary), #a855f7);
                    }
                    #connectedState h3 {
                        margin: 0;
                        font-size: 14px;
                        display: flex;
                        align-items: center;
                        gap: 8px;
                        color: var(--vscode-foreground);
                    }
                    .server-info {
                        display: flex;
                        flex-direction: column;
                        gap: 6px;
                    }
                    #connectedState p {
                        margin: 0;
                        font-size: 12px;
                        font-family: var(--vscode-editor-font-family);
                        word-break: break-all;
                        background: var(--vscode-editor-background);
                        padding: 10px 12px;
                        border-radius: var(--border-radius);
                        border: 1px solid var(--vscode-widget-border);
                        color: var(--vscode-foreground);
                    }
                    #configForm {
                        display: flex;
                        flex-direction: column;
                        gap: 16px;
                    }
                    #status {
                        font-size: 13px;
                        text-align: center;
                        margin-top: -4px;
                    }
                    .error { color: var(--brand-danger); }
                    .success { color: var(--brand-primary); }
                </style>
            </head>
            <body>
                <div class="brand-header">
                    <div class="brand-logo">
                        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M8 4L4 12L8 20" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                            <path d="M16 4L20 12L16 20" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                            <path d="M14 4L10 20" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                        </svg>
                    </div>
                    <h2 class="brand-title">Snipo</h2>
                </div>

                <div id="connectedState">
                    <h3>✅ Active Connection</h3>
                    <div class="server-info">
                        <label>Server URL</label>
                        <p id="connectedUrl"></p>
                    </div>
                    <div class="button-group">
                        <button id="dashboardBtn">🌐 Dashboard</button>
                        <button id="modifyBtn" class="secondary">✏️ Modify</button>
                    </div>
                    <div class="button-group">
                        <button id="disconnectBtn" class="danger">🚪 Disconnect</button>
                    </div>
                </div>

                <div id="configForm">
                    <div class="input-group">
                        <label for="apiUrl">Server URL</label>
                        <input type="text" id="apiUrl" placeholder="e.g. http://localhost:3000" />
                    </div>
                    
                    <div class="input-group">
                        <label for="apiToken">API Token</label>
                        <input type="password" id="apiToken" placeholder="Your Snipo API Token" />
                    </div>

                    <div class="button-group">
                        <button id="saveBtn">Connect</button>
                        <button id="cancelBtn" class="secondary" style="display: none;">Cancel</button>
                    </div>
                    <div id="status"></div>
                </div>

                <script nonce="${nonce}">
                    const vscode = acquireVsCodeApi();
                    let isCurrentlyConfigured = false;

                    const saveBtn = document.getElementById('saveBtn');
                    const cancelBtn = document.getElementById('cancelBtn');
                    const modifyBtn = document.getElementById('modifyBtn');
                    const disconnectBtn = document.getElementById('disconnectBtn');
                    const dashboardBtn = document.getElementById('dashboardBtn');
                    
                    const apiUrlInput = document.getElementById('apiUrl');
                    const apiTokenInput = document.getElementById('apiToken');
                    const statusDiv = document.getElementById('status');
                    
                    const connectedState = document.getElementById('connectedState');
                    const configForm = document.getElementById('configForm');
                    const connectedUrl = document.getElementById('connectedUrl');

                    function showConnectedState(url) {
                        isCurrentlyConfigured = true;
                        connectedState.style.display = 'flex';
                        configForm.style.display = 'none';
                        connectedUrl.textContent = url || apiUrlInput.value;
                    }

                    function showConfigForm() {
                        connectedState.style.display = 'none';
                        configForm.style.display = 'flex';
                        statusDiv.textContent = '';
                        cancelBtn.style.display = isCurrentlyConfigured ? 'block' : 'none';
                    }

                    modifyBtn.addEventListener('click', showConfigForm);
                    
                    cancelBtn.addEventListener('click', () => {
                        showConnectedState(apiUrlInput.value);
                    });

                    disconnectBtn.addEventListener('click', () => {
                        if (confirm('Are you sure you want to disconnect?')) {
                            vscode.postMessage({ type: 'disconnect' });
                        }
                    });

                    dashboardBtn.addEventListener('click', () => {
                        vscode.postMessage({ type: 'openDashboard' });
                    });

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
                            case 'init':
                                apiUrlInput.value = message.value.apiUrl || '';
                                if (message.value.isConfigured) {
                                    showConnectedState(message.value.apiUrl);
                                } else {
                                    isCurrentlyConfigured = false;
                                    showConfigForm();
                                }
                                break;
                            case 'saving':
                                saveBtn.disabled = true;
                                saveBtn.textContent = 'Connecting...';
                                statusDiv.textContent = '';
                                break;
                            case 'success':
                                saveBtn.disabled = false;
                                saveBtn.textContent = 'Connect';
                                apiTokenInput.value = '';
                                showConnectedState(message.value.apiUrl);
                                break;
                            case 'error':
                                saveBtn.disabled = false;
                                saveBtn.textContent = 'Connect';
                                statusDiv.textContent = 'Connection failed.';
                                statusDiv.className = 'error';
                                break;
                            case 'disconnected':
                                isCurrentlyConfigured = false;
                                apiUrlInput.value = '';
                                apiTokenInput.value = '';
                                showConfigForm();
                                statusDiv.textContent = 'Disconnected successfully.';
                                statusDiv.className = 'success';
                                break;
                        }
                    });
                </script>
            </body>
            </html>`;
    }
}

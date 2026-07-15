import * as assert from 'assert';
import * as vscode from 'vscode';
import { SettingsViewProvider } from '../settingsView';

suite('SettingsView Test Suite', () => {
    test('Provider renders HTML', () => {
        const mockContext: any = {
            secrets: { store: () => Promise.resolve() }
        };
        const mockUri: any = vscode.Uri.parse('file:///tmp');
        
        const provider = new SettingsViewProvider(mockUri, mockContext);
        
        const mockWebview: any = {
            options: {},
            html: '',
            onDidReceiveMessage: () => {}
        };
        const mockWebviewView: any = {
            webview: mockWebview
        };

        provider.resolveWebviewView(mockWebviewView, {} as any, {} as any);

        assert.strictEqual(typeof mockWebview.html, 'string');
        assert.ok(mockWebview.html.includes('<html lang="en">'));
        assert.ok(mockWebview.html.includes('id="apiUrl"'));
        assert.ok(mockWebview.html.includes('id="apiToken"'));
    });
});

import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { SettingsViewProvider } from '../settingsView';

suite('SettingsView Test Suite', () => {
    test('Provider renders HTML', () => {
        const mockContext: any = {
            secrets: { store: () => Promise.resolve() }
        };
        const mockUri: any = vscode.Uri.parse('file:///tmp');
        
        const provider = new SettingsViewProvider(mockUri, mockContext);
        
        const mockWebview: any = {
            cspSource: 'https://vscode-webview.net',
            options: {},
            html: '',
            onDidReceiveMessage: () => {},
            postMessage: () => Promise.resolve(true)
        };
        const mockWebviewView: any = {
            webview: mockWebview,
            onDidChangeVisibility: () => {}
        };

        provider.resolveWebviewView(mockWebviewView, {} as any, {} as any);

        assert.strictEqual(typeof mockWebview.html, 'string');
        assert.ok(mockWebview.html.includes('<html lang="en">'));
        assert.ok(mockWebview.html.includes('id="apiUrl"'));
        assert.ok(mockWebview.html.includes('id="apiToken"'));
    });
    test('dashboard link normalizes a scheme-less server URL', async () => {
        let receive: (message: any) => Promise<void> = async () => {};
        const provider=new SettingsViewProvider(vscode.Uri.file('/tmp'),{} as any);
        const view:any={webview:{cspSource:'test',onDidReceiveMessage:(fn:any)=>{receive=fn},postMessage:async()=>true},onDidChangeVisibility(){}};
        const get=sinon.stub(vscode.workspace,'getConfiguration').returns({inspect:()=>({globalValue:' localhost:9288/snipo/ '})} as any);
        const open=sinon.stub(vscode.env,'openExternal').resolves(true);
        try {
            provider.resolveWebviewView(view,{} as any,{} as any);
            await receive({type:'openDashboard'});
            assert.strictEqual(open.firstCall.args[0].scheme,'http');
            assert.strictEqual(open.firstCall.args[0].authority,'localhost:9288');
            assert.strictEqual(open.firstCall.args[0].path,'/snipo/');
        } finally {get.restore();open.restore();}
    });

});

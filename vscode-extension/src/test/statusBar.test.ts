import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { SnipoStatusBar } from '../statusBar';
import { api } from '../api';

suite('StatusBar Test Suite', () => {
    let mockContext: any;
    let createStatusBarItemStub: sinon.SinonStub;

    setup(() => {
        mockContext = {
            subscriptions: [],
            secrets: { onDidChange: sinon.spy() }
        };
        
        createStatusBarItemStub = sinon.stub(vscode.window, 'createStatusBarItem').returns({
            text: '',
            tooltip: '',
            command: '',
            show: sinon.spy(),
            hide: sinon.spy(),
            dispose: sinon.spy()
        } as any);
    });

    teardown(() => {
        sinon.restore();
    });

    test('StatusBar initializes correctly', async () => {
        sinon.stub(api, 'isConfigured').resolves(true);
        const statusBar = new SnipoStatusBar(mockContext);
        
        // Wait for update promises to resolve
        await new Promise(resolve => setTimeout(resolve, 0));
        
        assert.strictEqual(createStatusBarItemStub.calledOnce, true);
        const statusBarItem = createStatusBarItemStub.returnValues[0];
        assert.strictEqual(statusBarItem.text, '$(repo) Snipo');
        assert.strictEqual(statusBarItem.command, 'snipo.searchAndInsert');
        assert.strictEqual(statusBarItem.show.calledOnce, true);
    });
});

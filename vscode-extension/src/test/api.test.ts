import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { api } from '../api';

suite('SnipoAPI Test Suite', () => {
    vscode.window.showInformationMessage('Start all API tests.');

    const apiUrl = 'http://127.0.0.1:3000';
    const apiToken = 'test-token';

    let fetchStub: sinon.SinonStub;

    setup(async () => {
        const config = vscode.workspace.getConfiguration('snipo');
        await config.update('apiUrl', apiUrl, vscode.ConfigurationTarget.Global);
        
        const ext = vscode.extensions.getExtension('muhammadelashri.snipo');
        if (ext && !ext.isActive) {
            await ext.activate();
        }
        
        fetchStub = sinon.stub(global, 'fetch');
    });

    teardown(() => {
        sinon.restore();
    });

    test('verifyConfiguration - success', async () => {
        fetchStub.resolves({ ok: true, status: 200, json: async () => ({}) });

        const isValid = await api.verifyConfiguration(apiUrl, apiToken);
        assert.strictEqual(isValid, true);
        assert.strictEqual(fetchStub.called, true);
    });

    test('verifyConfiguration - failure', async () => {
        fetchStub.rejects(new Error('Network error'));

        const isValid = await api.verifyConfiguration(apiUrl, apiToken);
        assert.strictEqual(isValid, false);
    });

    test('getSnippets - success', async () => {
        fetchStub.resolves({
            ok: true,
            status: 200,
            json: async () => ({
                data: [
                    { id: '1', title: 'Test Snippet 1', content: 'console.log("test")', language: 'javascript' }
                ]
            })
        });

        const snippets = await api.getSnippets();
        assert.strictEqual(snippets.length, 1);
        assert.strictEqual(snippets[0].title, 'Test Snippet 1');
    });

    test('searchSnippets - abort signal', async () => {
        const cancelError = new Error('Canceled');
        cancelError.name = 'AbortError';
        
        fetchStub.rejects(cancelError);

        const abortController = new AbortController();
        const searchPromise = api.searchSnippets('test', abortController.signal);
        
        abortController.abort();
        
        const snippets = await searchPromise;
        // Aborted requests should return empty array based on our catch block
        assert.deepStrictEqual(snippets, []);
    });
});

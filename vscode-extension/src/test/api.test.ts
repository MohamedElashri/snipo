import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { api } from '../api';
import axios from 'axios';

suite('SnipoAPI Test Suite', () => {
    vscode.window.showInformationMessage('Start all API tests.');

    const apiUrl = 'http://127.0.0.1:3000';
    const apiToken = 'test-token';

    let axiosCreateStub: sinon.SinonStub;
    let clientGetStub: sinon.SinonStub;

    setup(async () => {
        const config = vscode.workspace.getConfiguration('snipo');
        await config.update('apiUrl', apiUrl, vscode.ConfigurationTarget.Global);
        
        clientGetStub = sinon.stub((api as any).client, 'get');
        
        axiosCreateStub = sinon.stub(axios, 'create').returns({
            get: clientGetStub
        } as any);
    });

    teardown(() => {
        sinon.restore();
    });

    test('verifyConfiguration - success', async () => {
        const tempClientGetStub = sinon.stub().resolves({ status: 200 });
        axiosCreateStub.returns({ get: tempClientGetStub } as any);

        const isValid = await api.verifyConfiguration(apiUrl, apiToken);
        assert.strictEqual(isValid, true);
        assert.strictEqual(tempClientGetStub.calledWith('/api/v1/snippets', { params: { limit: 1 } }), true);
    });

    test('verifyConfiguration - failure', async () => {
        const tempClientGetStub = sinon.stub().rejects(new Error('Network error'));
        axiosCreateStub.returns({ get: tempClientGetStub } as any);

        const isValid = await api.verifyConfiguration(apiUrl, apiToken);
        assert.strictEqual(isValid, false);
    });

    test('getSnippets - success', async () => {
        clientGetStub.resolves({
            data: {
                data: [
                    { id: '1', title: 'Test Snippet 1', content: 'console.log("test")', language: 'javascript' }
                ]
            }
        });

        const snippets = await api.getSnippets();
        assert.strictEqual(snippets.length, 1);
        assert.strictEqual(snippets[0].title, 'Test Snippet 1');
    });

    test('searchSnippets - abort signal', async () => {
        // Axios throws a Cancel object when aborted
        const cancelError = new Error('Canceled');
        (cancelError as any).__CANCEL__ = true; // Axios isCancel check uses this
        sinon.stub(axios, 'isCancel').returns(true);
        
        clientGetStub.rejects(cancelError);

        const abortController = new AbortController();
        const searchPromise = api.searchSnippets('test', abortController.signal);
        
        abortController.abort();
        
        const snippets = await searchPromise;
        // Aborted requests should return empty array based on our catch block
        assert.deepStrictEqual(snippets, []);
    });
});

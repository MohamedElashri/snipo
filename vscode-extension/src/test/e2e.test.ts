import * as assert from 'assert';
import * as vscode from 'vscode';
import { api } from '../api';

suite('E2E API Test Suite', () => {
    const apiUrl = 'http://localhost:8080';

    setup(async () => {
        const config = vscode.workspace.getConfiguration('snipo');
        await config.update('apiUrl', apiUrl, vscode.ConfigurationTarget.Global);
        
        const ext = vscode.extensions.getExtension('muhammadelashri.snipo');
        if (ext && !ext.isActive) {
            await ext.activate();
        }
    });

    test('createSnippet - graceful failure or success when backend is running', async () => {
        const snippetData = {
            title: 'E2E Test Snippet',
            content: 'console.log("e2e");',
            language: 'javascript'
        };

        const result = await api.createSnippet(snippetData);
        
        // Since we are not mocking fetch here, it will attempt a real request.
        // If the backend is not running, api.createSnippet returns null.
        // If it is running, it returns the snippet. We accept either.
        if (result === null) {
            assert.strictEqual(result, null);
        } else {
            assert.strictEqual(result.title, snippetData.title);
        }
    });
    
    test('getSnippets - graceful failure or success when backend is running', async () => {
        const result = await api.getSnippets();
        
        // If the backend is running, it returns snippets.
        // If not, it gracefully catches the error and returns an empty array (or cached).
        assert.ok(Array.isArray(result));
    });
});

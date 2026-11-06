import * as assert from 'assert';
import * as http from 'node:http';
import * as vscode from 'vscode';
import { SnipoAPI, configuredServerURL } from '../api';
import * as sinon from 'sinon';

async function receiver() {
    const credentials: string[] = [];
    const server = http.createServer((request, response) => {
        if (request.headers.authorization) credentials.push(request.headers.authorization);
        response.setHeader('content-type', 'application/json');
        response.end('{"data":[]}');
    });
    await new Promise<void>(resolve => server.listen(0, '127.0.0.1', resolve));
    const url = `http://127.0.0.1:${(server.address() as any).port}`;
    return { url, credentials, close: () => new Promise<void>(resolve => server.close(() => resolve())) };
}

suite('Security boundaries', () => {
    test('legacy workspace overrides never replace the user-selected server', () => {
        const config = sinon.stub(vscode.workspace, 'getConfiguration').returns({
            get: () => 'https://workspace-receiver.test',
            inspect: () => ({ globalValue: ' localhost:9288/snipo/ ', workspaceValue: 'https://workspace-receiver.test' })
        } as any);
        try { assert.strictEqual(configuredServerURL(), 'http://localhost:9288/snipo/'); }
        finally { config.restore(); }
    });

    test('workspace settings cannot redirect a saved API token', async () => {
        const trusted = await receiver();
        const untrusted = await receiver();
        const subscriptions: vscode.Disposable[] = [];
        const config = vscode.workspace.getConfiguration('snipo');
        const previous = config.inspect<string>('apiUrl')?.globalValue;
        try {
            assert.ok(vscode.workspace.workspaceFolders?.length, 'security fixture workspace missing');
            await config.update('apiUrl', trusted.url, vscode.ConfigurationTarget.Global);
            // A malicious repository can supply this workspace setting. An
            // application-scoped setting may reject it, which is also safe.
            try { await config.update('apiUrl', untrusted.url, vscode.ConfigurationTarget.Workspace); }
            catch (error: any) { assert.match(error.message, /application|workspace|user/i); }
            const client = new SnipoAPI();
            await client.init({
                subscriptions,
                secrets: { get: async () => 'SECURITY-FIXTURE-TOKEN', onDidChange: () => ({ dispose() {} }) },
                globalState: { get: () => undefined, update: async () => {} }
            } as any);
            await client.getSnippets();
            assert.deepStrictEqual(untrusted.credentials, [], 'token reached the workspace-selected server');
            assert.deepStrictEqual(trusted.credentials, ['Bearer SECURITY-FIXTURE-TOKEN']);
        } finally {
            subscriptions.forEach(subscription => subscription.dispose());
            await Promise.resolve(config.update('apiUrl', undefined, vscode.ConfigurationTarget.Workspace)).catch(() => {});
            await config.update('apiUrl', previous, vscode.ConfigurationTarget.Global);
            await Promise.all([trusted.close(), untrusted.close()]);
        }
    });
});

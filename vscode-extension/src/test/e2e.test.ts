import * as assert from 'assert';
import * as http from 'http';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { api, SnipoAPI } from '../api';

suite('HTTP API integration', () => {
    let server: http.Server;
    let client: SnipoAPI;
    let oldUrl: string | undefined;
    let subscriptions: vscode.Disposable[];
    const snippet = { id: 'integration-1', title: 'Integration snippet', content: 'hello', language: 'plaintext' };

    setup(async () => {
        // Use a private ephemeral server; never send test writes to a user's instance.
        server = http.createServer((req, res) => {
            req.resume();
            req.on('end', () => {
                res.setHeader('Content-Type', 'application/json');
                if (req.url?.includes('limit=50')) {
                    const page=Number(new URL(req.url, 'http://test').searchParams.get('page') || 1);
                    res.end(JSON.stringify({data:page===1?Array.from({length:50},(_,i)=>({...snippet,id:String(i)})):[{...snippet,id:'50'}],pagination:{page,totalPages:2}}));
                } else if (req.url === '/api/v1/snippets' && req.method === 'POST') {
                    res.writeHead(201);
                    res.end(JSON.stringify({ data: snippet }));
                } else {
                    res.end(JSON.stringify({ data: [snippet] }));
                }
            });
        });
        await new Promise<void>(resolve => server.listen(0, '127.0.0.1', resolve));
        const port = (server.address() as import('net').AddressInfo).port;
        const config = vscode.workspace.getConfiguration('snipo');
        oldUrl = config.get<string>('apiUrl');
        await config.update('apiUrl', `http://127.0.0.1:${port}`, vscode.ConfigurationTarget.Global);
        client = new SnipoAPI();
        subscriptions = [];
        await client.init({
            secrets: { get: async () => 'test-token', onDidChange: () => ({ dispose() {} }) },
            globalState: { update: async () => {}, get: () => undefined },
            subscriptions
        } as any);
    });

    teardown(async () => {
        subscriptions.forEach(subscription => subscription.dispose());
        server.closeAllConnections();
        await new Promise<void>((resolve, reject) => server.close(error => error ? reject(error) : resolve()));
        await vscode.workspace.getConfiguration('snipo').update('apiUrl', oldUrl, vscode.ConfigurationTarget.Global);
    });

    test('create returns the snippet inside the server envelope', async () => {
        const result = await client.createSnippet(snippet);
        assert.strictEqual(result?.id, snippet.id);
        assert.strictEqual(result?.content, snippet.content);
    });

    test('list returns snippets from the server envelope', async () => {
        const result = await client.getSnippets();
        assert.strictEqual(result.length, 51);
        assert.strictEqual(result[50].id, '50');
    });
});

suite('Notes in the editor', () => {
    test('opens a native side-by-side preview of an unsaved Markdown note', async function () {
        this.timeout(10000);
        const extension = vscode.extensions.getExtension('muhammadelashri.snipo');
        assert.ok(extension);
        await extension.activate();
        const stubs = sinon.createSandbox();
        stubs.stub(api, 'isConfigured').resolves(true);
        stubs.stub(vscode.window, 'showQuickPick').callsFake(async items => (await items)[0]);
        try {
            await vscode.commands.executeCommand('snipo.newNote');
            const editor = vscode.window.activeTextEditor;
            assert.ok(editor);
            assert.strictEqual(editor.document.languageId, 'markdown');
            assert.ok(await editor.edit(edit => edit.insert(new vscode.Position(0, 0), '# Draft preview')));

            await vscode.commands.executeCommand('snipo.previewNote');
            const tabs = () => vscode.window.tabGroups.all.flatMap(group => group.tabs);
            const hasMarkdownPreview = () => tabs().some(tab => tab.input instanceof vscode.TabInputWebview && tab.input.viewType.includes('markdown.preview'));
            for (let attempt = 0; !hasMarkdownPreview() && attempt < 100; attempt++) {
                await new Promise(resolve => setTimeout(resolve, 50));
            }
            const openTabs = tabs().map(tab => ({ label: tab.label, viewType: (tab.input as vscode.TabInputWebview)?.viewType }));
            assert.ok(hasMarkdownPreview(), `Expected the Markdown preview tab, found ${JSON.stringify(openTabs)}`);
            assert.strictEqual(editor.document.getText(), '# Draft preview');
        } finally {
            stubs.restore();
        }
    });
});

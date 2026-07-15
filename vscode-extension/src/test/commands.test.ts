import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { searchAndInsertSnippet, saveSelectedSnippet } from '../commands';
import { api } from '../api';

suite('Commands Test Suite', () => {
    let showQuickPickStub: sinon.SinonStub;
    let showInputBoxStub: sinon.SinonStub;
    let insertSnippetStub: sinon.SinonStub;
    let isConfiguredStub: sinon.SinonStub;

    setup(() => {
        showQuickPickStub = sinon.stub(vscode.window, 'showQuickPick');
        showInputBoxStub = sinon.stub(vscode.window, 'showInputBox');
        isConfiguredStub = sinon.stub(api, 'isConfigured').resolves(true);
        
        // Mock active editor
        insertSnippetStub = sinon.stub();
        sinon.stub(vscode.window, 'activeTextEditor').get(() => ({
            insertSnippet: insertSnippetStub,
            selection: new vscode.Selection(0, 0, 0, 0),
            document: {
                getText: () => 'const test = true;',
                languageId: 'typescript'
            }
        }));
    });

    teardown(() => {
        sinon.restore();
    });

    test('saveSelectedSnippet - success path', async () => {
        const createSnippetStub = sinon.stub(api, 'createSnippet').resolves({
            id: '1', title: 'Test', description: '', content: 'const test = true;', language: 'typescript', is_favorite: false, is_public: false
        });

        // Mock the input box for title
        showInputBoxStub.onFirstCall().resolves('My Snippet');
        // Mock the input box for description
        showInputBoxStub.onSecondCall().resolves('A test snippet');
        // Mock the quick pick for language
        showQuickPickStub.resolves({ label: 'typescript' });

        await saveSelectedSnippet();

        assert.strictEqual(createSnippetStub.calledOnce, true);
        const args = createSnippetStub.firstCall.args[0];
        assert.strictEqual(args.title, 'My Snippet');
        assert.strictEqual(args.content, 'const test = true;');
        assert.strictEqual(args.language, 'typescript');
    });

    test('saveSelectedSnippet - fails if unconfigured', async () => {
        isConfiguredStub.resolves(false);
        const errorMessageStub = sinon.stub(vscode.window, 'showErrorMessage');
        
        await saveSelectedSnippet();

        assert.strictEqual(errorMessageStub.calledOnce, true);
        assert.strictEqual(showInputBoxStub.called, false);
    });
});

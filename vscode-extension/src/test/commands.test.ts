import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { useTemplate } from '../templates';
import { searchAndInsertSnippet, saveSelectedSnippet, updateSnippet, openInSnipo, openInVSCode, newNote, previewOpenNote, saveOpenNote } from '../commands';
import { api } from '../api';
import { SnipoCodeLensProvider } from '../providers';

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

    test('Save File CodeLens saves its whole document, not the active selection', async () => {
        const create = sinon.stub(api, 'createSnippet').resolves(null);
        const document = await vscode.workspace.openTextDocument({
            content: 'first line\nsecond line\n', language: 'python'
        });
        const cancellation = new vscode.CancellationTokenSource();
        const lenses = await new SnipoCodeLensProvider().provideCodeLenses(document, cancellation.token);
        const command = lenses![0].command!;
        assert.strictEqual(command.command, 'snipo.saveSelected');
        showInputBoxStub.onFirstCall().resolves('Whole file');
        showInputBoxStub.onSecondCall().resolves('');
        showQuickPickStub.callsFake(async items => items[0]);
        await saveSelectedSnippet(command.arguments![0]);
        assert.strictEqual(create.firstCall.args[0].content, document.getText());
        assert.strictEqual(create.firstCall.args[0].language, 'python');
        cancellation.dispose();
    });

    for (const step of ['description', 'language']) {
        test(`saveSelectedSnippet - cancelling ${step} does not save`, async () => {
            const create = sinon.stub(api, 'createSnippet');
            showInputBoxStub.onFirstCall().resolves('Title');
            showInputBoxStub.onSecondCall().resolves(step === 'description' ? undefined : '');
            showQuickPickStub.resolves(step === 'language' ? undefined : { label: 'typescript' });
            await saveSelectedSnippet();
            assert.strictEqual(create.called, false);
        });
    }

    test('updateSnippet - cancelling language does not overwrite snippet', async () => {
        const update = sinon.stub(api, 'updateSnippet');
        let accept: () => void = () => {};
        const picker: any = {
            selectedItems: [{ snippet: { id: 'existing', title: 'Existing' } }],
            onDidChangeValue: () => {}, onDidHide: () => {},
            onDidAccept: (fn: () => void) => { accept = fn; },
            show: () => accept(), hide: () => {}, dispose: () => {}
        };
        sinon.stub(vscode.window, 'createQuickPick').returns(picker);
        sinon.stub(api, 'getSnippets').resolves([]);
        showQuickPickStub.resolves(undefined);
        await updateSnippet();
        assert.strictEqual(update.called, false);
    });

    for (const fileCount of [0, 1, 2]) {
        test(`update preserves metadata and ${fileCount} existing files`, async () => {
            const current: any = {
                id: 'existing', title: 'Latest title', description: 'Keep description',
                content: 'old', language: 'python', is_public: true, is_archived: true,
                expires_at: '2030-01-01T00:00:00Z', folders: [{ id: 7 }],
                files: Array.from({ length: fileCount }, (_, i) => ({
                    id: i + 1, filename: `file${i}.py`, content: `original ${i}`, language: 'python'
                }))
            };
            const get = sinon.stub(api, 'getSnippet').resolves(current);
            const update = sinon.stub(api, 'updateSnippet').resolves(null);
            let accept: () => void = () => {};
            const picker: any = {
                selectedItems: [{ snippet: { id: 'existing', title: 'Stale title' } }],
                onDidChangeValue: () => {}, onDidHide: () => {},
                onDidAccept: (fn: () => void) => { accept = fn; },
                show: () => accept(), hide: () => {}, dispose: () => {}
            };
            sinon.stub(vscode.window, 'createQuickPick').returns(picker);
            sinon.stub(api, 'getSnippets').resolves([]);
            showQuickPickStub.resolves({ label: 'typescript' });
            await updateSnippet();
            assert.strictEqual(get.calledOnceWith('existing'), true);
            const payload = update.firstCall.args[1];
            assert.deepStrictEqual(payload, {
                title: current.title, description: current.description,
                is_public: true, is_archived: true, expires_at: current.expires_at,
                folder_id: 7, content: 'const test = true;', language: 'typescript',
                files: current.files.map((file: any, i: number) => i === 0
                    ? { ...file, content: 'const test = true;', language: 'typescript' } : file)
            });
            // Failed reads must never fall back to overwriting from the picker.
            get.resolves(null);
            update.resetHistory();
            sinon.stub(vscode.window, 'showErrorMessage');
            await updateSnippet();
            assert.strictEqual(update.called, false);
        });
    }

    test('saveSelectedSnippet - fails if unconfigured', async () => {
        isConfiguredStub.resolves(false);
        const errorMessageStub = sinon.stub(vscode.window, 'showErrorMessage');
        
        await saveSelectedSnippet();

        assert.strictEqual(errorMessageStub.calledOnce, true);
        assert.strictEqual(showInputBoxStub.called, false);
    });
    test('search picker ignores stale initial and search responses', async () => {
        const clock = sinon.useFakeTimers();
        let change: (value: string) => void = () => {};
        let hide: () => void = () => {};
        const picker: any = {
            items: [], selectedItems: [], busy: false,
            onDidChangeValue: (fn: typeof change) => { change = fn; },
            onDidAccept: () => {}, onDidHide: (fn: typeof hide) => { hide = fn; },
            show: () => {}, dispose: sinon.stub(), hide: () => hide()
        };
        sinon.stub(vscode.window, 'createQuickPick').returns(picker);
        let resolveInitial!: (items: any[]) => void;
        let resolveOld!: (items: any[]) => void;
        sinon.stub(api, 'getSnippets').returns(new Promise(resolve => { resolveInitial = resolve; }));
        const search = sinon.stub(api, 'searchSnippets');
        search.onFirstCall().returns(new Promise(resolve => { resolveOld = resolve; }));
        search.onSecondCall().resolves([{ id: 'new', title: 'Newest' } as any]);
        const pending = searchAndInsertSnippet();
        await clock.tickAsync(0);
        change('old');
        await clock.tickAsync(300);
        change('new');
        await clock.tickAsync(300);
        assert.strictEqual(picker.items[0].snippet.id, 'new');
        resolveInitial([{ id: 'initial', title: 'Initial' }]);
        resolveOld([{ id: 'old', title: 'Old' }]);
        await clock.tickAsync(0);
        assert.strictEqual(picker.items[0].snippet.id, 'new');
        assert.strictEqual(picker.busy, false);
        hide();
        await pending;
        assert.strictEqual(picker.dispose.calledOnce, true);
    });

    test('new editor languages are offered and LaTeX is detected', async () => {
        sinon.stub(api, 'createSnippet').resolves(null);
        showInputBoxStub.onFirstCall().resolves('TeX');showInputBoxStub.onSecondCall().resolves('');
        await saveSelectedSnippet({languageId:'latex',getText:()=> 'sample'} as any);
        const items=showQuickPickStub.firstCall.args[0] as vscode.QuickPickItem[];
        for(const lang of ['tex','bibtex','cuda']) assert.ok(items.some(i=>i.label===lang),lang+' missing');
        assert.strictEqual(items[0].label,'tex');
    });
    test('open browser accepts the same schemeless URL as the API', async () => {
        sinon.stub(vscode.workspace,'getConfiguration').returns({inspect:()=>({globalValue:'localhost:9288/snipo/'})} as any);
        const open=sinon.stub(vscode.env,'openExternal').resolves(true);
        await openInSnipo('example');
        assert.strictEqual(open.calledOnce,true);
        assert.strictEqual(open.firstCall.args[0].scheme,'http');
        assert.strictEqual(open.firstCall.args[0].authority,'localhost:9288');
        assert.strictEqual(open.firstCall.args[0].path,'/snipo/');
        assert.strictEqual(open.firstCall.args[0].query,'snippet=example');
    });

    test('opening a sidebar snippet fetches its current content', async () => {
        const fresh = {id:'current',title:'Current',content:'updated on server',language:'plaintext'} as any;
        const get = sinon.stub(api,'getSnippet').resolves(fresh);
        const open = sinon.stub(vscode.workspace,'openTextDocument').resolves({} as any);
        sinon.stub(vscode.window,'showTextDocument').resolves({} as any);
        await openInVSCode({...fresh,content:'stale sidebar content'});
        assert.strictEqual(get.calledWith('current'),true);
        assert.strictEqual((open.firstCall.args[0] as any).content,'updated on server');
    });

    test('editing a note sends its current revision and preserves organization', async () => {
        const note = {
            id: 'runbook', kind: 'note' as const, revision: 8, title: 'Runbook', description: '',
            content: '# Runbook', language: 'markdown', is_favorite: false, is_public: false,
            folders: [{ id: 12, name: 'Deployment' }], tags: [{ id: 4, name: 'ops', color: '' }]
        };
        const doc = { uri: vscode.Uri.parse('untitled:Runbook'), getText: () => '# Edited', languageId: 'markdown' };
        sinon.stub(api, 'getSnippet').resolves(note);
        const update = sinon.stub(api, 'updateSnippet').resolves({ ...note, revision: 9 });
        sinon.stub(vscode.workspace, 'openTextDocument').resolves(doc as vscode.TextDocument);
        sinon.stub(vscode.window, 'showTextDocument').resolves({} as vscode.TextEditor);
        sinon.stub(vscode.window, 'activeTextEditor').get(() => ({ document: doc }));

        await openInVSCode(note);
        await saveOpenNote();
        assert.strictEqual(update.firstCall.args[0], note.id);
        assert.deepStrictEqual(update.firstCall.args[1], {
            kind: 'note', revision: 8, title: 'Runbook', description: '', content: '# Edited',
            language: 'markdown', is_public: false, is_archived: undefined, expires_at: undefined,
            folder_id: 12, tags: ['ops']
        });
    });

    test('creates a private Markdown note, previews its draft, and updates the newly saved revision', async () => {
        const doc = { uri: vscode.Uri.parse('untitled:NewNote'), getText: () => '# Release notes\n\nChanges', languageId: 'markdown' };
        const open = sinon.stub(vscode.workspace, 'openTextDocument').resolves(doc as vscode.TextDocument);
        sinon.stub(vscode.window, 'showTextDocument').resolves({} as vscode.TextEditor);
        sinon.stub(vscode.window, 'activeTextEditor').get(() => ({ document: doc }));
        const execute = sinon.stub(vscode.commands, 'executeCommand').resolves(undefined);
        const note = {
            id: 'new-note', kind: 'note' as const, revision: 1, title: 'Release notes', description: '',
            content: doc.getText(), language: 'markdown', is_favorite: false, is_public: false,
        };
        const create = sinon.stub(api, 'createSnippet').resolves(note);
        const update = sinon.stub(api, 'updateSnippet').resolves({ ...note, revision: 2 });
        showQuickPickStub.callsFake(async items => items[0]);

        await newNote();
        assert.deepStrictEqual(open.firstCall.args[0], { language: 'markdown', content: '' });
        await previewOpenNote();
        assert.ok(execute.calledWith('markdown.showPreviewToSide', doc.uri));
        showInputBoxStub.onFirstCall().resolves(undefined);
        await saveOpenNote();
        assert.ok(create.notCalled);
        showInputBoxStub.onSecondCall().resolves('Release notes');
        await saveOpenNote();
        assert.strictEqual(showInputBoxStub.secondCall.args[0].value, 'Release notes');
        assert.ok(create.calledOnceWithExactly({
            kind: 'note', title: 'Release notes', content: doc.getText(), language: 'markdown', is_public: false,
        }));
        await saveOpenNote();
        assert.strictEqual(update.firstCall.args[0], note.id);
        assert.strictEqual(update.firstCall.args[1].revision, 1);
    });

});

suite('Template Commands', () => {
    teardown(() => sinon.restore());

    test('reports a render error for a template without fields', async () => {
        const summary = { id: 'template-1', revision: 1, name: 'Blank note', purpose: '', filenames: ['note.md'], favorite: false };
        sinon.stub(api, 'isConfigured').resolves(true);
        sinon.stub(api, 'listTemplates').resolves([summary]);
        sinon.stub(api, 'getTemplate').resolves({ ...summary, definition: { name: summary.name, purpose: '', fields: [] } });
        sinon.stub(api, 'renderTemplate').rejects(new Error('Revision unavailable'));
        sinon.stub(vscode.window, 'showQuickPick').callsFake(async (items: any) => items[0]);
        const showError = sinon.stub(vscode.window, 'showErrorMessage').resolves();
        await useTemplate();
        assert.strictEqual(showError.callCount, 1);
        assert.deepStrictEqual(showError.firstCall.args, ['Revision unavailable']);
    });

    test('reviews all files and saves the edited secondary file', async () => {
        const origin = { id: 'bundle', revision: 2, name: 'Runbook bundle' };
        const files = [
            { filename: 'runbook.md', content: '# Runbook', language: 'markdown' },
            { filename: 'health.sh', content: 'echo health', language: 'bash' },
            { filename: '.env.example', content: 'PORT=8080', language: 'plaintext' }
        ];
        sinon.stub(api, 'isConfigured').resolves(true);
        sinon.stub(api, 'listTemplates').resolves([{ ...origin, purpose: 'Service files', filenames: files.map(file => file.filename), favorite: false }]);
        sinon.stub(api, 'getTemplate').resolves({ ...origin, definition: { name: origin.name, purpose: 'Service files', fields: [] } });
        const render = sinon.stub(api, 'renderTemplate').resolves({
            title: 'Running Redis', description: '', files, content: files[0].content, language: files[0].language,
            template_origin: origin, is_public: false
        });
        const create = sinon.stub(api, 'createSnippet').resolves({
            id: 'created', title: 'Running Redis', description: '', content: files[0].content, language: files[0].language,
            is_favorite: false, is_public: false
        });
        const choices = sinon.stub(vscode.window, 'showQuickPick');
        choices.onFirstCall().callsFake(async (items: any) => items[0]);
        choices.onSecondCall().callsFake(async (items: any) => items.find((item: any) => item.label === 'health.sh'));
        choices.onThirdCall().callsFake(async (items: any) => items.find((item: any) => item.action === 'save'));
        sinon.stub(vscode.window, 'showInformationMessage').callsFake(async () => {
            const editor = vscode.window.activeTextEditor!;
            assert.ok(editor.document.isUntitled);
            assert.ok(editor.document.uri.path.endsWith('/health.sh'));
            assert.strictEqual(editor.document.getText(), 'echo health');
            await editor.edit(edit => edit.insert(new vscode.Position(0, 0), '# Edited\n'));
            return 'Return to template actions' as any;
        });
        sinon.stub(vscode.window, 'showInputBox').resolves('Running Redis');
        try {
            await useTemplate();
            assert.ok(render.calledWith('bundle', 2));
            const saved = create.firstCall.args[0];
            assert.strictEqual(saved.files?.length, 3);
            assert.strictEqual(saved.files[1].content, '# Edited\necho health');
            assert.deepStrictEqual(saved.template_origin, origin);
        } finally {
            await vscode.commands.executeCommand('workbench.action.revertAndCloseActiveEditor');
        }
    });
});

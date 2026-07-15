import * as assert from 'assert';
import * as sinon from 'sinon';
import { SnippetTreeProvider } from '../treeView';
import { api } from '../api';

suite('TreeView Test Suite', () => {
    let getSnippetsStub: sinon.SinonStub;
    let getRecentSnippetsStub: sinon.SinonStub;

    setup(() => {
        getSnippetsStub = sinon.stub(api, 'getSnippets').resolves([
            { id: '1', title: 'Fav Snippet', content: '', language: 'js', is_favorite: true, is_public: false, description: '' }
        ]);
        getRecentSnippetsStub = sinon.stub(api, 'getRecentSnippets').resolves([
            { id: '2', title: 'Recent Snippet', content: '', language: 'ts', is_favorite: false, is_public: false, description: '' }
        ]);
        sinon.stub(api, 'isConfigured').resolves(true);
    });

    teardown(() => {
        sinon.restore();
    });

    test('TreeProvider fetches favorites', async () => {
        const provider = new SnippetTreeProvider('favorites');
        const elements = await provider.getChildren();
        
        assert.strictEqual(getSnippetsStub.calledOnce, true);
        assert.strictEqual(elements.length, 1);
        assert.strictEqual(elements[0].snippet.title, 'Fav Snippet');
    });

    test('TreeProvider fetches recent', async () => {
        const provider = new SnippetTreeProvider('recent');
        const elements = await provider.getChildren();
        
        assert.strictEqual(getRecentSnippetsStub.calledOnce, true);
        assert.strictEqual(elements.length, 1);
        assert.strictEqual(elements[0].snippet.title, 'Recent Snippet');
    });
});

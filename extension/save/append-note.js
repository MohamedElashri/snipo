function addAppendNotePane(modal, actions, close) {
    function element(tag, text, className) {
        const node = document.createElement(tag);
        if (text) node.textContent = text;
        if (className) node.className = className;
        if (tag === 'button') node.type = 'button';
        return node;
    }

    function request(action, values = {}) {
        return new Promise((resolve, reject) => {
            chrome.runtime.sendMessage({ action, draftId, ...values }, response => {
                const error = chrome.runtime.lastError?.message || response?.error;
                if (!response?.success) reject(new Error(error || 'Could not reach Snipo'));
                else resolve(response.data);
            });
        });
    }

    const createPane = element('div');
    createPane.append(...modal.children);
    const openButton = element('button', 'Append to a note', 'snipo-btn-secondary');
    actions.before(openButton);

    const appendPane = element('form');
    appendPane.hidden = true;
    const heading = element('h2', 'Append to a note');
    const searchGroup = element('label', 'Find a note by title or filename', 'snipo-form-group');
    const search = element('input', '', 'snipo-input');
    search.type = 'search';
    search.autocomplete = 'off';
    searchGroup.append(search);
    const selectGroup = element('label', 'Note or Markdown file', 'snipo-form-group');
    const select = element('select', '', 'snipo-input');
    select.size = 4;
    selectGroup.append(select);

    const status = element('p', '', 'snipo-note-status');
    status.setAttribute('role', 'status');
    const more = element('button', 'Load more', 'snipo-btn-secondary');
    more.hidden = true;
    const buttons = element('div', '', 'snipo-modal-actions');
    const back = element('button', 'Back', 'snipo-btn-secondary');
    const submit = element('button', 'Append selection', 'snipo-btn-primary');
    submit.type = 'submit';
    submit.disabled = true;
    buttons.append(back, submit);
    appendPane.append(heading, searchGroup, selectGroup, status, more, buttons);
    modal.append(createPane, appendPane);

    let requestNumber = 0;
    let page = 1;
    let debounce;
    let saving = false;

    async function loadTargets(nextPage) {
        const current = ++requestNumber;
        select.disabled = true;
        submit.disabled = true;
        more.disabled = true;
        status.textContent = 'Finding notes…';
        try {
            const result = await request('listDraftNoteTargets', { query: search.value.trim(), page: nextPage });
            if (current !== requestNumber) return;
            for (const target of result.targets) {
                const title = target.filename ? target.title + ' — ' + target.filename : target.title;
                const option = element('option', title);
                option.value = target.snippet_id;
                if (target.file_id !== undefined) option.dataset.fileId = target.file_id;
                select.append(option);
            }
            page = nextPage;
            more.textContent = 'Load more';
            more.hidden = !result.has_more;
            status.textContent = select.options.length ? '' : 'No matching private Markdown notes.';
        } catch (error) {
            if (current !== requestNumber) return;
            status.textContent = error.message;
            more.textContent = 'Retry';
            more.hidden = false;
        } finally {
            if (current === requestNumber) {
                more.disabled = false;
                select.disabled = !select.options.length;
                submit.disabled = !select.value;
            }
        }
    }

    function resetSearch() {
        ++requestNumber;
        clearTimeout(debounce);
        page = 0;
        select.replaceChildren();
        select.disabled = true;
        submit.disabled = true;
        more.hidden = true;
        status.textContent = 'Finding notes…';
    }

    openButton.addEventListener('click', () => {
        createPane.hidden = true;
        appendPane.hidden = false;
        showToast(true, '');
        search.value = '';
        resetSearch();
        search.focus();
        void loadTargets(1);
    });
    back.addEventListener('click', () => {
        resetSearch();
        appendPane.hidden = true;
        createPane.hidden = false;
        openButton.focus();
    });
    search.addEventListener('input', () => {
        resetSearch();
        debounce = setTimeout(() => loadTargets(1), 250);
    });
    select.addEventListener('change', () => { submit.disabled = !select.value; });
    more.addEventListener('click', () => loadTargets(page + 1));
    appendPane.addEventListener('submit', async event => {
        event.preventDefault();
        const target = select.selectedOptions[0];
        if (saving || submit.disabled || !target) return;
        saving = true;
        search.disabled = select.disabled = more.disabled = back.disabled = submit.disabled = true;
        status.textContent = 'Appending selection…';
        try {
            await request('appendDraftToNote', {
                snippetId: target.value,
                fileId: target.dataset.fileId === undefined ? undefined : Number(target.dataset.fileId),
            });
            close();
        } catch (error) {
            status.textContent = error.message;
            saving = false;
            search.disabled = select.disabled = more.disabled = back.disabled = submit.disabled = false;
        }
    });
}

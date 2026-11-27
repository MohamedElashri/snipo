import * as vscode from 'vscode';
import { randomUUID } from 'node:crypto';
import type { TemplateDraft, TemplateField } from '@snipo/api-client';
import { api } from './api';

type FieldValues = Record<string, string>;

function templateError(error: unknown): string {
    const problem = error as { message?: string; response?: { data?: { error?: { message?: string; details?: { message: string }[] } } } };
    const data = problem.response?.data?.error;
    return data?.details?.map(detail => detail.message).join(' · ') || data?.message || problem.message || 'Template request failed.';
}

async function enterField(field: TemplateField, value: string): Promise<string | undefined> {
    const prompt = field.help || field.label;
    if (field.type === 'choice' || field.type === 'toggle') {
        const choices = field.type === 'toggle' ? ['true', 'false'] : field.choices ?? [];
        const items = choices.map(choice => ({
            label: choice, value: choice, description: choice === value ? 'Current value' : undefined
        }));
        if (!field.required) items.unshift({ label: 'Leave blank', value: '', description: undefined });
        return (await vscode.window.showQuickPick(items, { title: field.label, placeHolder: prompt }))?.value;
    }
    if (field.type === 'multiline') {
        const method = await vscode.window.showQuickPick([
            { label: 'Use clipboard text', description: 'Preserves line breaks', value: 'clipboard' },
            { label: 'Type a line', value: 'type' },
            { label: value ? 'Keep current value' : 'Leave blank', value: 'current' }
        ], { title: field.label, placeHolder: prompt });
        if (method === undefined) return undefined;
        if (method.value === 'clipboard') return vscode.env.clipboard.readText();
        if (method.value === 'current') return value;
    }
    return vscode.window.showInputBox({
        title: field.label,
        prompt: field.type === 'date' ? `${prompt} (YYYY-MM-DD)` : prompt,
        value,
        validateInput: text => {
            if (field.required && !text.trim()) return `Enter ${field.label}`;
            if (text && field.type === 'number' && !Number.isFinite(Number(text))) return 'Enter a number';
            if (text && field.type === 'date' && !/^\d{4}-\d{2}-\d{2}$/.test(text)) return 'Enter a date in YYYY-MM-DD format';
            return undefined;
        }
    });
}

async function useDraft(draft: TemplateDraft, target: vscode.TextEditor | undefined) {
    const batch = randomUUID();
    const documents = new Map<number, vscode.TextDocument>();

    async function openFile(index: number) {
        let document = documents.get(index);
        if (!document || document.isClosed) {
            const file = draft.files[index];
            const uri = vscode.Uri.from({ scheme: 'untitled', path: `Snipo-${batch}/${file.filename}` });
            document = await vscode.workspace.openTextDocument(uri);
            const edit = new vscode.WorkspaceEdit();
            edit.insert(uri, new vscode.Position(0, 0), file.content);
            if (file.content && !await vscode.workspace.applyEdit(edit)) throw new Error(`Could not open ${file.filename}`);
            const language = ['bash', 'shell'].includes(file.language) ? 'shellscript' : file.language;
            if (vscode.languages.getLanguages && (await vscode.languages.getLanguages()).includes(language)) {
                document = await vscode.languages.setTextDocumentLanguage(document, language);
            }
            documents.set(index, document);
        }
        return vscode.window.showTextDocument(document, { preview: false });
    }

    const items: (vscode.QuickPickItem & { action: string; index?: number })[] = [
        { label: 'Open all files', action: 'open' },
        { label: 'Save snippet to Snipo', action: 'save', description: draft.title },
        ...draft.files.map((file, index) => ({
            label: file.filename, description: file.language, detail: 'Review or edit this file', action: 'review', index
        }))
    ];
    if (target) items.unshift({ label: 'Insert a file into the active editor', action: 'insert' });

    for (;;) {
        const action = await vscode.window.showQuickPick(items, { title: draft.title, placeHolder: 'Review files or choose what to do with the rendered draft' });
        if (!action) return;
        if (action.action === 'review') {
            await openFile(action.index!);
            const next = await vscode.window.showInformationMessage('The rendered file is open for review.', 'Return to template actions');
            if (!next) return;
            continue;
        }
        if (action.action === 'open') {
            for (const [index] of draft.files.entries()) await openFile(index);
        } else if (action.action === 'insert') {
            const chosen = await vscode.window.showQuickPick(draft.files.map((file, index) => ({ label: file.filename, index })), {
                placeHolder: 'Choose a file to insert'
            });
            if (!chosen || !target) return;
            const content = documents.get(chosen.index)?.getText() ?? draft.files[chosen.index].content;
            const editor = await vscode.window.showTextDocument(target.document, target.viewColumn);
            await editor.insertSnippet(new vscode.SnippetString().appendText(content));
        } else {
            const title = await vscode.window.showInputBox({
                title: 'Save snippet to Snipo', prompt: 'Snippet title', value: draft.title,
                validateInput: value => value.trim() ? undefined : 'Enter a title'
            });
            if (title === undefined) continue;
            const files = draft.files.map((file, index) => ({
                ...file, content: documents.get(index)?.getText() ?? file.content
            }));
            const saved = await api.createSnippet({ ...draft, title, files, content: '' });
            if (!saved) continue;
        }
        return;
    }
}

export async function useTemplate() {
    if (!await api.isConfigured()) {
        vscode.window.showErrorMessage('Configure Snipo before using a template.');
        return;
    }
    const target = vscode.window.activeTextEditor;
    try {
        const templates = await api.listTemplates();
        if (templates === null) {
            vscode.window.showInformationMessage('This Snipo server does not support templates yet.');
            return;
        }
        if (templates.length === 0) {
            vscode.window.showInformationMessage('Create a template in the Snipo web library to get started.');
            return;
        }
        const selected = await vscode.window.showQuickPick(templates.map(item => ({
            label: item.favorite ? `$(star-full) ${item.name}` : item.name,
            description: item.purpose, detail: item.filenames.join(', '), item
        })), { title: 'Use Snipo Template', matchOnDescription: true, matchOnDetail: true });
        if (!selected) return;
        const template = await api.getTemplate(selected.item.id, selected.item.revision);
        const fields = template.definition.fields ?? [];
        const presets = template.definition.presets ?? [];
        let preset: FieldValues = {};
        if (presets.length) {
            const selection = await vscode.window.showQuickPick([
                { label: 'Template defaults', values: {} },
                ...presets.map(item => ({ label: item.name, values: item.values }))
            ], { title: 'Template preset' });
            if (!selection) return;
            preset = selection.values ?? {};
        }
        const values: FieldValues = Object.create(null);
        for (;;) {
            for (const field of fields) {
                const value = await enterField(field, values[field.key] ?? (Object.prototype.hasOwnProperty.call(preset, field.key) ? preset[field.key] : field.default ?? ''));
                if (value === undefined) return;
                values[field.key] = value;
            }
            let draft: TemplateDraft;
            try {
                draft = await api.renderTemplate(template.id, template.revision, values);
            } catch (error: unknown) {
                if (!fields.length) throw error;
                if (await vscode.window.showErrorMessage(templateError(error), 'Edit values') !== 'Edit values') return;
                continue;
            }
            await useDraft(draft, target);
            return;
        }
    } catch (error: unknown) {
        vscode.window.showErrorMessage(templateError(error));
    }
}

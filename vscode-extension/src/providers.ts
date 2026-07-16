import * as vscode from 'vscode';
import { api } from './api';

export class SnipoHoverProvider implements vscode.HoverProvider {
    async provideHover(document: vscode.TextDocument, position: vscode.Position, token: vscode.CancellationToken): Promise<vscode.Hover | null> {
        // Look for snipo:[id] in the hovered text
        const range = document.getWordRangeAtPosition(position, /snipo:[a-zA-Z0-9_-]+/);
        if (!range) {
            return null;
        }

        const word = document.getText(range);
        const snippetId = word.split(':')[1];

        if (!snippetId) {
            return null;
        }

        const snippet = await api.getSnippet(snippetId);
        if (!snippet) {
            return new vscode.Hover(new vscode.MarkdownString(`*Snipo snippet ${snippetId} not found or unavailable offline.*`));
        }

        const markdown = new vscode.MarkdownString();
        markdown.appendMarkdown(`**Snipo Snippet:** ${snippet.title}\n\n`);
        if (snippet.description) {
            markdown.appendMarkdown(`${snippet.description}\n\n`);
        }
        markdown.appendCodeblock(snippet.content, snippet.language || 'plaintext');
        
        return new vscode.Hover(markdown);
    }
}

export class SnipoCodeLensProvider implements vscode.CodeLensProvider {
    private onDidChangeCodeLensesEmitter: vscode.EventEmitter<void> = new vscode.EventEmitter<void>();
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this.onDidChangeCodeLensesEmitter.event;

    // Regexes for common function/class declarations based on supported languages
    private functionRegexes = [
        /^(?:export\s+)?(?:async\s+)?function\s+[a-zA-Z0-9_]+\s*\(/, // JS/TS function
        /^(?:export\s+)?class\s+[a-zA-Z0-9_]+/, // JS/TS/Java/C#/PHP/Ruby/Python class
        /^(?:public|private|protected)?\s*(?:static\s+)?(?:class|interface|enum)\s+[a-zA-Z0-9_]+/, // Java/C# class
        /^(?:public|private|protected)?\s*(?:static\s+)?[a-zA-Z0-9_<>]+\s+[a-zA-Z0-9_]+\s*\([^)]*\)\s*\{?/, // Java/C# method
        /^def\s+[a-zA-Z0-9_]+\s*\(/, // Python/Ruby function
        /^func\s+(?:\([^)]+\)\s+)?[a-zA-Z0-9_]+\s*\(/, // Go function/method
        /^fn\s+[a-zA-Z0-9_]+\s*\(/, // Rust function
        /^(?:pub\s+)?(?:struct|enum|trait)\s+[a-zA-Z0-9_]+/, // Rust struct/enum
        /^struct\s+[a-zA-Z0-9_]+/, // C/C++/Go/Swift struct
    ];

    provideCodeLenses(document: vscode.TextDocument, token: vscode.CancellationToken): vscode.ProviderResult<vscode.CodeLens[]> {
        const lenses: vscode.CodeLens[] = [];
        
        // Add a global "Save file to Snipo" at the very top of the file
        const topRange = new vscode.Range(0, 0, 0, 0);
        lenses.push(new vscode.CodeLens(topRange, {
            title: "Save File to Snipo",
            command: "snipo.saveSelected", // This will capture the whole file if nothing is selected or if we handle it in the command
            arguments: [document]
        }));

        // Limit the number of lines we scan to avoid performance issues on huge files
        const maxLines = Math.min(document.lineCount, 10000);

        for (let i = 0; i < maxLines; i++) {
            const line = document.lineAt(i);
            const text = line.text.trim();

            if (text.length === 0) continue;

            for (const regex of this.functionRegexes) {
                if (regex.test(text)) {
                    const range = new vscode.Range(i, 0, i, line.text.length);
                    const command: vscode.Command = {
                        title: "Save to Snipo",
                        command: "snipo.saveSelected",
                    };
                    lenses.push(new vscode.CodeLens(range, command));
                    break; // Only add one lens per line
                }
            }
        }

        return lenses;
    }
}

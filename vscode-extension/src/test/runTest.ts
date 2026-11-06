import * as path from 'path';
import * as fs from 'node:fs/promises';
import * as os from 'node:os';
import { runTests } from '@vscode/test-electron';

async function main() {
    const fixtureRoot = await fs.mkdtemp(path.join(os.tmpdir(), 'snipo-vscode-security-'));
    const workspace = path.join(fixtureRoot, 'workspace');
    await fs.mkdir(workspace);
    try {
        // The folder containing the Extension Manifest package.json
        // Passed to `--extensionDevelopmentPath`
        const extensionDevelopmentPath = path.resolve(__dirname, '../../');

        // The path to test runner
        // Passed to --extensionTestsPath
        const extensionTestsPath = path.resolve(__dirname, './index');

        // Download VS Code, unzip it and run the integration test
        await runTests({ 
            extensionDevelopmentPath, 
            extensionTestsPath,
            launchArgs: [
                workspace,
                '--user-data-dir=' + path.join(fixtureRoot, 'profile'),
                '--extensions-dir=' + path.join(fixtureRoot, 'extensions'),
                '--skip-welcome',
                '--skip-release-notes',
                '--disable-workspace-trust',
                '--disable-gpu',
                '--disable-software-rasterizer',
                '--no-sandbox'
            ]
        });
    } catch (err) {
        console.error('Failed to run tests', err);
        process.exitCode = 1;
    } finally {
        await fs.rm(fixtureRoot, { recursive: true, force: true });
    }
}

main();

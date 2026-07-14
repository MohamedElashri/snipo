"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.SnipoStatusBar = void 0;
const vscode = __importStar(require("vscode"));
const api_1 = require("./api");
class SnipoStatusBar {
    statusBarItem;
    constructor(context) {
        this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
        this.statusBarItem.command = 'snipo.searchAndInsert';
        this.update();
        // Listen for configuration changes to update the status bar
        vscode.workspace.onDidChangeConfiguration(e => {
            if (e.affectsConfiguration('snipo.apiUrl')) {
                this.update();
            }
        });
        context.secrets.onDidChange(e => {
            if (e.key === 'snipo.apiToken') {
                this.update();
            }
        });
    }
    async update() {
        if (await api_1.api.isConfigured()) {
            this.statusBarItem.text = '$(repo) Snipo';
            this.statusBarItem.tooltip = 'Snipo: Search and Insert Snippet';
            this.statusBarItem.backgroundColor = undefined;
        }
        else {
            this.statusBarItem.text = '$(warning) Snipo';
            this.statusBarItem.tooltip = 'Snipo is not configured. Click to search (will prompt setup).';
            this.statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.warningBackground');
        }
        this.statusBarItem.show();
    }
    dispose() {
        this.statusBarItem.dispose();
    }
}
exports.SnipoStatusBar = SnipoStatusBar;
//# sourceMappingURL=statusBar.js.map
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.deactivate = exports.activate = void 0;
// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
const vscode = require("vscode");
const child_process_1 = require("child_process");
async function openInUntitled(content, language) {
    const document = await vscode.workspace.openTextDocument({
        language,
        content,
    });
    vscode.window.showTextDocument(document);
}
// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
function activate(context) {
    // Use the console to output diagnostic information (console.log) and errors (console.error)
    // This line of code will only be executed once when your extension is activated
    console.log('Congratulations, your extension "semantic-diff-tool" is now active!');
    // The command has been defined in the package.json file
    // Now provide the implementation of the command with registerCommand
    // The commandId parameter must match the command field in package.json
    let disposable = vscode.commands.registerCommand('semantic-diff-tool.sdt', () => {
        // The code you place here will be executed every time your command is executed
        // Display a message box to the user
        (0, child_process_1.exec)('sdt semantic -m -d', {}, (err, stdout, stderr) => {
            if (err) {
                vscode.window.showInformationMessage("Unable to run `sdt` (is it installed?)");
            }
            else {
                openInUntitled(`${Date()}\n` +
                    `---------------------------------------------------------\n` +
                    `${stderr.replace(/\u001b\[.*?m/g, '')}` +
                    `---------------------------------------------------------\n` +
                    `${stdout}`);
            }
        });
    });
    context.subscriptions.push(disposable);
}
exports.activate = activate;
// This method is called when your extension is deactivated
function deactivate() { }
exports.deactivate = deactivate;
//# sourceMappingURL=extension.js.map
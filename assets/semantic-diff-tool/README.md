# semantic-diff-tool README

The extension "semantic-diff-tool" wraps the [external tool
`sdt`](https://www.sdt.dev/). The command-line tool `sdt` compares source
files to identify which changes create semantic differences in the program
operation, and specifically to exclude many changes which cannot be
functionally important to the operation of a program or library.

## Features

The purpose of this extension is simply to provide an call to an external
utility via a VS Code command, and display the resulting analysis within VS
Code.

\!\[feature X\]\(images/feature-x.png\)

## Requirements

See the main documentation for the Semantic Diff Tool for instructions on
how to install it, and make it available in your development environment.
As an extension, its operation is exposed purely as a call to a shell
command.

### 0.0.1

Alpha release (mostly learning how to write VS Code extension)


---

## Following extension guidelines

Ensure that you've read through the extensions guidelines and follow the
best practices for creating your extension.

* [Extension Guidelines](https://code.visualstudio.com/api/references/extension-guidelines)



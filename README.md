<h1 align="center"> gh-hook </h1>

<p align="center">
🪝 A GitHub CLI extension to easily manage your repository webhooks.
<br/>
</p>

## ✨ Features
- Create a repository webhook
- Delete one or more repository webhooks
- List all repository webhooks

## 📼 Demo

![gh-hook demo](vhs-tapes/demo.gif)

## 📦️ Installation

1. Install the `gh` CLI (requires v2.0.0 at a minimum).
2. Install this extension:
   ```sh
   gh extension install lucasmelin/gh-hook
   ```

## 🧑‍💻 Usage

Run using `gh hook`. Run `gh hook --help` for more info.

## Development

```sh
# Install the action locally
gh extension install .; gh hook
# View changes
go build && gh hook
```

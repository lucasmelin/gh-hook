<!--suppress HtmlDeprecatedAttribute -->
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

### Creating a webhook via a JSON file

By default, this extension will prompt for all the information needed to create a webhook when run with `gh hook create`. However, the `--file` flag allows for passing the webhook data via a JSON file instead, if you prefer:

```sh
$ cat hook.json
{
  "active": true,
  "events": [
    "push",
    "pull_request"
  ],
  "config": {
    "url": "https://example.com",
    "content_type": "json",
    "insecure_ssl": "0",
    "secret": "somesecretpassphrase"
  }
}

$ gh hook create --file hook.json
Creating new webhook for gh-hook
Successfully created hook 🪝

$ gh hook list
✓ 404339664 - https://example.com (pull_request, push)
```

## Development

```sh
# Install the action locally
gh extension install .; gh hook
# View changes
go build && gh hook
```

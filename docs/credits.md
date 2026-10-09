# Credits and scope

## Upstream

The memory server this workflow is built around is **not** vendored here and is **not** mine. It is an
MIT-licensed open-source project by its authors; the licence and the copyright belong to them, and this
repository contains only the parts written here: the redactor, the rules, the templates and the notes.

If you are looking for the server, use the upstream project and its own documentation. This repository
assumes you have one running and describes how a session behaves around it.

## What is deliberately not here

- **No credentials, and no configuration copied from a working machine.** Every value in this repository is
  a placeholder: `<memory-server-url>`, `<token-from-your-environment>`. The environment file is expected
  to live outside the repository and to be ignored by git.
- **No dotfiles.** A working setup on the machine this was written on had a token hardcoded in an editor
  extension and in an MCP configuration file. That is exactly the shape of leak `redact` exists to catch, so
  neither file is reproduced here — and if that setup is ever published, the token has to be rotated first,
  because it is in the history the moment it is committed.
- **No hostnames, no internal URLs, no customer data.** If a note needed one of those, it was rewritten
  without it or left out.

## Why this file exists

Because "open source" is a claim about somebody's licence and somebody's work, and a repository that borrows
both without saying so is a repository that cannot be trusted with the rest of its claims either.

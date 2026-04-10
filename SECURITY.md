# Security

## Supported versions

Security fixes land on the default branch (this repository uses **`main`**). There is no separate long-term support (LTS) line yet; if that changes, this file will list supported release lines and their end dates.

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security-sensitive reports.

Instead, use one of the following:

1. **GitHub private vulnerability reporting** — If enabled for this repository, use **Security → Report a vulnerability** on the GitHub project page.
2. **Maintainer contact** — If private reporting is not available, contact the repository maintainers through a channel they publish in the README or org profile (for example, a security email).

Include:

- A short description of the issue and its impact
- Steps to reproduce or a proof of concept (if safe to share)
- Affected versions or commits, if known

We aim to acknowledge receipt within a few business days and to coordinate disclosure after a fix is available.

## Scope

This project is an agent runtime and tooling. Reports about third-party services (cloud LLM APIs, external MCP servers, etc.) should go to those vendors unless the issue is clearly in **this** codebase’s handling of credentials, sessions, or network I/O.

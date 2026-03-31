# Security Policy

## Supported Scope

Security reports are welcome for the publicly committed contents of this repository, including:

- application source code
- build and release scripts
- public documentation
- accidentally committed secrets or sensitive files

Local-only files ignored by Git, such as personal databases or private override files, are out of scope for public triage but should still be reported privately if you believe they were exposed by mistake.

## Reporting a Vulnerability

Please do not open a public issue for:

- secrets
- tokens
- cookies
- personal paths
- exploit details that can be abused before a fix is available

Instead, contact the maintainer through a private channel associated with the repository hosting platform, or create a private security advisory if the platform supports it.

When reporting, include:

- affected file or feature
- reproduction steps
- impact
- whether sensitive data may already be exposed

## Response Expectations

- Acknowledge receipt as soon as practical
- Confirm whether the report is reproducible
- Fix or mitigate exposed secrets immediately if needed
- Publish a public summary after remediation when safe

# Security Policy

## Supported versions

Only the latest minor release gets security fixes.

| Version | Supported |
| ------- | --------- |
| 0.5.x   | Yes       |
| < 0.5   | No        |

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Report vulnerabilities privately through [GitHub Security Advisories](https://github.com/lechefran/webtest/security/advisories/new). Please include:

- a description of the issue and its impact
- steps or a minimal program that reproduces it
- the affected version or versions

You can expect an acknowledgement within 7 days. Once a fix is ready, it will be released and an advisory published, with credit to you unless you prefer to stay anonymous.

## Security model

`webtest` is a testing and benchmarking client. Some inputs are treated as trusted configuration by design:

- **Request URLs** are unrestricted unless `WebClientOptions.AllowedOrigins` is set. Set it whenever URLs can be influenced by untrusted input.
- **`WebClientOptions.FilePath`** and the path passed to **`ReadCsv`** are trusted local paths. They must not come from untrusted input.

Reports showing that these documented trust boundaries can be bypassed are in scope. Reports that rely on passing untrusted input into a trusted parameter are generally not.

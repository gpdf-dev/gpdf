# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do NOT open a public Issue**
2. Use [GitHub Security Advisories](https://github.com/gpdf-dev/gpdf/security/advisories/new) to report privately

We will acknowledge receipt within 48 hours and provide a timeline for a fix.

## Scope

gpdf is a PDF generation library. Security concerns include:

- Malicious font files causing crashes or unexpected behavior
- Malicious image files causing crashes or unexpected behavior
- PDF output that could exploit viewer vulnerabilities
- Denial of service through crafted input (e.g., deeply nested structures)

## Disclosure Policy

- Vulnerabilities will be fixed before public disclosure
- Credit will be given to reporters (unless anonymity is requested)
- A security advisory will be published on GitHub after the fix is released

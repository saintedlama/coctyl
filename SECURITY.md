# Security Policy

## Supported Versions

Coctyl is actively developed. We provide security patches and bug fixes for the current release and master branch.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

The Coctyl team takes the security of our software seriously. If you believe you have found a security vulnerability in Coctyl, please report it to us as described below.

**Please do not report security vulnerabilities through public GitHub issues, discussions, or pull requests.**

### How to Report

Please send an email directly to [christoph.walcher@gmail.com](mailto:christoph.walcher@gmail.com) with the subject line `[SECURITY] Coctyl Vulnerability Report` or use the private **GitHub Security Advisory** reporting feature on GitHub.

Please include as much of the following information as possible:
- Type of vulnerability (e.g., denial of service via malformed AST, resource exhaustion).
- Detailed steps to reproduce the issue (including sample `.go` source code causing unexpected behavior).
- The version or commit hash of Coctyl you tested on.
- Any potential impact on systems executing Coctyl.
- Whether you have identified a mitigation or fix.

### Response Timeline

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours.
- **Assessment**: We will evaluate the impact and confirm the vulnerability within 5 business days.
- **Resolution**: We will work on a fix in private and keep you informed of our progress.
- **Disclosure**: Once a patch is released, we will publicly credit you for the discovery (unless you prefer to remain anonymous).
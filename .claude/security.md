---
version: 1.0
status: draft
last-updated: 2026-04-21
owner: "Thiago"
---

# Security

> Covers security requirements, threat model, compliance requirements, and controls.
> Review and update at each sprint boundary or when architecture changes. Bump `version` in frontmatter.

## Security Requirements

| Requirement | Priority | Status | Owner |
|---|---|---|---|
| Must run trivy scan after each code change | High | Pending | Claude |
| Must run docker build on the Dockerfile, then run trivy scan img on that image after each code change | High | Pending | Claude |

## Threat Model

### Protected Assets

| Asset | Sensitivity | Data Classification |
|---|---|---|
| User credentials | Critical | Restricted |
| Client data / PII | High | Confidential |
| API keys / secrets | Critical | Restricted |

### Threats (STRIDE)

| Threat | Category | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| Threat 1 | Spoofing | Medium | High | [Control] |
| Threat 2 | Information Disclosure | Low | High | [Control] |
| Threat 3 | SQL injection, and other types of injections | Low | High | [Control] |

## Compliance Requirements

- [ ] GDPR data residency, retention, subject rights
- [ ] information security management

## Security Controls

| Control | Implementation | Status |
|---|---|---|
| Authentication | Identity Provider with MFA provider defined at inception | Pending |
| Data encryption at rest | AES-256 via cloud provider storage encryption | Pending |
| Data encryption in transit | TLS 1.3 minimum | Pending |
| Secret management | cloud secret vault no secrets in code; specific tooling defined at inception | Pending |
| Dependency scanning | Dependabot | Pending |
| SAST | SonarQube | Pending |

## OWASP Top 10 Coverage

| Risk | Mitigation in Place |
|---|---|
| A01 Broken Access Control | [Control or Pending] |
| A02 Cryptographic Failures | [Control or Pending] |
| A03 Injection | [Control or Pending] |
| A04 Insecure Design | [Control or Pending] |
| A05 Security Misconfiguration | [Control or Pending] |
| A06 Vulnerable Components | [Control or Pending] |
| A07 Auth and Session Failures | [Control or Pending] |
| A08 Software and Data Integrity Failures | [Control or Pending] |
| A09 Logging and Monitoring Failures | [Control or Pending] |
| A10 Server-Side Request Forgery | [Control or Pending] |

## Security Review History

| Date | Reviewer | Findings | Resolution |
|---|---|---|---|
| 2026-04-26 | Thiago | Initial baseline review | Open |

## Change Log

| Version | Date | Author | Summary |
|---|---|---|---|
| 1.0 | 2026-04-26 | Thiago | Initial security baseline from inception |
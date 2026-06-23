#### OWASP Top 10 Security Audit
- **Broken Access Control**: Are authorization checks performed on every endpoint? Can users access resources they shouldn't? Check for IDOR patterns, missing role checks, direct object references.
- **Cryptographic Failures**: Are sensitive fields encrypted at rest and in transit? Are weak algorithms (MD5, SHA-1, DES, RC4) used? Are hardcoded encryption keys present?
- **Injection**: SQL/NoSQL/LDAP/OS command injection vectors — user input concatenated into queries, shell commands, or eval() calls without parameterization
- **Insecure Design**: Missing rate limiting, missing input validation at architecture level, trust boundary violations
- **Security Misconfiguration**: Default credentials, debug mode enabled in production, overly permissive CORS, verbose error messages with stack traces
- **Vulnerable Components**: Outdated dependencies with known CVEs, deprecated libraries, unmaintained packages
- **Authentication Failures**: Weak password policies, missing MFA, session fixation, improper JWT validation (missing signature verification, no expiration checks, alg=none attacks)
- **Data Integrity Failures**: Unsigned software updates, missing CSP headers, unsafe deserialization
- **Logging & Monitoring**: Missing security-relevant logging, logging sensitive data (PII, secrets), insufficient alerting on suspicious activity
- **SSRF**: User-controlled URLs fetched by the server without allowlist validation

#### Dependency Security
- Are package-lock.json, yarn.lock, Gemfile.lock, go.sum, Cargo.lock, or requirements.txt pinned to specific versions?
- Are there dependencies with known security advisories?
- Are private registry URLs or auth tokens exposed in package configuration?
- Is the dependency supply chain protected (sigstore, SLSA, or similar)?

#### Infrastructure Security
- Dockerfiles: running as root, using `latest` tag, exposing unnecessary ports, shell injection in RUN commands
- CI/CD pipelines: secrets exposed in logs, untrusted checkout of PRs, missing approval gates
- Kubernetes manifests: containers running as root, privileged mode, hostPath mounts, secrets in environment variables
- Terraform/CloudFormation: publicly accessible S3 buckets, overly permissive IAM policies, security groups allowing 0.0.0.0/0

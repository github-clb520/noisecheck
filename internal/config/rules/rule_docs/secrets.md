#### Hardcoded Secrets Detection 🔑
- API keys, access tokens, or secret keys hardcoded in source code (e.g., `sk-*`, `ghp_*`, `AKIA*`, `-----BEGIN.*PRIVATE KEY-----`)
- Database connection strings containing hardcoded passwords (`postgres://user:password@host`, `mysql://user:password@`)
- OAuth tokens, JWT secrets, session secrets, or encryption keys in configuration files
- AWS/GCP/Azure credential files or service account JSON keys committed to the repository
- Slack webhooks, Discord bot tokens, GitHub tokens, or other service integration tokens
- `.env` files or environment configuration files containing production secrets being committed
- Hardcoded encryption keys or IV/nonce values
- Debug backdoors, hardcoded admin accounts, or hardcoded bypass credentials
- Internal URLs, API endpoints, or IP addresses with embedded authentication

#### Secret Management Best Practices
- Use environment variables or a secrets manager (Vault, AWS Secrets Manager, etc.) instead of hardcoded values
- Config files containing secrets should have placeholder values in version control with real values injected at runtime
- Secrets in logs, error messages, or stack traces are a finding
- Expired or rotated credentials left in code should be flagged as stale

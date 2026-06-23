#### Correctness
- Is the logic correct? Are there missing boundary conditions?
- Are exceptions handled properly?
- Is it thread-safe in concurrent scenarios?
- Are there off-by-one errors, integer overflows, or null pointer dereferences?
- Are API responses validated before use (status codes, null checks)?

#### Security 🔒
- Hardcoded secrets, API keys, tokens, passwords, or credentials in source code
- SQL injection, NoSQL injection, or LDAP injection vulnerabilities
- Cross-site scripting (XSS), Cross-site request forgery (CSRF)
- Insecure direct object references (IDOR) — are authorization checks performed?
- Server-side request forgery (SSRF) risks
- Path traversal — user input used in file system operations without sanitization
- Command injection — user input passed to shell exec or system commands
- Insecure deserialization of user-controlled data
- Missing or improper authentication/authorization checks on API endpoints
- Exposed internal infrastructure details in error messages or logs
- Use of known-vulnerable dependencies or deprecated packages with CVEs
- Improper certificate validation or disabled TLS verification
- Overly permissive CORS configuration (`Access-Control-Allow-Origin: *`)
- Sensitive data in URLs, query parameters, or log output

#### Performance
- N+1 queries, unnecessary database round-trips, or missing database indexes
- Unbounded loops or list iterations over user-supplied data
- Memory leaks — goroutines/channels not closed, event listeners not unregistered
- Large object allocations in hot paths
- Redundant computations that could be cached or memoized
- Blocking calls in async/event-loop contexts
- Unoptimized asset bundling or large dependency trees

#### Maintainability
- Is the code clear and easy to understand?
- Do names accurately express intent?
- Does it follow the project's existing code style and architecture patterns?
- Deeply nested conditionals that could be flattened or extracted
- Functions or methods exceeding reasonable complexity (too many parameters, too many responsibilities)
- Duplicated logic that should be extracted to shared utilities
- Missing or misleading comments on non-obvious business logic
- Overly complex type hierarchies or unnecessary abstractions (YAGNI violations)

#### Test Coverage
- Do critical logic paths have corresponding test cases?
- Do test cases cover boundary conditions, error paths, and edge cases?
- Are there tests for security-critical paths (authentication, authorization, input validation)?
- Are mocks/stubs used appropriately — not over-mocked, not missing?
- Are test assertions specific enough (not just `assert true`)?

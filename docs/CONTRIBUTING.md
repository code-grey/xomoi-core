# Contributing to Xomoi-Core

Thank you for your interest in contributing to Xomoi-Core. This project is dedicated to digital sovereignty and follows a strict architectural standard to maintain its "Sovereign Monolith" vision.

## Technical Standards

1. Zero-Dependency Mandate: Avoid third-party libraries unless absolutely necessary. For routing, use the Go 1.26 standard library net/http ServeMux.
2. Memory Efficiency: Prioritize zero-allocation paths for telemetry ingestion. Utilize sync.Pool and pre-allocated buffers.
3. Persistence Policy: Never write to the database synchronously on every sensor update. All persistent logic must pass through the Snapshot and Janitor services.
4. Interface Purity: All database operations must be contained within the internal/repository package. Do not leak SQL syntax into the business logic.
5. No Node.js at Runtime: The UI must be compiled to static assets and embedded using go:embed. No JavaScript runtime is permitted in the production binary.

## Development Workflow

1. Fork the repository and create your feature branch.
2. Ensure all code follows the Repository Pattern.
3. Verify that any new security logic is isolated or correctly integrated into the HMAC-Lite middleware.
4. Submit a Pull Request with a clear technical rationale for the change.

## Security

If you discover a security vulnerability, please report it via a private channel rather than opening a public issue. Protection of user data is our highest priority.

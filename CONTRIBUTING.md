# CONTRIBUTING TO XOMOI-CORE: THE SOVEREIGN FORGE

Welcome, contributor. By contributing to Xomoi-Core, you are building the foundation of Digital Freedom. 

Xomoi-Core is designed as a **Sovereign Monolith**—a single binary containing everything. We prioritize code purity, zero-dependency Go, and static-memory C++ SDKs.

## 1. OS-Agnostic Environment
You can contribute to Xomoi-Core from **Linux, macOS, or Windows**. We use tools that run identically on all platforms.

### Prerequisites:
- **Go 1.26+**
- **Task** (Taskfile runner): `task` or `go-task`.
- **Protoc** (Protobuf Compiler): `protoc` must be in your PATH.
- **Node.js 20+** (For the Svelte 5 UI build).

## 2. The Build Workflow
Instead of complex Makefiles, we use `Taskfile.yml`. These commands work on Bash, Zsh, and PowerShell:

- `task proto:go`: Generate Go internal models (Phase 1.1).
- `task proto:sdk`: Generate the adaptive C++ SDK (Phase 7).
- `task ui:build`: Compile the Svelte 5 frontend (Phase 6).
- `task build`: Build the final `xomoi` binary for your OS.
- `task clean`: Remove all generated artifacts.

## 3. Contribution Rules (The Rulebook)
Before writing any code, you MUST read the `docs/rulebook.md`. Key highlights:
- **No 3rd-Party Web Frameworks:** Use `net/http` ServeMux.
- **Zero-Allocation Ingestion:** Telemetry must not trigger heap escapes.
- **Repository Pattern:** Logic and storage must be separated by interfaces.

## 4. Submitting Changes
1. Fork the repository.
2. Create a branch: `feature/your-feature-name`.
3. Ensure all tests pass: `task test` (Coming soon).
4. Submit a PR with a detailed description of the architectural impact.

Join us in reclaiming the edge.

# CONTRIBUTING TO XOMOI-CORE: THE SOVEREIGN FORGE

Welcome, contributor. By contributing to Xomoi-Core, you are building the foundation of Digital Freedom. 

Xomoi-Core is designed as a **Sovereign Monolith**—a single binary containing everything. We prioritize code purity, zero-dependency Go, and static-memory C++ SDKs.

## 1. OS-Agnostic Environment
You can contribute to Xomoi-Core from **Linux, macOS, or Windows**. We use tools that run identically on all platforms.

### Prerequisites:
- **Go 1.26+**
- **Task** (Taskfile runner): `task` or `go-task`.
- **Node.js 20+** (For the Svelte 5 UI build).
- **Python 3.x & NanoPB** (For compiling the Protobuf headers).

## 2. The Build Workflow
Instead of complex Makefiles, we use `Taskfile.yml`. These commands work on Bash, Zsh, and PowerShell:

- `task proto:sdk`: Compile the `xomoi.proto` schema into `nanopb` C++ headers using the strict constraints in `xomoi.options`.
- `task ui:build`: Compile the Svelte 5 frontend using Vite.
- `task build`: Build the final `xomoi.exe` binary.

## 3. Contribution Rules (The Rulebook)
We have extremely strict architectural rules to maintain zero-bloat:
- **No 3rd-Party Web Frameworks in Go:** Use `net/http` ServeMux.
- **No Heavy Frontend Routers:** Use Svelte 5 `$effect` bindings to `window.location.hash`.
- **Zero-Allocation Ingestion:** Telemetry must not trigger heap escapes in Go.
- **Protobuf over JSON:** All IoT data must use the `.proto` schema. ESP32s cannot handle parsing large JSON strings.
- **Static Strings in C++:** When contributing to the SDK, you must use fixed-size `char` arrays dictated by `xomoi.options`. Dynamic allocation (`std::string`) is strictly forbidden to prevent heap fragmentation.

## 4. Submitting Changes
1. Fork the repository.
2. Create a branch: `feature/your-feature-name`.
3. Ensure all tests pass.
4. Submit a PR explaining *why* your architecture is mathematically optimal.

Join us in reclaiming the edge.

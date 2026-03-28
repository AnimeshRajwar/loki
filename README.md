# Loki

A minimal, educational Git-style version control system written in Go. Loki demonstrates the core ideas behind Git: object storage, staging, commits, and a simple CLI.

---

## Features

- **init**: Initialize a new Loki repository (displays current path on init)
- **add \<files\>**: Stage files for commit with multi-file support; errors on missing files, success confirmation for present files
- **commit -m "message"**: Commit staged files
- **status**: Show staged files and changes to be committed; handles empty index gracefully
- **log**: Show commit log with full output after each commit
- **User credential support**: Set and store user identity for commits
- **Colored CLI output**: All CLI outputs use colored text for better readability
- **Repository guard**: Commands cannot run when repository is not initialized
- **CI workflow**: Automated CI pipeline included
- Real object storage: blobs, trees, and commits (Git-style)

---

## Quick Start

1. **Clone or download this repository**
2. **Build the CLI:**

   **On Linux/macOS:**
```sh
   go build -o loki ./cmd/loki
```

   **On Windows (Command Prompt or PowerShell):**
```bat
   go build -o loki.exe cmd/loki/main.go
```

3. **Run commands:**

   **On Linux/macOS:**
```sh
   ./loki init
   ./loki add myfile.txt
   ./loki commit -m "first commit"
   ./loki status
   ./loki log
```

   **On Windows (Command Prompt or PowerShell):**
```bat
   .\loki init
   .\loki add myfile.txt
   .\loki commit -m "first commit"
   .\loki status
   .\loki log
```

**Optional:**
To run `loki` from anywhere, move it to your PATH:
```sh
sudo mv loki /usr/local/bin/
```

---

## How It Works

- **Blobs**: Store file contents (content-addressed, like Git)
- **Trees**: Store directory structure and file metadata
- **Commits**: Store project history and point to tree objects
- All objects are hashed (SHA-1) and stored in `.loki/objects/` using a split directory structure
- **User credentials**: Stored locally and attached to each commit
- **Colored output**: Uses terminal color codes for intuitive CLI feedback
- **Init path display**: `loki init` now shows the full path of the initialized repository

---

## Project Structure
```
loki/
├── cmd/
│   └── loki/
│       └── main.go                # CLI entry point
│
├── internal/
│   ├── commands/                  # Command handlers (init, add, commit, status, log, ...)
│   ├── core/                      # Core repository logic (index, repo, hash)
│   ├── models/                    # Data structures (blob, tree, commit)
│   └── storage/                   # Object storage (interface, file impl)
│
├── .github/
│   └── workflows/                 # CI workflow definitions
│
├── docs/
│   └── architecture.md            # Detailed architecture documentation
│
├── .gitignore
├── go.mod
├── README.md
└── project_structure.md
```

For a detailed explanation, see [`docs/architecture.md`](docs/architecture.md).

---

## CLI Behavior

### `loki init`
Initializes a new repository and displays the **current working directory path**:
```sh
Initialized empty Loki repository at /home/user/myproject/.loki
```

### `loki add <files>`
- Accepts **multiple files** at once: `./loki add file1.go file2.go`
- Shows an **error** if a specified file does not exist
- Shows a **success message** for each file successfully staged

### `loki status`
- Displays all staged files ready to be committed
- Gracefully handles an **empty index** (no staged files) without errors

### `loki log`
- Displays the full commit history
- Output is shown correctly **after every commit**

### `loki commit -m "message"`
- Requires user credentials to be set before committing
- Attaches author identity to each commit

---

## Example Workflow
```sh
./loki init
# Add files to staging (supports multiple files)
./loki add main.go README.md
# Commit staged files
./loki commit -m "Initial commit"
# Check status
./loki status
# View commit log
./loki log
```

---

## CI

This project includes a GitHub Actions CI workflow that automatically builds and tests Loki on every push and pull request.

---

## Contributing

Contributions, bug reports, and questions are welcome! Please open an issue or pull request.

---

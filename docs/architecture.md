# Loki Project Structure & File Responsibilities

Loki is a minimal, educational version control system inspired by Git. Below is an explanation of the project structure and the role of each major file/folder.

## Directory Layout

```
loki/
├── cmd/
│   └── loki/
│       └── main.go
│
├── internal/
│   ├── commands/
│   ├── core/
│   ├── models/
│   └── storage/
│
├── .gitignore
├── go.mod
├── README.md
└── project_structure.md
```

## Folder & File Responsibilities

### cmd/loki/main.go
- **Purpose:** Entry point for the CLI. Parses command-line arguments and dispatches to the appropriate command handler.
- **How it works:** Imports `internal/commands` and calls functions like `Init()`, `Add()`, `Commit()`, etc., based on user input.

### internal/commands/
- **Purpose:** Implements each CLI command as a function.
- **Files:**
  - `init.go`: Initializes a new Loki repository (`.loki/` folder).
  - `add.go`: Stages files for commit by adding them to the index.
  - `commit.go`: Creates a new commit from staged files.
  - `status.go`: Shows which files are staged, including their status as `added`, `modified`, or `deleted`.
  - `log.go`: Prints the commit log.
  - `help.go`: Prints help/usage information.

### internal/core/
- **Purpose:** Core repository logic and high-level operations.
- **Files:**
  - `repository.go`: Defines the `Repository` struct and methods for add, commit, status, and log. Orchestrates the main VCS operations.
  - `index.go`: Manages the staging area (index). Handles adding files, saving/loading the index, and writing tree objects.
  - `hash.go`: Helper for decoding SHA-1 hashes from hex strings.

### internal/models/
- **Purpose:** Data structures for versioned objects.
- **Files:**
  - `types.go`: Defines object types (blob, tree, commit).
  - `blob.go`: Implements the Blob object (file content snapshot).
  - `tree.go`: Implements the Tree object (directory structure snapshot).
  - `commit.go`: Implements the Commit object (history snapshot).

### internal/storage/
- **Purpose:** Handles object storage and retrieval.
- **Files:**
  - `storage.go`: Defines the storage interface and file-based storage implementation.
  - `objects.go`: Handles writing objects (blobs, trees, commits) to disk, using Git-style hashing and compression.

## How It All Fits Together

- **main.go** parses commands and calls the appropriate handler in `internal/commands/`.
- Each command handler (e.g., `add.go`, `commit.go`) uses the `Repository` from `internal/core/` to perform actions.
- The `Repository` uses the `Index` to track staged files and writes objects (blobs, trees, commits) using the storage layer.
- All objects are stored in `.loki/objects/` using SHA-1 hashes, similar to Git.

## Extending Loki
- New commands can be added in `internal/commands/`.
- New object types or storage backends can be added in `internal/models/` and `internal/storage/`.
- The architecture is modular, making it easy to add features like branches, checkout, or diff in the future.

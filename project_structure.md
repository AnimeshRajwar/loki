```
loki/
├── cmd/
│   └── loki/
│       └── main.go                # Entry point: CLI argument parsing
│
├── internal/
│   ├── commands/                  # Command handlers (called by main.go)
│   │   ├── init.go                # loki init
│   │   ├── add.go                 # loki add <files>
│   │   ├── commit.go              # loki commit -m "message"
│   │   ├── status.go              # loki status
│   │   ├── log.go                 # loki log
│   │   └── help.go                # loki help
│   │
│   ├── core/                      # Core repository logic
│   │   ├── repository.go          # Repository struct & high-level operations
│   │   ├── index.go               # Staging area (index file) operations
│   │   └── hash.go                # Hash decode helper
│   │
│   ├── models/                    # Data structures
│   │   ├── types.go               # ObjectType enum
│   │   ├── blob.go                # Blob struct & methods
│   │   ├── tree.go                # Tree struct & methods
│   │   └── commit.go              # Commit struct & methods
│   │
│   └── storage/                   # Storage abstraction layer
│       ├── storage.go             # Storage interface + FileStorage impl
│       └── objects.go             # Object storage (read/write/compress)
│
│   └── utils/                     # CLI and internal utility helpers
│       └── cli.go                 # Utility functions for CLI operations
│
├── .gitignore
├── go.mod
├── README.md
└── project_structure.md
```


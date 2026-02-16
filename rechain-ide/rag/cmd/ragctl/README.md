# ragctl

Watches a repo and periodically re-indexes files into the RAG service.

## Usage

```powershell
# index current directory every 10 seconds
./ragctl -rag http://localhost:8083 -repo myrepo -root . -interval 10
```

## Notes
- Skips .git, node_modules, and vendor directories.
- Indexes .go and .md files by default.

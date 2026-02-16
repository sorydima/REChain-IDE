# Git FAQ - Frequently Asked Questions

Common questions and answers about Git usage in the REChain Quantum-CrossAI IDE Engine project.

## Getting Started

### Q: How do I clone the repository?
```bash
git clone https://github.com/rechain-ai/rechain-ide.git
cd rechain-ide
```

### Q: How do I set up the repository for development?
```bash
./scripts/setup-repo.sh
```
This installs hooks, checks tools, and configures git.

### Q: What branch should I work on?
- **New features**: Create from `develop` → `feature/your-feature`
- **Bug fixes**: Create from `develop` → `bugfix/fix-name`
- **Hotfixes**: Create from `main` → `hotfix/critical-fix`

## Commits

### Q: What commit message format should I use?
Use [Conventional Commits](https://www.conventionalcommits.org/):
```
<type>[(scope)]: <description>

Examples:
feat(auth): add OAuth login
fix(parser): handle empty input
docs(api): update endpoints
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

### Q: I made a typo in my commit message. How do I fix it?
```bash
# Fix last commit message
git commit --amend -m "correct message"

# Push to remote (if already pushed)
git push origin branch-name --force-with-lease
```

### Q: I forgot to add a file to my last commit
```bash
# Add the forgotten file
git add forgotten-file.go

# Amend the commit
git commit --amend --no-edit
```

### Q: How do I undo my last commit but keep the changes?
```bash
git reset HEAD~1 --soft
# Now your changes are staged, ready for a new commit
```

## Branches

### Q: How do I create a new branch?
```bash
# Using the script (recommended)
./scripts/create-branch.sh feature my-feature

# Or manually
git checkout develop
git pull origin develop
git checkout -b feature/my-feature
```

### Q: How do I switch between branches?
```bash
# Switch to existing branch
git checkout branch-name

# Create and switch to new branch
git checkout -b new-branch

# Switch to previous branch
git checkout -
```

### Q: How do I delete a branch?
```bash
# Delete local branch
git branch -d branch-name

# Force delete (if not merged)
git branch -D branch-name

# Delete remote branch
git push origin --delete branch-name

# Clean up all merged branches
make clean-branches
```

### Q: How do I rename a branch?
```bash
# Rename current branch
git branch -m new-name

# Rename other branch
git branch -m old-name new-name
```

## Staging and Unstaging

### Q: How do I stage files?
```bash
# Stage specific file
git add filename.go

# Stage all changes
git add .

# Stage all Go files
git add *.go

# Interactively stage portions
git add -p
```

### Q: How do I unstage files?
```bash
# Unstage specific file
git restore --staged filename.go

# Or with older git
git reset HEAD filename.go

# Unstage all files
git restore --staged .
```

### Q: How do I discard local changes?
```bash
# Discard changes in specific file
git restore filename.go

# Discard all changes (careful!)
git restore .

# Or older git
git checkout -- filename.go
```

## Pulling and Pushing

### Q: How do I pull latest changes?
```bash
# Pull with merge commit
git pull origin main

# Pull with rebase (recommended)
git pull --rebase origin main
```

### Q: How do I push my branch?
```bash
# Push new branch
git push -u origin branch-name

# Push existing branch
git push origin branch-name

# Force push (use carefully!)
git push origin branch-name --force-with-lease
```

### Q: I get "rejected - non-fast-forward" error
```bash
# Fetch latest changes
git fetch origin

# Rebase your branch
git pull --rebase origin main

# Then push
git push origin branch-name
```

## Merging and Rebasing

### Q: What's the difference between merge and rebase?
- **Merge**: Creates a merge commit, preserves history
- **Rebase**: Rewrites history, linear commit history

Use merge for public branches, rebase for feature branches before merging.

### Q: How do I rebase my branch?
```bash
# Rebase onto develop
git checkout feature-branch
git rebase develop

# If conflicts, resolve them, then:
git add .
git rebase --continue
```

### Q: How do I abort a rebase?
```bash
git rebase --abort
```

### Q: How do I resolve merge conflicts?
```bash
# See conflicted files
git status

# Edit files to resolve conflicts
# Look for <<<<<<<, =======, >>>>>>> markers

# Stage resolved files
git add resolved-file.go

# Continue merge
git merge --continue

# Or abort
git merge --abort
```

## History and Inspection

### Q: How do I see commit history?
```bash
# Simple log
git log --oneline

# With graph
git log --oneline --graph --decorate --all

# Specific file history
git log --oneline -- filename.go

# With statistics
git log --stat
```

### Q: How do I see what changed in a commit?
```bash
git show commit-hash

# Or just stats
git show --stat commit-hash
```

### Q: How do I find who changed a line?
```bash
git blame filename.go

# Specific lines
git blame -L 10,20 filename.go
```

## Stashing

### Q: How do I stash changes?
```bash
# Stash current changes
git stash push -m "description"

# Or just
git stash
```

### Q: How do I apply stashed changes?
```bash
# List stashes
git stash list

# Apply most recent stash
git stash pop

# Apply specific stash
git stash pop stash@{2}

# Apply without removing from stash
git stash apply
```

### Q: How do I delete a stash?
```bash
# Drop most recent
git stash drop

# Drop specific
git stash drop stash@{1}

# Clear all stashes
git stash clear
```

## Tags

### Q: How do I create a tag?
```bash
# Lightweight tag
git tag v1.0.0

# Annotated tag (recommended)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push tags
git push origin v1.0.0

# Push all tags
git push origin --tags
```

### Q: How do I delete a tag?
```bash
# Delete local
git tag -d v1.0.0

# Delete remote
git push origin --delete v1.0.0
```

## Remotes

### Q: How do I add a remote?
```bash
git remote add origin https://github.com/rechain-ai/rechain-ide.git
git remote add upstream https://github.com/original/repo.git
```

### Q: How do I change remote URL?
```bash
git remote set-url origin https://new-url.git
```

### Q: How do I see configured remotes?
```bash
git remote -v
```

## Troubleshooting

### Q: I accidentally committed to main instead of a branch
```bash
# Create new branch from current state
git checkout -b feature-branch

# Reset main to previous commit
git checkout main
git reset HEAD~1 --hard

# Or if you already pushed (careful!)
git checkout main
git reset HEAD~1 --hard
git push origin main --force-with-lease
```

### Q: I committed a secret/password!
```bash
# If not pushed yet
git reset HEAD~1

# If already pushed (more complex)
# See git-filter-repo or BFG Repo-Cleaner
# Contact: security@rechain.ai
```

### Q: My repository is too large/slow
```bash
# Garbage collect
git gc --aggressive --prune=now

# Check what's taking space
git count-objects -vH

# See large files
git rev-list --objects --all | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' | awk '/^blob/ {print $3" "$4}' | sort -rn | head -20
```

### Q: Line ending issues on Windows
```bash
# Configure Git
git config core.autocrlf true

# Re-normalize
git add --renormalize .
git commit -m "Normalize line endings"
```

## Hooks

### Q: How do I install git hooks?
```bash
make git-hooks
# or
./scripts/install-hooks.sh
```

### Q: How do I bypass hooks temporarily?
```bash
# Skip pre-commit
git commit -m "message" --no-verify

# Skip pre-push
git push origin branch --no-verify
```

## Configuration

### Q: Where is my git config?
```bash
# Global config (all repos)
cat ~/.gitconfig

# Local config (this repo only)
cat .git/config

# System config
# Usually /etc/gitconfig or C:\Program Files\Git\etc\gitconfig
```

### Q: How do I set my name and email?
```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Or per-repo (remove --global)
git config user.name "Your Name"
```

## Best Practices

### Q: How often should I commit?
- Commit when you complete a logical unit of work
- Don't commit broken code
- Make commits atomic and focused

### Q: Should I use merge or rebase?
- **Feature branches**: Rebase onto develop before merging
- **Public/shared branches**: Always merge, never rebase

### Q: How do I write good commit messages?
- Use present tense ("Add feature" not "Added feature")
- First line under 72 characters
- Blank line, then detailed description if needed
- Reference issues: "Fixes #123"

## Still Have Questions?

- Check [GIT_CONFIGURATION.md](./GIT_CONFIGURATION.md)
- See [DEVELOPMENT_WORKFLOW.md](./DEVELOPMENT_WORKFLOW.md)
- Join [Discord](https://discord.gg/rechain)
- Create an issue with the "question" label

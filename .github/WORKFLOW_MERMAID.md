# Mermaid Workflow Diagrams
# These diagrams can be rendered in GitHub/GitLab Markdown

## Branching Strategy

```mermaid
graph TD
    A[main - Production] --> B[develop - Integration]
    B --> C[feature/login]
    B --> D[feature/api-v2]
    B --> E[bugfix/memory-leak]
    C --> F[PR to develop]
    D --> F
    E --> F
    F --> G[Merge to develop]
    G --> H[Release Branch]
    H --> I[Merge to main]
    I --> J[Tag v1.0.0]
    J --> K[Deploy Production]
    
    A --> L[hotfix/critical-fix]
    L --> M[PR to main]
    M --> N[Merge to main]
    N --> O[Cherry-pick to develop]
```

## PR Lifecycle

```mermaid
graph LR
    A[Create Branch] --> B[Make Changes]
    B --> C[Commit & Push]
    C --> D[Create PR]
    D --> E{CI Passes?}
    E -->|No| F[Fix Issues]
    F --> C
    E -->|Yes| G[Code Review]
    G --> H{Changes Requested?}
    H -->|Yes| I[Address Feedback]
    I --> C
    H -->|No| J[Approve]
    J --> K[Merge]
    K --> L[Delete Branch]
```

## CI/CD Pipeline

```mermaid
graph LR
    A[Push] --> B[Build]
    B --> C[Test]
    C --> D[Security Scan]
    D --> E{Pass?}
    E -->|No| F[Notify Failure]
    E -->|Yes| G[Deploy Staging]
    G --> H[Integration Tests]
    H --> I{Release?}
    I -->|Yes| J[Deploy Production]
    I -->|No| K[Done]
    J --> K
```

## Issue Lifecycle

```mermaid
graph TD
    A[New Issue] --> B{Triage}
    B --> C[Accepted]
    B --> D[Duplicate]
    B --> E[Invalid]
    C --> F[Backlog]
    F --> G[In Progress]
    G --> H[In Review]
    H --> I[Testing]
    I --> J[Done]
    D --> K[Close]
    E --> K
```

## Git Operations Flow

```mermaid
graph TD
    A[Working Directory] -->|git add| B[Staging Area]
    B -->|git commit| C[Local Repository]
    C -->|git push| D[Remote Repository]
    D -->|PR/MR| E[Code Review]
    E -->|Merge| F[Main Branch]
    F -->|git pull| C
```

## Release Process

```mermaid
graph TD
    A[develop branch] -->|Feature freeze| B[release/v1.2.0]
    B --> C[Version bump]
    C --> D[Final testing]
    D --> E{Tests pass?}
    E -->|No| F[Fix issues]
    F --> D
    E -->|Yes| G[PR to main]
    G --> H[Merge to main]
    H --> I[Tag v1.2.0]
    I --> J[Deploy production]
    H --> K[Merge back to develop]
```

## Platform Sync

```mermaid
graph LR
    A[GitHub - Primary] -->|Push mirror| B[GitLab - Mirror]
    A --> C[Issues/PRs]
    B --> D[CI/CD Pipeline]
    C --> E[Development]
    D --> E
    E --> F[Deploy]
```

## Conventional Commits

```mermaid
graph TD
    A[Code Change] --> B{Type?}
    B -->|New feature| C[feat: description]
    B -->|Bug fix| D[fix: description]
    B -->|Docs| E[docs: description]
    B -->|Refactor| F[refactor: description]
    B -->|Test| G[test: description]
    B -->|Chore| H[chore: description]
    C --> I[Commit]
    D --> I
    E --> I
    F --> I
    G --> I
    H --> I
```

## Fork Workflow

```mermaid
graph LR
    A[Upstream Repo] -->|Fork| B[Your Fork]
    B --> C[Clone Locally]
    C --> D[Create Branch]
    D --> E[Make Changes]
    E --> F[Push to Fork]
    F --> G[Create PR]
    G --> H[Code Review]
    H -->|Approved| I[Merge to Upstream]
    I --> J[Sync Fork]
```

## Hotfix Process

```mermaid
graph TD
    A[Critical Bug in Production] -->|Emergency| B[Create hotfix branch from main]
    B --> C[Fix bug]
    C --> D[Test locally]
    D --> E[Expedited review]
    E --> F[Merge to main]
    F --> G[Tag hotfix version]
    G --> H[Deploy immediately]
    F --> I[Cherry-pick to develop]
```

## Stash Workflow

```mermaid
graph LR
    A[Working on feature] -->|Emergency bug| B[git stash]
    B --> C[Checkout main]
    C --> D[Create hotfix branch]
    D --> E[Fix bug]
    E --> F[Commit & push]
    F --> G[Checkout feature]
    G --> H[git stash pop]
    H --> I[Continue work]
```

---

## Usage

These diagrams can be included in Markdown files:

```markdown
```mermaid
[diagram code]
```
```

GitHub and GitLab both render Mermaid diagrams in:
- Issue descriptions
- Pull/Merge request descriptions
- Wiki pages
- Markdown files

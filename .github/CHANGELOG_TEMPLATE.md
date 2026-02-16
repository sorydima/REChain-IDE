# Changelog Template
# Template for maintaining CHANGELOG.md following Keep a Changelog format

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Now removed features

### Fixed
- Bug fixes

### Security
- Security improvements

## [1.0.0] - YYYY-MM-DD

### Added
- Initial release
- Feature 1
- Feature 2

### Changed
- Changed something

### Fixed
- Fixed bug #123

### Security
- Security fix

---

## Template Categories

### Added
- New features
- New functionality
- New integrations
- New documentation

### Changed
- Changes in existing functionality
- Performance improvements
- UI/UX changes
- Dependency updates

### Deprecated
- Features marked for removal
- Warnings about future changes

### Removed
- Removed features
- Removed functionality
- Removed dependencies

### Fixed
- Bug fixes
- Performance fixes
- Corrected errors

### Security
- Security fixes
- Vulnerability patches
- Authentication improvements

## Writing Guidelines

### Do
- Keep it human-readable
- Group changes by type
- Reference issues/PRs when possible
- Use present tense ("Add feature" not "Added feature")
- Include migration notes for breaking changes

### Don't
- Use technical jargon without explanation
- Include commit hashes (use tags/versions)
- Mix different types of changes in one entry
- Forget to update before releasing

## Versioning

Format: MAJOR.MINOR.PATCH

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Example Entries

### Good
```markdown
### Added
- Add support for Solidity 0.9 compilation (#456)
- Implement dark mode theme (#234)

### Fixed
- Fix memory leak in parser (#789)
- Resolve race condition in worker pool (#567)
```

### Bad
```markdown
- stuff was added
- fixed bugs
- updated code
```

## Release Checklist

Before creating a release:
- [ ] Update version number
- [ ] Update CHANGELOG.md with new version
- [ ] Ensure all PRs are merged
- [ ] Ensure all issues are closed/resolved
- [ ] Run full test suite
- [ ] Update documentation
- [ ] Create git tag
- [ ] Build release artifacts
- [ ] Publish release notes

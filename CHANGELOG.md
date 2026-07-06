# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> [!NOTE]
> Moving forward, changes will be tracked and linked using their corresponding GitHub issue numbers (e.g., `[#123]`). Historical commits are linked directly to their commit hashes.



## [Unreleased]

### Added

- Auto-discover and import configurations (aliases, environment variables, custom functions, snippets) from local shell profiles, ~/.config, VS Code user snippets, and Pet manager TOML configs during setup ([#1](https://github.com/aaryansinhaa/reshell/issues/1))
- `changelog` to keep track of the changes in the project ([#2](https://github.com/aaryansinhaa/reshell/issues/2))
- `reshell` environment manager implementation ([fb60a6e](https://github.com/aaryansinhaa/reshell/commit/fb60a6e))
- Git-based version control, TUI history revert, and tests ([dc56c33](https://github.com/aaryansinhaa/reshell/commit/dc56c33))
- Multi-profile switcher via CLI and TUI ([94aa0f6](https://github.com/aaryansinhaa/reshell/commit/94aa0f6))
- Profile-specific Git history isolation and history clearing ([c091775](https://github.com/aaryansinhaa/reshell/commit/c091775))
- Unit tests for scripts parser and snippets manager ([8d58273](https://github.com/aaryansinhaa/reshell/commit/8d58273))
- GitHub Actions workflow for tests and build verification ([f36e74e](https://github.com/aaryansinhaa/reshell/commit/f36e74e))
- GitHub Actions workflow to deploy mkdocs documentation to GitHub Pages ([0c925cb](https://github.com/aaryansinhaa/reshell/commit/0c925cb))
- Docker publishing pipeline and viewing script execution on the viewport ([d36350c](https://github.com/aaryansinhaa/reshell/commit/d36350c))
- Comprehensive unit tests and UX logic enhancement ([f0c5bbc](https://github.com/aaryansinhaa/reshell/commit/f0c5bbc))
- Code coverage reporting in CI ([0bfe434](https://github.com/aaryansinhaa/reshell/commit/0bfe434))
- MIT License ([c7c01d5](https://github.com/aaryansinhaa/reshell/commit/c7c01d5))

### Changed

- Overhaul of documentation to make it more accessible and less verbose ([6076df0](https://github.com/aaryansinhaa/reshell/commit/6076df0))
- Add GitHub repository card and Shields.io star badge to docs ([5d2708c](https://github.com/aaryansinhaa/reshell/commit/5d2708c), [8cfe61b](https://github.com/aaryansinhaa/reshell/commit/8cfe61b))
- CI/CD workflow improvements: only deploy documentation when docs files change ([b4ddcab](https://github.com/aaryansinhaa/reshell/commit/b4ddcab)), modify Go setup and version configuration in Docker publishing ([4342e78](https://github.com/aaryansinhaa/reshell/commit/4342e78), [8b55bc0](https://github.com/aaryansinhaa/reshell/commit/8b55bc0)), add Codecov token to coverage report upload step ([faada9b](https://github.com/aaryansinhaa/reshell/commit/faada9b))
- Documentation updates: add screenshots to README/docs ([8f78b5f](https://github.com/aaryansinhaa/reshell/commit/8f78b5f)), add test coverage section ([0620a2a](https://github.com/aaryansinhaa/reshell/commit/0620a2a)), general documentation enhancement ([9548975](https://github.com/aaryansinhaa/reshell/commit/9548975)), minor README/docs fixes/edits ([73f89fe](https://github.com/aaryansinhaa/reshell/commit/73f89fe), [2424dd9](https://github.com/aaryansinhaa/reshell/commit/2424dd9), [e26dce4](https://github.com/aaryansinhaa/reshell/commit/e26dce4), [4f94f0a](https://github.com/aaryansinhaa/reshell/commit/4f94f0a))

### Fixed

- favourite and tag issues and other minor issues in the snippets tab ([#6](https://github.com/aaryansinhaa/reshell/issues/6))
- Formatting issues in the `docs/` files ([#7](https://github.com/aaryansinhaa/reshell/issues/7))
- Preview window formatting issues ([9696b83](https://github.com/aaryansinhaa/reshell/commit/9696b83))
- Formatting issue for larger snippets and improved conflict checking ([980c0a7](https://github.com/aaryansinhaa/reshell/commit/980c0a7))
- CI fixes: disable setup-go cache to prevent extraction failure ([798a89e](https://github.com/aaryansinhaa/reshell/commit/798a89e)), remove Python dependency caching to fix setup-python error ([715bfd6](https://github.com/aaryansinhaa/reshell/commit/715bfd6))
- Minor formatting fixes in docs ([d294866](https://github.com/aaryansinhaa/reshell/commit/d294866))

### Security

- Hardened security measures ([f0c5bbc](https://github.com/aaryansinhaa/reshell/commit/f0c5bbc))
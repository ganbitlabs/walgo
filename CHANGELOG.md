# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-10-16

### Fixed
- **Deploy safety:** `--dry-run` never publishes, even when the cache is unavailable. Tool updates are never prompted for or installed without a terminal, in `--quiet` mode or during a dry run. A deploy that succeeded on-chain is no longer reported as failed: the object ID falls back to `ws-resources.json`, and bookkeeping errors are warnings that print it.
- **Toolchain gate:** deploy, update and `projects update` refuse to run with sui < 1.75.0, walrus < 1.52.0 or site-builder < 2.12.0 (`--skip-version-check` downgrades this to a warning).
- **Networks:** cost estimates use the network the deploy resolves to. `projects update` and `domain` check the active network. `launch` defaults to the active network, refuses to update a site recorded on another network, saves the network to `walgo.yaml`, and makes an imported wallet the active address.
- **Unattended runs:** end of input at any confirmation now means "no".
- **Optimizer:** HTML, CSS and JS minification now uses tdewolff/minify (the library Hugo uses). The previous minifiers could break JavaScript strings and template literals, inline whitespace and CSS selectors. Each optimizer runs only when its `minifyHTML`/`minifyCSS`/`minifyJS` flag is on.
- **Builds:** `publishDir` is honored from `walgo.yaml` or the Hugo config (absolute paths included) and is refused when it contains the site or a Hugo source folder. All Hugo config formats are detected (`hugo.*`, `config.*`, `config/_default/`). Plain static sites deploy without a Hugo config. User `.gz`/`.br`/`.map` files are no longer deleted. A missing `ws-resources.json` is created.
- **Security:** AI-generated plans can't write outside `content/`. Theme installs accept only `https://github.com/<owner>/<repo>` and can't extract outside `themes/`. Secrets are not echoed while typed and are scrubbed from errors.
- **`walgo.yaml`:** updating the project ID keeps comments and key order.
- **Desktop:** deploys and updates go through the same pipeline as the CLI and refuse on a network or wallet mismatch. Deploys can be cancelled, and quitting during one asks first. Failures stay on screen with Hugo's output. The preview server only stops its own Hugo process, picks a free port, and reports success only once Hugo is serving. File access is limited to registered site folders. Deleting a project keeps its folder unless you opt in, refuses to touch home or another project, and stops if the on-chain destroy fails. The markdown preview is sanitized, and links open in the system browser. The editor never saves an empty buffer over a file it failed to read.

### Added
- `site.kind: hugo|static` in `walgo.yaml` to override site detection.

### Changed
- `walgo projects update` runs through the shared deploy pipeline.
- The projects database stores absolute site paths.
- CI builds and tests the desktop app on every pull request. Release builds pin the Wails CLI and stamp the product version.

## [0.2.1] - 2025-01-21

### Fixed
- **Version display:** Fixed `walgo version` returning incorrect version number
- **Obsidian import:** Fixed `REF_NOT_FOUND` errors by changing default link style from `relref` to plain markdown links
- **README:** Corrected command from `import-obsidian` to `import <vault>`
- **HTTP deployer:** Added compatibility with site-builder v2 API response formats

### Added
- **Hugo Extended check:** `walgo doctor` now detects if Hugo Extended is installed and warns if standard Hugo is used
- **Link style option:** New `--link-style` flag for `walgo import` command (`markdown` or `relref`)
- **Benchmark script:** Added `scripts/benchmark.sh` to measure caching performance improvements
- **LinkStyle config:** New `linkStyle` option in `walgo.yaml` under `obsidian` section

### Changed
- Updated README to document Hugo Extended requirement
- Improved error messages in HTTP deployer to show raw response on parse failures

## [0.2.0] - 2025-01-14

### Added - Phase 2: Performance & Enhanced Features

#### Incremental Builds & Caching
- SQLite-based file caching system for intelligent change detection
- SHA-256 file hashing for accurate modification tracking
- Deployment plan analysis showing added/modified/unchanged files
- Cache integration with `build`, `deploy`, and `update` commands
- **Performance:** 50-90% faster redeploys on unchanged content

#### Brotli Compression
- Automatic Brotli compression for text assets (HTML, CSS, JS, JSON, SVG)
- Smart compression that only applies when beneficial
- Configurable compression levels (default: level 6)
- `walgo compress` standalone command for manual compression
- **Storage Savings:** ~30-64% reduction in file sizes

#### Cache-Control Headers
- Automatic `ws-resources.json` generation for Walrus Sites
- Intelligent immutable vs mutable asset detection
- Content-Type and Content-Encoding header management
- Configurable max-age for different asset types
- **Result:** Proper browser caching for decentralized sites

#### Enhanced Obsidian Import
- Full wikilink conversion: `[[link]]`, `[[link|alias]]`, `[[link#heading]]`
- Transclusion support: `![[file]]` embedded in Hugo shortcodes
- Block reference support: `[[file^block-id]]`
- Frontmatter preservation for tags and aliases
- `--dry-run` mode to preview import without copying files

#### Telemetry Module (Opt-in)
- Local-only metrics collection in `~/.walgo/metrics.json`
- Build metrics: Hugo duration, compression stats, file counts
- Deploy metrics: Upload duration, file changes, deployment stats
- **Privacy:** No PII collection, entirely local storage
- `--telemetry` flag for `build`, `deploy`, and `update` commands

#### Dry-Run Modes
- `walgo deploy --dry-run` - Preview deployment plan
- `walgo update --dry-run` - Preview update plan
- `walgo import --dry-run` - Preview Obsidian import

### Added - Other Features
- New `deploy-http` command to publish the built site via Walrus HTTP APIs (publisher/aggregator) on Testnet. No wallet/funds required; returns quiltId and patchIds.
- New `doctor` command to diagnose environment issues (binaries, Sui env/address, gas), and optionally fix tildes in `sites-config.yaml` with `--fix-paths`.
- New `setup-deps` command to download and install `site-builder` and `walrus` to a managed bin dir and wire `walrus_binary` in `sites-config.yaml`.
- `walgo setup --force` flag to overwrite an existing `sites-config.yaml` with absolute wallet/config paths.
- Optional `--network devnet` to scaffold config for HTTP-only workflows. (On-chain deploy on devnet still requires proper Walrus client objects.)

### Changed
- Build command now includes compression and ws-resources.json generation by default
- Deploy and update commands show cache analysis and file change statistics
- Removed deprecated `--context` flag usage in site-builder calls. Network selection now comes from `~/.config/walrus/sites-config.yaml`.
- Improved `init` guidance to show HTTP vs on-chain deployment options.

### Fixed
- Fixed ineffectual assignment in build command step counter
- Replaced deprecated `filepath.HasPrefix` with `strings.HasPrefix`
- Fixed unhandled errors in compression and cache cleanup paths
- Code formatting now passes `gofmt` and `golangci-lint` checks
- Security warnings resolved with proper error handling and nosec justifications

### Performance Improvements
- **50-90% faster redeploys** through incremental caching
- **30-64% storage reduction** via Brotli compression
- **Optimized upload** - only changed files are uploaded
- **Efficient builds** - skip unnecessary recompression

### Known limitations
- On-chain `deploy`/`update` still require a correctly configured Walrus client and a funded Sui wallet on the selected network. Testnet HTTP publishing can be used as a no-funds alternative.

## [0.1.0] - 2024-12-19

### Added
- Initial public release
- Complete Hugo integration with `init`, `build`, `serve`, and `new` commands
- Full Walrus Sites deployment support with `deploy` and `update` commands
- Obsidian vault import functionality with wikilink conversion
- SuiNS domain configuration guidance via `domain` command
- Site status monitoring and object ID conversion utilities
- Asset optimization engine for HTML, CSS, and JavaScript
- Comprehensive CLI with help system and configuration management
- YAML-based configuration with sensible defaults
- Multi-format frontmatter support (YAML, TOML, JSON)
- Cross-platform support (macOS, Linux, Windows)

### Features
- 🚀 Hugo site initialization with Walrus-specific configuration
- 📦 One-command deployment to decentralized storage
- 🔄 Efficient site updates without full redeployment
- 📝 Content creation with `walgo new` command
- 🌐 SuiNS domain management guidance
- 📊 Site resource monitoring and status checking
- 🔧 Obsidian to Hugo migration toolkit
- ⚡ Built-in asset optimization and minification
- 🎯 Flexible configuration system
- 📋 Comprehensive documentation and examples

[Unreleased]: https://github.com/ganbitlabs/walgo/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/ganbitlabs/walgo/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/ganbitlabs/walgo/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ganbitlabs/walgo/releases/tag/v0.1.0 
<p align="center">
  <img src="github_walgo.png" alt="Walgo — Markdown to Web3, Simplified" width="100%" />
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License" /></a>
  <a href="https://github.com/selimozten/walgo/actions/workflows/ci.yml"><img src="https://github.com/selimozten/walgo/actions/workflows/ci.yml/badge.svg" alt="Build Status" /></a>
  <a href="https://walrus.xyz"><img src="https://img.shields.io/badge/Walrus%20RFP-Winner-F97316.svg" alt="Walrus RFP Winner" /></a>
</p>

<p align="center">
  <strong>The official CLI tool for deploying static sites to Walrus decentralized storage.</strong><br />
  Deploy your Hugo sites to the decentralized web with an interactive wizard. No blockchain experience required.
</p>

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Configuration](#configuration)
- [Documentation](#documentation)
- [Requirements](#requirements)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [About](#about)
- [License](#license)

---

## Features

| Feature                    | Description                                                                  |
| -------------------------- | ---------------------------------------------------------------------------- |
| **Interactive Deployment** | `walgo launch` guides you step-by-step through the entire deployment process |
| **Project Management**     | Track all your sites, deployment history, and updates with `walgo projects`  |
| **AI Content Generation**  | Generate and update blog posts with AI assistance                            |
| **Free Testing Mode**      | HTTP deployment requires no wallet or cryptocurrency                         |
| **Asset Optimization**     | Automatic HTML, CSS, and JavaScript minification                             |
| **Brotli Compression**     | Pre-compressed files for faster load times                                   |
| **Obsidian Integration**   | One-command site creation from Obsidian vaults                               |
| **Desktop App**            | Optional GUI for visual project management                                   |
| **SuiNS Support**          | Custom domain names via Sui Name Service                                     |

---

## Installation

### Quick Install (Recommended)

**macOS/Linux:**

```bash
# Prerequisites: Install Git first
# macOS: brew install git
# Ubuntu/Debian: sudo apt install git
# Fedora: sudo dnf install git
# Arch: sudo pacman -S git

curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash
```

**Windows:**

```powershell
# Prerequisites: Install Git first
# Using Chocolatey (recommended):
# choco install git
# Or download from: https://git-scm.com/download/win

irm https://raw.githubusercontent.com/selimozten/walgo/main/install.ps1 | iex
```

Or download the binary manually from the [releases page](https://github.com/selimozten/walgo/releases/latest).

<details>
<summary><strong>🪟 Windows Installation Details (Click to Expand)</strong></summary>

#### Method 1: PowerShell One-Line Install (Recommended)

```powershell
irm https://raw.githubusercontent.com/selimozten/walgo/main/install.ps1 | iex
```

This automatically detects your architecture, downloads binary, and installs it.

#### Method 2: Download and Install Manually

1. Download from [releases page](https://github.com/selimozten/walgo/releases/latest):

   - `walgo-windows-amd64.exe` (64-bit)
   - `walgo-windows-arm64.exe` (ARM64)

2. Add to PATH (PowerShell as Administrator):

   ```powershell
   New-Item -ItemType Directory -Force -Path "C:\Program Files\walgo"
   Move-Item walgo.exe "C:\Program Files\walgo\"
   [Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\walgo", "Machine")
   ```

3. **Important:** Close and reopen terminal for PATH changes to take effect.

#### Common Windows Issues

**"execution of scripts is disabled"**

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Windows Defender SmartScreen warning**

1. Click "More info"
2. Click "Run anyway"
3. Safe to run - source code is verified on GitHub

For complete Windows installation guide, see [docs/INSTALLATION.md](docs/INSTALLATION.md#windows).

</details>

### Using Go

```bash
go install github.com/selimozten/walgo@latest
```

### Build from Source

```bash
git clone https://github.com/selimozten/walgo.git
cd walgo
make build
```

### Verify Installation

```bash
walgo version
```

### Desktop App (Optional)

Download the desktop GUI from the [releases page](https://github.com/selimozten/walgo/releases).

#### macOS Installation

**Security Note:** The app is ad-hoc signed but not notarized with Apple. On first launch, macOS may show an "unidentified developer" warning.

**Method 1: Right-click to Open (Recommended)**

1. Download and extract `walgo-desktop_*_darwin_*.tar.gz`
2. Move `Walgo.app` to your Applications folder
3. **Right-click** on `Walgo.app` and select **"Open"**
4. Click **"Open"** again in the security dialog
5. The app will now open and be trusted for future launches

**Method 2: Terminal**

```bash
# Remove quarantine and open
xattr -cr /Applications/Walgo.app
open /Applications/Walgo.app
```

> If you used the install script (`curl ... | bash`), quarantine removal is handled automatically.

#### Windows Installation

1. Download `walgo-desktop_*_windows_amd64.zip`
2. Extract and run `walgo-desktop.exe`
3. Windows Defender may show a warning - click "More info" → "Run anyway"

#### Linux Installation

1. Download `walgo-desktop_*_linux_amd64.tar.gz`
2. Extract and run `./walgo-desktop`
3. Make executable if needed: `chmod +x walgo-desktop`

---

## Quick Start

### Option 1: Quickstart Command (Fastest)

Create, build, and deploy a new site in one command:

```bash
walgo quickstart my-blog
```

This will:

1. Create a new Hugo site with a theme
2. Add sample content
3. Build and optimize the site
4. Offer deployment options

### Option 2: Step-by-Step Setup

```bash
# 1. Create a new site
walgo init my-site
cd my-site

# 2. Build your site
walgo build

# 3. Deploy with the interactive wizard
walgo launch
```

### The Launch Wizard

`walgo launch` is the **recommended way to deploy**. It guides you through:

1. **Network Selection** — Choose testnet or mainnet
2. **Wallet Setup** — View balance, switch or create addresses
3. **Project Details** — Name your project for easy management
4. **Storage Duration** — Set how long to store your site (epochs)
5. **Cost Review** — See estimated gas fees before confirming
6. **Deployment** — Upload to Walrus with progress feedback
7. **SuiNS Setup** — Instructions to configure your custom domain

```bash
cd my-site
walgo launch
```

---

## Commands

### Core Workflow

| Command                   | Description                                     |
| ------------------------- | ----------------------------------------------- |
| `walgo quickstart <name>` | Create, build, and deploy a new site            |
| `walgo init <name>`       | Initialize a new Hugo site                      |
| `walgo build`             | Build site with optimization                    |
| `walgo serve`             | Start local development server                  |
| `walgo launch`            | **Interactive deployment wizard (recommended)** |

### Project Management

| Command                         | Description                      |
| ------------------------------- | -------------------------------- |
| `walgo projects`                | List all your deployed sites     |
| `walgo projects show <name>`    | View project details and history |
| `walgo projects update <name>`  | Update an existing deployment    |
| `walgo projects archive <name>` | Archive a project                |

### Content & AI

| Command                  | Description                                |
| ------------------------ | ------------------------------------------ |
| `walgo new <path>`       | Create new content file                    |
| `walgo import <vault>`   | Create site and import from Obsidian vault |
| `walgo ai configure`     | Set up AI provider (OpenAI/OpenRouter)     |
| `walgo ai generate`      | Generate content with AI                   |
| `walgo ai update <file>` | Update content with AI                     |

### Utilities

| Command             | Description                 |
| ------------------- | --------------------------- |
| `walgo doctor`      | Diagnose environment issues |
| `walgo status <id>` | Check deployment status     |
| `walgo optimize`    | Optimize HTML/CSS/JS files  |
| `walgo compress`    | Apply Brotli compression    |
| `walgo domain`      | SuiNS domain management     |

### Advanced Deployment

| Command             | Description                                  |
| ------------------- | -------------------------------------------- |
| `walgo deploy`      | Direct on-chain deployment (requires wallet) |
| `walgo deploy-http` | HTTP deployment (no wallet, testnet only)    |
| `walgo update <id>` | Update existing site by object ID            |

> **Note:** For most users, `walgo launch` is the recommended deployment method as it provides guidance and manages your projects automatically.

---

## Configuration

Walgo uses a `walgo.yaml` configuration file in your project directory:

```yaml
# Hugo settings
hugo:
  publishDir: "public"
  contentDir: "content"
  build_draft: false
  minify: true

# Walrus deployment settings
walrus:
  project_id: "" # Set automatically after first deploy
  epochs: 5 # Storage duration
  entrypoint: "index.html"

# Asset optimization
optimizer:
  enabled: true
  html:
    enabled: true
    minify_whitespace: true
    remove_comments: true
  css:
    enabled: true
    minify: true
  js:
    enabled: true
    minify: true

# Brotli compression
compress:
  enabled: true
  brotli_level: 6

# AI content generation (optional)
ai:
  enabled: false
  provider: "openai" # or "openrouter"
  model: "gpt-4"

# Obsidian import settings (optional)
obsidian:
  vault_path: ""
  include_drafts: false
  convert_wikilinks: true
```

### Environment Variables

| Variable        | Description                    |
| --------------- | ------------------------------ |
| `WALGO_ASCII=1` | Force ASCII output (no emojis) |
| `WALGO_EMOJI=1` | Force emoji output             |

---

## Documentation

Comprehensive documentation is available in the [`docs/`](docs/) directory:

### Getting Started

- [Quickstart Guide](docs/QUICKSTART.md) — Deploy in 2 minutes
- [Installation Guide](docs/INSTALLATION.md) — All platforms
- [Getting Started](docs/GETTING_STARTED.md) — First deployment walkthrough

### User Guides

- [Commands Reference](docs/COMMANDS.md) — All 29+ commands documented
- [Launch Wizard](docs/LAUNCH_WIZARD.md) — Interactive deployment guide
- [Configuration](docs/CONFIGURATION.md) — All options explained
- [Deployment Guide](docs/DEPLOYMENT.md) — HTTP vs on-chain deployment
- [AI Features](docs/AI_FEATURES.md) — Content generation with AI
- [Troubleshooting](docs/TROUBLESHOOTING.md) — Common issues

### Developer Guides

- [Architecture](docs/ARCHITECTURE.md) — System design
- [Development](docs/DEVELOPMENT.md) — Contributing setup
- [Contributing](docs/CONTRIBUTING.md) — How to contribute

---

## Requirements

### Required

- **[Git](https://git-scm.com)** — Version control system (required for installation and some features)

  ```bash
  # macOS
  brew install git

  # Ubuntu/Debian
  sudo apt install git

  # Fedora
  sudo dnf install git

  # Windows (Chocolatey)
  choco install git

  # Windows (Scoop)
  scoop install git

  # Or download from https://git-scm.com
  ```

- **[Hugo Extended](https://gohugo.io)** — Static site generator

  > **Important:** The **Extended** version is required for SCSS/SASS support.

  ```bash
  # macOS
  brew install hugo

  # Ubuntu/Debian
  sudo apt install hugo

  # Fedora
  sudo dnf install hugo

  # Windows (Chocolatey)
  choco install hugo-extended

  # Windows (Scoop)
  scoop install hugo-extended

  # Verify it's the extended version
  hugo version  # Should show "extended"
  ```

### For On-Chain Deployment

- **[site-builder](https://docs.walrus.site/walrus-sites/overview)** — Walrus Sites CLI

  ```bash
  walgo setup-deps  # Installs automatically
  ```

- **[Sui Wallet](https://docs.sui.io/guides/developer/getting-started/sui-install)** — With SUI tokens

  ```bash
  # Set up wallet for testnet
  walgo setup --network testnet
  ```

### Optional

- **Obsidian** — For vault import feature
- **OpenAI/OpenRouter API Key** — For AI content generation

---

## Troubleshooting

### Quick Diagnostics

```bash
walgo doctor           # Check your setup
walgo doctor -v        # Verbose diagnostics
walgo doctor --fix-all # Auto-fix common issues
```

### Common Issues

| Issue                    | Solution                                             |
| ------------------------ | ---------------------------------------------------- |
| `site-builder not found` | Run `walgo setup-deps`                               |
| `insufficient funds`     | Get testnet SUI from [faucet](https://faucet.sui.io) |
| `Hugo not extended`      | Reinstall Hugo Extended version                      |
| `network timeout`        | Retry with `--epochs 1` or check connectivity        |

### Getting Help

- **Documentation:** [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
- **GitHub Issues:** [Report a bug](https://github.com/selimozten/walgo/issues)
- **GitHub Discussions:** [Ask questions](https://github.com/selimozten/walgo/discussions)
- **Feedback Form:** [Submit feedback](https://tally.so/r/447YBO)

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/selimozten/walgo.git
cd walgo

# Install dependencies and build
make build

# Run tests
make test

# Run linter
make lint
```

### Project Structure

```
walgo/
├── cmd/           # CLI commands
├── internal/      # Internal packages
│   ├── ai/        # AI content generation
│   ├── cache/     # Deployment caching
│   ├── compress/  # Brotli compression
│   ├── config/    # Configuration management
│   ├── deployer/  # Deployment adapters
│   ├── hugo/      # Hugo integration
│   ├── launch/    # Launch wizard
│   ├── obsidian/  # Obsidian import
│   ├── optimizer/ # Asset optimization
│   ├── projects/  # Project management
│   ├── ui/        # Terminal UI helpers
│   └── walrus/    # Walrus CLI wrapper
├── docs/          # Documentation
├── desktop/       # Desktop app (Wails)
└── tests/         # Integration tests
```

---

## About

**Walgo** is developed by the [Ganbitera](https://ganbitera.io) team as a winner of the [Walrus RFP](https://walrus.xyz) (Request for Proposals) for developer tooling.

### Why Walgo?

- **Simple** — One command to deploy: `walgo launch`
- **Guided** — Interactive wizard for beginners
- **Powerful** — Project management, AI content, optimization
- **Flexible** — Free HTTP testing or permanent on-chain storage
- **Official** — Supported by the Walrus ecosystem

### What is Walrus?

[Walrus](https://walrus.xyz) is a decentralized storage network built on [Sui](https://sui.io) that enables permanent, censorship-resistant hosting of websites and applications.

### Links

- **Walrus Documentation:** [docs.walrus.site](https://docs.walrus.site)
- **Sui Documentation:** [docs.sui.io](https://docs.sui.io)
- **SuiNS (Domains):** [suins.io](https://suins.io) | [Tutorial](https://docs.wal.app/docs/walrus-sites/tutorial-suins)

---

## License

MIT License — See [LICENSE](LICENSE) for details.

---

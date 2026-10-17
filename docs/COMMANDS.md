# Walgo - Complete Command Reference

**Walgo v0.5.0** - Deploy Hugo Sites to Walrus Decentralized Storage

Complete reference for all Walgo commands with examples, flags, and workflows.

**Flag notation.** A flag shown without a value is a boolean switch. It is off unless its entry gives another default. `<int>` takes a whole number. Every other placeholder takes a string.

---

## Recommended Workflow

```bash
# 1. Create your site
walgo quickstart my-blog
cd my-blog

# 2. Build and deploy with the interactive wizard
walgo build
walgo launch    # <-- Recommended deployment method
```

> **Note:** `walgo launch` is the recommended way to deploy. It provides step-by-step guidance, project tracking, and cost estimation.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Site Management](#site-management)
- [Content Management](#content-management)
- [Build & Optimization](#build--optimization)
- [Deployment](#deployment)
- [Project Management](#project-management)
- [AI Features](#ai-features)
- [Desktop App](#desktop-app)
- [Setup & Configuration](#setup--configuration)
- [Diagnostics & Utilities](#diagnostics--utilities)
- [Common Workflows](#common-workflows)
- [Global Flags](#global-flags)

---

## Quick Start

### `walgo quickstart <site-name>`

**One-command setup: initialize and build your site**

```bash
walgo quickstart my-blog
walgo quickstart my-portfolio --skip-build
```

**What it does:**

1. Asks for a site type: Biolink (default), Blog, Docs, or Whitepaper
2. Creates a Hugo site in `./<site-name>` with a `walgo.yaml`
3. Installs the theme for that site type: Walgo Biolink, Ananke, Book, or Walgo Whitepaper
4. Adds sample content: a homepage for Blog, the theme's example site for the others
5. Builds the site
6. Records the site as a draft project

The site name may use only letters, numbers, hyphens, and underscores.

**Flags:**

- `--skip-build` - Skip the build step

**Next steps after quickstart:**

1. `cd my-blog`
2. `walgo serve` - Preview locally
3. `walgo launch` - Deploy with wizard

---

### `walgo init <site-name>`

**Initialize a new Hugo site**

```bash
walgo init my-blog
walgo init my-site --quiet
```

**What it does:**

- Creates `./<site-name>` and runs `hugo new site` in it, with a TOML Hugo config
- Writes `walgo.yaml`. If the Sui CLI is on testnet or mainnet, `walrus.network` is set to that network.
- Builds the site
- Records the site as a draft project

`walgo init` installs no theme. Add one with `walgo theme install`, or start with `walgo quickstart` instead.

**Flags:**

- `-q, --quiet` - Hide most progress messages

---

## Site Management

### `walgo new [slug]`

**Create new content in Hugo site**

```bash
walgo new my-first-post
walgo new about
walgo new my-first-post --no-build
walgo new my-first-post --serve
```

**What it does:**

- Lists the content types, which are the folders under `content/`. When there is more than one, it asks which to use. The default is the type with the most files, or `posts` in an empty site.
- Asks for the slug when you leave it out. A slug may use only letters, numbers, hyphens, and underscores.
- Runs `hugo new <type>/<slug>.md`, which applies the site's archetypes
- Builds the site, unless you pass `--no-build`

**Flags:**

- `--no-build` - Skip the build after creating the file
- `--serve` - Start the Hugo development server on port 1313 after creating the file

---

### `walgo serve`

**Start local development server**

```bash
walgo serve
walgo serve -D                  # Include drafts
walgo serve --port 8080         # Custom port
walgo serve -- --bind 0.0.0.0   # Pass other flags to hugo server
```

**What it does:**

- Stops other `hugo server` processes that are running
- Builds the site
- Runs `hugo server` in the current directory

Arguments after `--` go to `hugo server` unchanged. walgo drops any other flag it does not know, so put Hugo's own flags after `--`.

**Flags:**

- `-D, --drafts` - Include draft content (`hugo server -D`)
- `-E, --expired` - Include content whose expiry date has passed (`hugo server -E`)
- `-F, --future` - Include content with a future publish date (`hugo server -F`)
- `-p, --port <int>` - Port for the server (default: `0`, which keeps Hugo's default port, 1313)

**Output:**

```
Starting local Hugo development server...
...
Web Server is available at http://localhost:1313/ (bind address 127.0.0.1)
Press Ctrl+C to stop
```

---

### `walgo theme`

**Manage Hugo themes**

Run it with a subcommand: `install`, `list`, or `new`. Each one works on the site in the current directory. `install` and `new` refuse to run outside a Hugo site.

---

### `walgo theme install <github-url>`

**Install a Hugo theme from GitHub**

```bash
walgo theme install https://github.com/theNewDynamic/gohugo-theme-ananke
walgo theme install https://github.com/alex-shpak/hugo-book
```

**What it does:**

- Accepts only `https://github.com/<owner>/<repo>` URLs
- Downloads the repository's `main` branch archive, or `master` when there is no `main`
- Replaces every theme in `themes/` with the new one. If the download fails, the previous themes are restored.
- Sets the theme in the site's TOML Hugo config, then sets up archetypes and a favicon for it

The download times out after 2 minutes. To change that, set `WALGO_THEME_DOWNLOAD_TIMEOUT` to a Go duration such as `5m`.

---

### `walgo theme list`

**List installed themes**

```bash
walgo theme list
```

Lists the folders in `themes/`.

---

### `walgo theme new <name>`

**Create a new Hugo theme scaffold**

```bash
walgo theme new my-theme
```

Runs `hugo new theme <name>` and sets the new theme in the site's TOML Hugo config. It stops if `themes/<name>` already exists.

---

## Content Management

### `walgo import <vault-path>`

**Create a new Hugo site and import content from Obsidian vault**

```bash
# Basic import - creates site with vault name
walgo import ~/Documents/MyVault

# Custom site name
walgo import ~/Notes --site-name my-knowledge-base

# Specify parent directory for site creation
walgo import ~/Vault --parent-dir ~/Projects --site-name notes-site

# With additional options
walgo import ~/Vault --attachment-dir attachments --output-dir docs
```

**What it does:**

1. **Creates a new Hugo site** with walgo.yaml configuration
2. **Initializes Hugo** project structure
3. **Imports Obsidian content:**
   - Converts notes to Hugo format
   - Handles wikilinks → markdown links
   - Copies attachments (images, files) into `static/`
   - Rewrites YAML frontmatter in the `--frontmatter-format` format, keeping only `title`, `date`, `tags`, and `aliases` and adding `draft: false`
   - Adds frontmatter to notes that have none
4. **Builds the site**
5. **Creates draft project** for tracking

**Flags:**

- `-n, --site-name <name>` - Name for the new site (default: vault directory name)
- `-p, --parent-dir <dir>` - Parent directory for site creation (default: current directory)
- `-o, --output-dir <dir>` - Subdirectory of `content/` for imported files (default: `content/` itself)
- `--attachment-dir <dir>` - Folder under `static/` for attachments (default: `obsidian.attachmentDir` from the new `walgo.yaml`, which is `images`)
- `--convert-wikilinks` - Convert `[[links]]` to markdown (default: on; `--convert-wikilinks=false` turns it off)
- `--frontmatter-format <format>` - Format of the frontmatter that import writes or rewrites: `yaml`, `toml`, or `json` (default: `yaml`)
- `--link-style <style>` - Link conversion: `markdown` or `relref` (default: `markdown`)
- `--dry-run` - Analyze the vault and print what the import would do, then remove the new site folder

**Example:**

```bash
# Import vault and create site in current directory
walgo import ~/Documents/MyVault

# Result: Creates ./MyVault/ with imported content

# Custom setup
walgo import ~/Notes \
  --site-name my-blog \
  --parent-dir ~/Projects \
  --attachment-dir images

# Result: Creates ~/Projects/my-blog/ with imported content
```

---

## Build & Optimization

### `walgo build`

**Build Hugo site with optimization**

```bash
walgo build
```

**What it does:**

1. Runs `hugo build --environment production --minify --gc --cleanDestinationDir` into the publish directory. That is `hugo.publishDir` from `walgo.yaml`, else the Hugo config's `publishDir`, else `public/`.
2. Optimizes HTML, CSS, and JS, unless `optimizer.enabled` is `false` in `walgo.yaml`
3. Removes `.DS_Store` files from the publish directory
4. Writes `ws-resources.json` when `compress.generateWSResources` is `true`, as it is in a generated `walgo.yaml`

`site.kind` in `walgo.yaml` decides whether a site is a Hugo site or a static site. When `site.kind` is unset, a site without a Hugo config is static. For a static site, walgo skips steps 1 and 2 and deploys the publish directory as it is.

`walgo build` takes no flags. To skip optimization, set `optimizer.enabled: false` in `walgo.yaml`. To leave one file type alone, set `optimizer.html.minifyHTML`, `optimizer.css.minifyCSS`, or `optimizer.js.minifyJS` to `false`. See [OPTIMIZER.md](OPTIMIZER.md). `walgo optimize` runs the optimizer on demand, but it also does nothing while `optimizer.enabled` is `false`.

**Output Example:**

```
Building site...
Building Hugo site in /Users/me/my-blog...
...
Hugo site built successfully.
Static files generated in: /Users/me/my-blog/public
Optimizing assets...

Optimization Results
Files processed: 234
Files optimized: 89
...
Generating ws-resources.json...
Generated ws-resources.json (234 resources)

Build complete!

Next steps:
  - Preview: walgo serve
  - Deploy:  walgo launch
```

---

### `walgo optimize [directory]`

**Optimize HTML, CSS, and JavaScript files**

```bash
walgo optimize public/
walgo optimize --css=false --js=false  # HTML only
walgo optimize --verbose
```

**What it does:**

- With no directory, optimizes the site's publish directory, or the current directory when there is none
- Minifies HTML, CSS, and JavaScript with tdewolff/minify, using the `optimizer` section of `walgo.yaml`, or the defaults when walgo finds no `walgo.yaml`
- Shows optimization statistics

**Flags:**

- `--html` - Turn HTML optimization on or off. Unset, `optimizer.html.enabled` decides (default: on).
- `--css` - Turn CSS optimization on or off. Unset, `optimizer.css.enabled` decides (default: on).
- `--js` - Turn JavaScript optimization on or off. Unset, `optimizer.js.enabled` decides (default: on).
- `--remove-unused-css` - Not implemented. Has no effect.
- `-v, --verbose` - Show detailed output

---

### `walgo compress [directory]`

**Compress files with Brotli**

```bash
walgo compress public/
walgo compress dist/ --level 11
walgo compress --verbose
```

**What it does:**

- With no directory, compresses the publish directory from `walgo.yaml`
- Compresses only `.html`, `.htm`, `.css`, `.js`, `.mjs`, `.json`, `.xml`, `.txt`, `.md`, `.svg`, and `.wasm` files, and skips every other file
- Writes a Brotli `.br` file next to each of those files, or replaces the file with `--in-place`
- When Brotli does not make a file smaller, the `.br` file holds the original, uncompressed bytes, and `--in-place` leaves the file unchanged
- Shows compression statistics

**Flags:**

- `-l, --level <int>` - Compression level, 0 to 11 (default: 6)
- `-v, --verbose` - Show per-file statistics
- `--in-place` - Replace the original files with their compressed versions
- `--generate-ws-resources` - Write `ws-resources.json` into the directory

---

## Deployment

### `walgo launch` (Recommended)

**Interactive deployment wizard** — The recommended way to deploy your site.

```bash
walgo launch
```

**Why use `walgo launch`?**

- Step-by-step guidance through the entire process
- Automatic project tracking and history
- Cost estimation before you confirm
- Wallet management built-in
- SuiNS domain setup instructions (mainnet)

**What it does - 8 Steps:**

The wizard builds the site first, then runs these steps.

1. **Choose Network** - testnet or mainnet. walgo switches the Sui CLI to that network for the deploy and switches it back afterwards.
2. **Wallet & Balance Check**
   - View current addresses
   - Switch address
   - Create new address
   - Import existing address
3. **Project Details** - name, category, description, and image URL
4. **Storage Duration** (epochs)
5. **Verify Site Built**
6. **Review & Confirm** (shows WAL and SUI cost estimates)
7. **Deploy to Walrus**
8. **Success!**
   - Shows Object ID
   - Shows Suiscan and Suivision links
   - SuiNS setup instructions (mainnet)

**Features:**

- Saves project to database
- Tracks deployment history
- Saves the network to `walgo.yaml` as `walrus.network`
- User-friendly prompts
- Complete wallet management
- Gas estimation
- SuiNS configuration guide ([tutorial](https://docs.wal.app/docs/walrus-sites/tutorial-suins))

---

### `walgo deploy` (Advanced)

**Direct on-chain deployment** — For automation, scripts, or advanced users.

```bash
walgo deploy --epochs 1
walgo deploy --epochs 10
walgo deploy --dry-run
walgo deploy --epochs 5 --project-name my-blog
```

**What it does:**

- Builds the site, as `walgo build` does
- Deploys the publish directory to the active Sui environment's network. If `walrus.network` in `walgo.yaml` names another network, it stops.
- Updates the existing site when it finds the site's object ID in `walrus.projectID`, in `ws-resources.json`, or in the project database. `--force-new` publishes a new site instead.
- Stops when `sui`, `walrus`, or `site-builder` is missing or older than walgo's minimum version
- Saves the object ID to `walrus.projectID` in `walgo.yaml` and to `ws-resources.json`

**Flags:**

- `-e, --epochs <int>` - Storage duration in epochs, at most 53 (default: 1)
- `-f, --force` - Continue when the publish directory does not exist yet
- `--force-new` - Publish a new site even when the site already has an object ID
- `--dry-run` - Show the deployment plan without uploading
- `--save-project` - Save the deployment as a project named after the directory
- `--project-name <name>` - Save the deployment as a project with this name (default: directory name)
- `--category <category>` - Project category (default: `website`)
- `--description <text>` - Site description for the site metadata
- `--image-url <url>` - Site image URL for the site metadata
- `--skip-version-check` - Skip the tool-update offer, and deploy even when `sui`, `walrus`, or `site-builder` is missing or too old. The problems still print as warnings.
- `--telemetry` - Record deployment metrics in `~/.walgo/metrics.json`
- `-v, --verbose` - Show detailed output for debugging
- `-q, --quiet` - Hide progress output. The object ID still prints.

**Requirements:**

- `walgo.yaml` in the current directory
- `sui`, `walrus`, and `site-builder` at walgo's minimum versions (`walgo doctor` checks them)
- A Sui wallet with SUI for gas and WAL for storage

> **Tip:** For most users, `walgo launch` provides a better experience with project tracking and guided deployment.

**Output Example:**

```
Deploying to Walrus Sites...
  [1/5] Verifying site...
  [2/5] Checking environment...
...
Deployment Successful!

Site Object ID: 0x7b5a...8f3c

View Your Object on the Sui Network
   • Suiscan:    https://suiscan.xyz/testnet/object/0x7b5a...8f3c
   • Suivision:  https://testnet.suivision.xyz/package/0x7b5a...8f3c
```

---

### `walgo deploy-http` (Testing)

**Deploy via HTTP APIs (no wallet needed)**

```bash
walgo deploy-http \
  --publisher https://publisher.walrus-testnet.walrus.space \
  --aggregator https://aggregator.walrus-testnet.walrus.space \
  --epochs 1

# One blob per file, 20 uploads at a time
walgo deploy-http --publisher <url> --aggregator <url> --mode blobs --workers 20
```

**What it does:**

- Builds the site, then uploads the publish directory to the publisher as one Quilt, or as one blob per file with `--mode blobs`
- In quilt mode, prints the Quilt ID and an aggregator URL for each file
- No blockchain interaction
- No wallet/tokens needed

The publish directory must exist before the upload starts, so run `walgo build` first.

**Flags:**

- `--publisher <url>` - Publisher base URL (required)
- `--aggregator <url>` - Aggregator base URL (required)
- `-e, --epochs <int>` - Storage duration in epochs (default: 1)
- `--mode <mode>` - `quilt` or `blobs` (default: `quilt`)
- `--workers <int>` - Concurrent uploads in `blobs` mode, at most 50 (default: 10)
- `--retries <int>` - Retries per file for transient errors (default: 5)
- `--json` - Accepted, but has no effect. The output stays plain text.
- `-v, --verbose` - Accepted, but has no effect. The output is the same with or without it.

**Limitations:**

- Data expires after `--epochs` epochs
- Creates no site object, so it cannot be updated and cannot use SuiNS domains

---

### `walgo update [object-id]`

**Update existing Walrus site**

```bash
walgo update
walgo update 0x7b5a...8f3c --epochs 5
walgo update --dry-run
```

**What it does:**

- Takes the object ID from the argument, else from `ws-resources.json` in the publish directory, else from `walrus.projectID` in `walgo.yaml`
- Builds the site and updates that site object in place, on the active Sui environment's network
- Warns before the update when the site's storage has expired, because expired files are uploaded again at full price
- Updates the project record, if the site has one

The publish directory must exist before the update starts, so run `walgo build` first.

**Flags:**

- `-e, --epochs <int>` - Storage duration in epochs (default: 1)
- `--dry-run` - Compare the publish directory with the last deployment and print the changed files, without building or uploading. Run `walgo build` first so the comparison includes your latest edits.
- `-v, --verbose` - Show a detailed change summary
- `--telemetry` - Record update metrics in `~/.walgo/metrics.json`

---

## Project Management

### `walgo projects`

**List your projects**

```bash
walgo projects
walgo projects --network mainnet
```

Without a subcommand, `walgo projects` does what `walgo projects list` does.

**Flags:**

- `-n, --network <network>` - Show only projects on this network: `testnet` or `mainnet`
- `-s, --status <status>` - Show only projects with this status: `draft`, `active`, or `archived`

---

### Project Identification

All project commands support multiple ways to identify a project:

| Method | Example | Description |
| ------ | ------- | ----------- |
| `--id=<number>` | `--id=5` | Project ID (unambiguous, recommended) |
| `--name="<name>"` | `--name="My Site"` | Project name (supports spaces) |
| `<name\|id>` | `mysite` or `5` | Positional argument (legacy, no spaces) |

**Priority:** `--id` > `--name` > positional argument

**Examples:**
```bash
# For projects with spaces in names
walgo projects show --name="My Awesome Site"
walgo projects update --name="My Blog" --epochs 10
walgo projects delete --name="Test Site"

# For unambiguous identification by ID
walgo projects show --id=5
walgo projects edit --id=5 --new-name="Renamed Site"

# Legacy syntax (no spaces)
walgo projects show mysite
```

---

### `walgo projects list`

**List all your projects, drafts included**

```bash
walgo projects list
walgo projects list --network mainnet
walgo projects list --status active
walgo projects list --status archived
```

**What it shows:**

- Project name & category
- Network (testnet/mainnet)
- Object ID
- SuiNS domain
- Deployment count
- Last deployment date
- Storage epochs and estimated expiry
- Status (draft, active, or archived)

**Flags:**

- `-n, --network <network>` - Filter by network
- `-s, --status <status>` - Filter by status: `draft`, `active`, or `archived`

---

### `walgo projects show`

**Show detailed project information**

```bash
walgo projects show --name="My Blog"      # Name with spaces
walgo projects show --id=5                # By ID
walgo projects show mysite                # Legacy syntax
```

**What it shows:**

- Complete project details
- All metadata (name, category, description, image URL)
- Network and status
- Object ID and SuiNS domain
- Wallet address
- Site path
- Storage epochs and estimated expiry, plus the on-chain expiry when the active Sui network is the project's network
- Recent deployments
- Success/failure stats
- Gas fees spent
- Site URL, when a SuiNS domain is set

**Flags:**

- `--id <int>` - Project ID
- `--name <name>` - Project name (supports spaces)

**Output Example:**

```
╔═══════════════════════════════════════════════════════════╗
║  📦 My Blog                                               ║
╚═══════════════════════════════════════════════════════════╝

ℹ️ Project Information
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ID:              1
  Name:            My Blog
  Category:        blog
  Description:     A blog about tech and decentralization
  Image URL:       https://example.com/logo.png
  Network:         mainnet
  Status:          active
  Object ID:       0x7b5a...8f3c
  SuiNS:           myblog
  URL:             https://myblog.wal.app
  Wallet:          0xabc...def
  Site path:       /Users/me/projects/my-blog
  Created:         2026-01-15 10:30:00

ℹ️ Storage & Deployment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Total Storage Epochs:  10 (~5 months)
  Expires In (est.):     13 weeks, 4 days
  Last Deploy:           2026-01-20 14:15
  Total Deployments:     5
  Successful:            5
  First Deployment:      2026-01-15 10:30
```

---

### `walgo projects update`

**Redeploy existing project with new content**

```bash
walgo projects update --name="My Blog"                # Name with spaces
walgo projects update --id=5 --epochs 10              # By ID with epochs
walgo projects update mysite                          # Legacy syntax
```

**What it does:**

- Builds the site in the project's folder
- Updates the project's site object on Walrus
- Preserves project history
- Tracks new deployment

It stops when the active Sui network is not the project's network.

**Flags:**

- `--id <int>` - Project ID (unambiguous)
- `--name "<name>"` - Project name (supports spaces)
- `-e, --epochs <int>` - Storage duration in epochs (default: `0`, which reuses the project's last value)

---

### `walgo projects edit`

**Edit project metadata locally**

```bash
walgo projects edit --id=5 --new-name="New Blog Name"
walgo projects edit --name="My Blog" --description="A blog about tech"
walgo projects edit --id=5 --category portfolio --image-url "https://example.com/logo.png"
```

**What it does (2-step process):**

1. **Database update** - Saves changes to local SQLite database
2. **ws-resources.json update** - Updates the publish directory's metadata file

After editing, use `walgo projects update` to push changes to Walrus.

**Flags:**

- `--id <int>` - Project ID (unambiguous)
- `--name "<name>"` - Project name to identify (supports spaces)
- `--new-name <name>` - New project name (rename)
- `--category <category>` - New project category
- `--description <text>` - New project description
- `--image-url <url>` - New image/logo URL for the site
- `--suins <domain>` - New SuiNS domain

**Examples:**

```bash
# Edit metadata by ID
walgo projects edit --id=5 --new-name="New Name" --category blog

# Edit by name (with spaces)
walgo projects edit --name="My Site" --description="My awesome decentralized website"

# Change image URL
walgo projects edit --id=5 --image-url "https://example.com/new-logo.png"
```

**Output Example:**

```
✏️ Editing project: My Blog

Changes to apply:
  ✏️ Name: My Blog → tech-blog
  ✏️ Description: (empty) → A blog about tech and decentralization

💾 Step 1/2: Updating database...
  ✓ Database updated
📄 Step 2/2: Updating ws-resources.json...
  ✓ ws-resources.json updated

✓ Project metadata updated locally!

💡 To push changes to Walrus, run:
   walgo projects update tech-blog
```

---

### `walgo projects archive`

**Archive a project**

```bash
walgo projects archive --name="Old Blog"    # Name with spaces
walgo projects archive --id=5               # By ID
walgo projects archive oldblog              # Legacy syntax
```

**What it does:**

- Marks project as archived
- Preserves all data

Archived projects still appear in `walgo projects list`. Add `--status active` to list only active projects, or `--status archived` to list only archived ones.

**Flags:**

- `--id <int>` - Project ID (unambiguous)
- `--name "<name>"` - Project name (supports spaces)

---

### `walgo projects delete`

**Delete a project, its site on Walrus, and its site folder**

```bash
walgo projects delete --name="Test Site"    # Name with spaces
walgo projects delete --id=5                # By ID
walgo projects delete testsite              # Legacy syntax
```

**What it does:**

- Destroys the site object on Walrus, which costs gas. If that fails, it asks whether to delete the local data anyway.
- Deletes the project from the database
- Removes deployment history
- Deletes the project's site folder. It refuses when the folder contains your home directory or another project's site folder.
- Cannot be undone

**Flags:**

- `--id <int>` - Project ID (unambiguous)
- `--name "<name>"` - Project name (supports spaces)

**Warning:** Shows the estimated gas cost and prompts for confirmation first

---

## AI Features

### `walgo ai`

**AI content generation**

The parent of the AI subcommands below. They use the provider, model, and API key that `walgo ai configure` saves in `~/.walgo/ai-credentials.yaml`.

`generate`, `update`, `pipeline`, and `resume` check and fix the frontmatter of the content they write for the site's theme.

---

### `walgo ai configure`

**Set up AI provider credentials**

```bash
walgo ai configure
```

**What it does:**

- Interactive setup wizard
- Choose provider (OpenAI/OpenRouter)
- Enter API key
- Optional custom base URL
- Choose a model (default: `gpt-4` on OpenAI, `openai/gpt-4` on OpenRouter)
- Saves to ~/.walgo/ai-credentials.yaml

**Supported Providers:**

- OpenAI (GPT-3.5, GPT-4)
- OpenRouter (Multiple models)

---

### `walgo ai get`

**Show configured AI credentials**

```bash
walgo ai get
```

Shows each configured provider with its model, its base URL, and a masked API key.

---

### `walgo ai remove [provider]`

**Remove AI credentials**

```bash
walgo ai remove            # Remove every provider
walgo ai remove openai     # Remove one provider
```

Asks for confirmation before it removes anything.

---

### `walgo ai generate`

**Generate new content with AI**

```bash
walgo ai generate
walgo ai generate --no-build
walgo ai generate --serve
```

**What it does:**

- Shows the site's content types
- Asks what to create, as one free-text instruction
- The AI picks the content type and file name, and writes Hugo markdown with frontmatter
- Fixes the frontmatter for the site's theme
- Builds the site, unless you pass `--no-build`

**Flags:**

- `--no-build` - Skip the build after generating
- `--serve` - Start the Hugo development server on port 1313 after generating

**Example Session:**

```
What content do you want to create?
...
Your instructions: Write a tutorial on deploying Hugo sites to Walrus

🔄 Generating content (AI will determine type and filename)...

✅ Content generated successfully!
   Type: posts
   File: deploying-hugo-to-walrus.md
   Path: content/posts/deploying-hugo-to-walrus.md
```

**Features:**

- Hugo-aware prompts
- SEO-optimized titles and descriptions
- Proper frontmatter structure
- Markdown formatting
- Code examples when relevant

---

### `walgo ai update <file>`

**Update existing content with AI**

```bash
walgo ai update content/posts/my-post.md
```

**What it does:**

- Asks what to change
- Sends the file and your instruction to the AI, which keeps the original structure and style
- Asks before it saves the result over the file
- Fixes the frontmatter for the site's theme, then builds the site

**Example:**

```
📄 File: content/posts/my-post.md

Describe what changes you want to make: Add a section about performance optimization

🔄 Updating content...

✅ Content updated!

Save changes to file? [Y/n]: y

✅ File updated: content/posts/my-post.md
```

---

### `walgo ai pipeline`

**Generate a complete site using AI (plan + generate)**

```bash
walgo ai pipeline
walgo ai pipeline --parallel sequential
```

**What it does:**

- Asks for a site name, a site type, a description, and the target audience
- Creates the site in `./<site-name>` with a `walgo.yaml`, a Hugo config, and the theme for the site type
- `<site-name>` is the site name in lowercase, with spaces and underscores turned into hyphens and other symbols except dots dropped. `My Site!` becomes `my-site`.
- Runs full AI content generation pipeline:
  1. Plan: AI creates a site structure plan (JSON)
  2. Generate: AI writes each page, several at a time on larger plans (see `--parallel`)
- Saves plan to `<site-name>/.walgo/plan.json` for resumability
- If interrupted, `cd` into the site and run `walgo ai resume` to continue
- Applies post-pipeline fixes and validation, builds the site, and records a draft project

**Flags:**

- `--parallel <mode>` - `auto`, `sequential`, or `parallel` (default: `auto`, which picks the number of concurrent pages from the plan size, `--concurrent`, and `--rpm`)
- `--concurrent <int>` - Most pages generated at once (default: 5)
- `--rpm <int>` - Page requests sent to the AI per minute, after a first burst of up to `--concurrent` requests (default: 30). The planning request and HTTP-level retries do not count against it.
- `--dry-run` - Generate each page with the AI, then discard it instead of writing it to `content/`. Every other step still writes to disk: the site folder, `walgo.yaml`, `hugo.toml`, the theme, the plan, the content fixes, the build, and the draft project. The plan marks the discarded pages as done, so `walgo ai resume` does not write them later.
- `-v, --verbose` - Show verbose output

**Workflow:**

```
Site name: my-tech-blog

Site type:
  1) Blog
  2) Docs
Select [1]: 1

Describe your site (1-2 sentences): A personal tech blog

Target audience: Developers and tech enthusiasts

Setting up Hugo site...

Generating content...
...
Results:
   Total pages: 5
   Completed:   5
   Skipped:     0
   Failed:      0

Site generated successfully!
```

**Supported Site Types:**

- Blog - Classic blog with posts and about page
- Docs - Documentation site

---

### `walgo ai plan`

**Create a site plan without generating content**

```bash
walgo ai plan
```

**What it does:**

- Asks the same questions as `walgo ai pipeline`
- Creates `./<site-name>` with a `walgo.yaml`, using the site name exactly as typed
- Saves the plan to `<site-name>/.walgo/plan.json`
- Does not generate any content or install a theme
- Plan can be reviewed before generation
- To generate the pages, `cd` into the site and run `walgo ai resume`
- `walgo ai resume` does not add a Hugo config or a theme either, so `walgo build` treats the folder as a static site and fails. For a site you can build, use `walgo ai pipeline`.

**Flags:**

- `-v, --verbose` - Accepted, but has no effect. Planning prints the same output with or without it.

**Example:**

```
Site name: travel-blog
...
Plan created with 6 pages:
   1. content/_index.md
   ...

Plan saved to .walgo/plan.json
Run 'walgo ai resume' to generate content.
```

```bash
cd travel-blog
cat .walgo/plan.json
walgo ai resume
```

---

### `walgo ai resume`

**Resume content generation from existing plan**

```bash
walgo ai resume
```

**What it does:**

- Loads `.walgo/plan.json` from the current directory
- Generates the pages that are not done yet
- Applies post-pipeline fixes and validation
- Creates draft project entry

**Flags:**

- `--parallel <mode>` - `auto`, `sequential`, or `parallel` (default: `auto`)
- `--concurrent <int>` - Most pages generated at once (default: 5)
- `--rpm <int>` - Page requests sent to the AI per minute, after a first burst of up to `--concurrent` requests (default: 30). HTTP-level retries do not count against it.
- `--dry-run` - Generate each page with the AI, then discard it instead of writing it to `content/`. walgo still saves the plan with those pages marked done, runs the menu update and the content fixes, and records the draft project.
- `-v, --verbose` - Show verbose output

**Use Case:**
When `walgo ai pipeline` was interrupted (e.g., user stopped it, network issue), you can resume from where it left off.

**Example:**

```
Resuming AI Generation

✓ Using openai (gpt-4)
...
Results:
   Total pages: 5
   Completed:   5
   Skipped:     0
   Failed:      0

Validating and fixing content for theme...
   ✓ Content validated and fixed
   ✓ Created draft project

Site generated successfully!
```

---

## Desktop App

### `walgo desktop`

**Launch Walgo desktop application**

```bash
walgo desktop
```

**What it does:**

- Launches the installed Walgo Desktop GUI. walgo looks for it in `/Applications` and `~/Applications` on macOS, in `%LOCALAPPDATA%\Programs\Walgo` on Windows, and in `~/.local/bin` and `/usr/local/bin` on Linux.
- Provides graphical interface for all features:
  - Site Management (Create, Build, Deploy)
  - AI Content Generation (Configure, Generate, Update, Pipeline)
  - Project Management (List, Edit, Delete, Archive)
  - Wallet Integration (Network switching, Account management)
  - System Health (Dependency checking, Auto-install)
- Easier project management
- Visual feedback and progress tracking

**Platform Support:**

- macOS (Universal binary)
- Windows (amd64)
- Linux (amd64)

**Desktop App Features:**

#### Dashboard

- Wallet overview with SUI/WAL balances
- Network switching (testnet/mainnet)
- Account management (create, import, switch)
- Quick action cards for common tasks

#### Create

- QuickStart - Instant site creation
- AI Pipeline - Full AI site creation
- Init Site - Manual Hugo site initialization
- Obsidian Import - Import from vaults

#### Projects

- List all deployed sites
- Search and filter projects
- View project details and deployment history
- Edit project metadata
- Delete/archive projects
- Open project folders

#### AI Config

- Configure AI providers (OpenAI/OpenRouter)
- Manage API keys
- Set custom base URLs
- Model selection

#### Edit

- Edit site content
- Build sites
- Deploy to Walrus

#### System Health

- Check dependency status (Hugo, Sui, Walrus, Site Builder)
- Auto-install missing dependencies
- View network connectivity

See [DESKTOP_INTEGRATION.md](DESKTOP_INTEGRATION.md) for complete documentation.

---

## Setup & Configuration

### `walgo setup [network]`

**Configure site-builder for Walrus Sites**

```bash
walgo setup
walgo setup --network mainnet --force
walgo setup mainnet --force
```

**What it does:**

- Downloads the Walrus client config to `~/.config/walrus/client_config.yaml` when that file is missing, or always with `--force`
- Downloads `~/.config/walrus/sites-config.yaml` from the walrus-sites repository, or writes it from a template when the download fails
- For mainnet, sets the default context of both files to mainnet
- Lists any missing `sui`, `walrus`, or `site-builder` binary with install steps

If `site-builder` and its config are already in place, `walgo setup` changes nothing unless you pass `--force`.

**Flags:**

- `-n, --network <network>` - `testnet` or `mainnet` (default: `testnet`). A network given as an argument takes precedence.
- `--force` - Overwrite existing config

---

### `walgo setup-deps`

**Download and install dependencies**

```bash
walgo setup-deps
walgo setup-deps --network mainnet
walgo setup-deps --with-hugo
walgo setup-deps --with-sui=false
walgo setup-deps --legacy
```

**What it does:**

- Installs `suiup` if it is missing
- Runs `suiup install` for `sui` and `walrus` on the chosen network, and for `site-builder@mainnet`
- Prints the tool versions that your `PATH` resolves afterwards
- With `--legacy`, downloads `site-builder` and `walrus` straight into `--bin-dir` and points `sites-config.yaml` at that `walrus`. The legacy method does not install `sui`.

**Flags:**

- `--network <network>` - `testnet` or `mainnet` (default: `testnet`)
- `--with-sui` - Install the `sui` CLI (default: on; `--with-sui=false` skips it)
- `--with-walrus` - Install the `walrus` client (default: on)
- `--with-site-builder` - Install `site-builder` (default: on)
- `--with-hugo` - Check for Hugo and print install steps when it is missing. walgo does not install Hugo.
- `--legacy` - Download the binaries directly instead of using suiup
- `--bin-dir <dir>` - Install directory for `--legacy` (default: `~/.config/walgo/bin`)

---

## Diagnostics & Utilities

### `walgo doctor`

**Diagnose environment and configuration**

```bash
walgo doctor
walgo doctor --fix-paths
walgo doctor --verbose
```

**What it checks:**

- `suiup`, Hugo (Extended), `site-builder`, `walrus`, and `sui` on your `PATH`
- The `sui`, `walrus`, and `site-builder` versions against walgo's minimums
- Active Sui network and address
- SUI and WAL balances
- `~/.config/walrus/sites-config.yaml`, including `~` paths in it
- `walgo.yaml` in the current directory

**Flags:**

- `--fix-paths` - Rewrite `~` paths in `sites-config.yaml` as absolute paths
- `--fix-all` - Apply every automatic fix. Today that is the same rewrite `--fix-paths` does.
- `-v, --verbose` - Show detailed diagnostics, including versions and paths

**Output Example:**

```
Checking dependencies...

  ✓ suiup found (Sui tool manager)
  ✓ hugo found
  ✓ site-builder found
  ✓ walrus found
  ✓ sui found

  ✓ sui v1.75.0
  ✓ walrus v1.52.0
  ✓ site-builder v2.12.0

Checking Sui configuration...

  ✓ Active network: testnet
  ✓ Active address: 0xabc...def
  ✓ SUI balance: 5.00 SUI
  ✓ WAL balance: 2.00 WAL

Checking configuration files...

  ✓ sites-config.yaml found at /Users/me/.config/walrus/sites-config.yaml
  ✓ walgo.yaml found in current directory
...
All checks passed! Your environment is ready.
```

---

### `walgo status [object-id]`

**Check status of deployed Walrus site**

```bash
walgo status
walgo status 0x7b5a...8f3c
```

**What it shows:**

- Object ID
- Number of resources (files), from `site-builder sitemap`
- When the storage expires, or how many resources have expired

Without an argument, it uses `walrus.projectID` from `walgo.yaml`.

---

### `walgo domain [domain-name]`

**Get SuiNS domain configuration instructions**

```bash
walgo domain
walgo domain myblog
```

**What it shows:**

- On mainnet, the steps to link a SuiNS name to the site at https://suins.io, with the site's object ID from `walrus.projectID`
- With a domain name, the address the site will have: `https://<domain>.wal.app`
- On testnet, a note that SuiNS works only on mainnet

The network is the one the project was deployed on, else `walrus.network`, else the active Sui environment. You link the name in the SuiNS web app. walgo prints the steps.

---

### `walgo version`

**Show version information**

```bash
walgo version
walgo version --short
walgo version --check-updates
```

**Output:**

```
Walgo v0.5.0
Commit:  6c0da69
Built:   2026-10-16

This is a beta release. Please report issues at:
   https://github.com/ganbitlabs/walgo/issues
```

**Flags:**

- `--short` - Print only the version number
- `--check-updates` - Check GitHub for a newer release

---

### `walgo uninstall`

**Uninstall Walgo CLI and/or desktop app**

```bash
walgo uninstall
walgo uninstall --all --force
walgo uninstall --cli --force
walgo uninstall --desktop --force
walgo uninstall --all --keep-cache --force
```

**What it does:**

- Interactive uninstall wizard
- Removes CLI binary
- Removes desktop app
- Removes cache and data, unless you pass `--keep-cache`
- **NEVER** deletes Sui wallet or blockchain data

**Flags:**

- `-a, --all` - Uninstall both CLI and desktop app
- `--cli` - Uninstall CLI only
- `--desktop` - Uninstall desktop app only
- `-f, --force` - Skip confirmation prompts
- `--keep-cache` - Keep cache and data files

**What Gets Removed:**

- Walgo CLI binary
- Walgo Desktop app
- Cache and data in `~/.walgo` and `~/.config/walgo` (unless --keep-cache)

**What Stays Safe:**

- Your SUI balance (on blockchain)
- Deployed sites (on Walrus)
- Sui wallet (~/.sui/)
- All on-chain data

See [UNINSTALL.md](UNINSTALL.md) for details.

---

### `walgo completion`

**Generate shell autocompletion**

```bash
# Bash
walgo completion bash > /etc/bash_completion.d/walgo

# Zsh
walgo completion zsh > "${fpath[1]}/_walgo"

# Fish
walgo completion fish > ~/.config/fish/completions/walgo.fish

# PowerShell
walgo completion powershell > walgo.ps1
```

**Supported Shells:**

- bash
- zsh
- fish
- powershell

Each shell subcommand accepts `--no-descriptions` to leave out completion descriptions.

---

## Common Workflows

### Complete First Deploy (Recommended)

```bash
# 1. Create your site
walgo quickstart my-site
cd my-site

# 2. Preview locally
walgo serve

# 3. Deploy with the interactive wizard
walgo launch
# The wizard guides you through:
# - Network selection (testnet/mainnet)
# - Wallet setup
# - Project naming
# - Storage duration
# - Cost confirmation
# - Deployment

# 4. Configure SuiNS (post-deployment, mainnet only)
# Follow the instructions shown after deployment
# to link your domain at suins.io
```

---

### AI-Powered Content Creation

```bash
# 1. Configure AI
walgo ai configure

# 2. Generate content
walgo ai generate

# 3. Review and build
cat content/posts/your-new-post.md
walgo build

# 4. Deploy
walgo launch
```

---

### Update Existing Site

```bash
# 1. Make changes
cd my-site
walgo ai update content/posts/old-post.md
# OR edit manually

# 2. Build
walgo build

# 3. Update project (use --name for names with spaces, or --id)
walgo projects update --name="My Site"
# or
walgo projects update --id=5
```

---

### Project Management

```bash
# List all projects
walgo projects list

# Show details (various ways to identify project)
walgo projects show --name="My Blog"      # Name with spaces
walgo projects show --id=5                # By ID (unambiguous)
walgo projects show mysite                # Legacy syntax (no spaces)

# Update site content
walgo projects update --name="My Blog"
walgo projects update --id=5 --epochs 10  # Update with new epochs

# Edit metadata locally
walgo projects edit --id=5 --new-name="New Name" --description="Updated description"

# Then push to Walrus
walgo projects update --id=5

# Archive old project
walgo projects archive --name="Old Site"

# Delete a project: destroys its site on Walrus and deletes its folder
walgo projects delete --id=5
```

### Metadata Management

```bash
# Rename project (use --new-name flag)
walgo projects edit --id=5 --new-name="My Awesome Site"

# Update description for wallet/explorer display
walgo projects edit --name="My Site" --description="A decentralized blog about Web3"

# Update category
walgo projects edit --id=5 --category portfolio

# Update site logo/image
walgo projects edit --id=5 --image-url="https://example.com/logo.png"

# Update multiple fields at once
walgo projects edit --id=5 \
  --new-name="New Name" \
  --description="New description" \
  --category blog \
  --image-url="https://example.com/logo.png"

# Push metadata changes to Walrus
walgo projects update --id=5
```

---

### Obsidian to Blog

```bash
# 1. Import Obsidian vault (creates site automatically)
walgo import ~/Documents/MyVault --site-name my-knowledge-base

# 2. Navigate to created site
cd my-knowledge-base

# 3. Build and deploy
walgo build
walgo launch
```

---

### Choosing the Network

```bash
# walgo deploy uses the active Sui environment
sui client switch --env mainnet
walgo deploy --epochs 10
```

`walgo deploy`, `walgo update`, and `walgo projects update` deploy to the active Sui environment. If `walrus.network` in `walgo.yaml` names another network, `walgo deploy` and `walgo update` stop with a mismatch error. `walgo projects update` checks the network recorded for the project instead, and stops when it is not the active one. `walgo launch` asks for the network, switches the Sui CLI for the deploy, and offers to publish a new site when the existing one lives on the other network.

---

## Global Flags

Available for all commands:

| Flag               | Description                                                      |
| ------------------ | ---------------------------------------------------------------- |
| `--config <path>`  | Config file to load (default: ./walgo.yaml or ~/.walgo.yaml)     |
| `--help` / `-h`    | Show help for command                                            |

walgo loads the `--config` file and prints `Using config file: <path>`. Commands read their settings from `walgo.yaml` in the site directory, so the file named by `--config` does not change what they do.

`walgo` itself, with no command, also accepts `-t, --toggle`. It has no effect.

---

## Quick Command Reference

**Init & Build:**

- `quickstart` - One-command setup (recommended for new users)
- `init` - Create new Hugo site
- `build` - Build site with optimization
- `serve` - Local development server
- `theme` - Install, list, or create Hugo themes

**Content:**

- `new` - Create new content
- `import` - Import from Obsidian
- `ai generate` - AI content generation
- `ai update` - AI content updates
- `ai pipeline` - Generate a complete site with AI
- `ai plan` - Create an AI site plan
- `ai resume` - Continue generating from a plan

**Deploy:**

- `launch` - Interactive deployment wizard (**recommended**)
- `projects update` - Update existing project
- `deploy` - Direct on-chain deployment (advanced)
- `deploy-http` - HTTP deployment for testing (no wallet)
- `update` - Update site by object ID (advanced)

**Projects:**

- `projects` - List all projects (same as `projects list`)
- `projects list` - List all projects
- `projects show` - Show project details (use `--name="..."` or `--id=N`)
- `projects update` - Update project on-chain (use `--name="..."` or `--id=N`)
- `projects edit` - Edit project metadata locally (use `--new-name` to rename)
- `projects archive` - Archive project (use `--name="..."` or `--id=N`)
- `projects delete` - Delete project, its on-chain site, and its folder (use `--name="..."` or `--id=N`)

**Optimize:**

- `optimize` - Optimize assets
- `compress` - Brotli compression

**Setup:**

- `setup` - Write the site-builder configuration
- `setup-deps` - Install dependencies
- `ai configure` - Configure AI provider
- `ai get` - Show AI credentials
- `ai remove` - Remove AI credentials

**Diagnostics:**

- `doctor` - System diagnostics
- `status` - Check deployment status
- `domain` - SuiNS domain setup instructions
- `version` - Show version
- `uninstall` - Uninstall Walgo

**Desktop:**

- `desktop` - Launch desktop app

---

## Resources

- **GitHub**: https://github.com/ganbitlabs/walgo
- **SuiNS**: https://suins.io (mainnet) | https://testnet.suins.io (testnet)
- **Walrus Docs**: https://docs.walrus.site
- **Hugo Docs**: https://gohugo.io/documentation/

---

## Getting Help

For any command, use `--help`:

```bash
walgo --help
walgo init --help
walgo deploy --help
walgo projects --help
```

---

**Total Commands: 24, plus 21 subcommands**

Last updated: v0.5.0

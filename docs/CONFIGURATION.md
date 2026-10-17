# Configuration Reference

Complete reference for configuring Walgo.

## Table of Contents

- [Configuration File](#configuration-file)
- [Site Configuration](#site-configuration)
- [Hugo Configuration](#hugo-configuration)
- [Walrus Configuration](#walrus-configuration)
- [Optimizer Configuration](#optimizer-configuration)
- [Compression Configuration](#compression-configuration)
- [Cache Configuration](#cache-configuration)
- [Obsidian Configuration](#obsidian-configuration)
- [Environment Variables](#environment-variables)
- [Configuration Examples](#configuration-examples)

## Configuration File

Walgo reads its settings from a YAML file named `walgo.yaml` in the site directory.

### File Locations

Commands that need site settings read `walgo.yaml` from the site directory. That is the current directory, with these exceptions:

- `walgo projects update` and `walgo projects edit` read it from the project's site folder.
- `walgo optimize <directory>` and `walgo compress <directory> --generate-ws-resources` read it from `<directory>`.

No other file supplies settings. The global `--config` flag, `./walgo.yml`, and `~/.walgo.yaml` only make walgo print `Using config file: <path>`. No command takes settings from them.

If `walgo.yaml` is missing, commands that need it stop with `walgo.yaml not found in <directory>`. `walgo optimize` uses the optimizer defaults instead.

`walgo init` and the other commands that create a site write a `walgo.yaml` with the defaults. A key that `walgo.yaml` leaves out takes the default listed for it below.

### Basic Structure

```yaml
# Site kind: hugo or static. Leave unset to detect it.
site:
  kind: ""

# Hugo site configuration
hugo:
  version: ""
  baseURL: ""
  publishDir: "public"
  contentDir: "content"
  resourceDir: "resources"

# Walrus deployment configuration
walrus:
  projectID: ""
  bucketName: ""
  entrypoint: "index.html"
  network: ""  # walgo init sets the active Sui network
  suinsDomain: ""

# Asset optimizer configuration
optimizer:
  enabled: true
  verbose: false
  outputDir: ""
  html:
    enabled: true
    minifyHTML: true
    removeComments: true
    removeWhitespace: true
    compressInlineCSS: true
    compressInlineJS: true
  css:
    enabled: true
    minifyCSS: true
    removeComments: true
    removeUnused: false
    autoprefixer: false
    compressColors: true
  js:
    enabled: true
    minifyJS: true
    removeComments: true
    obfuscate: false
    sourceMaps: false
  skipPatterns:
    - "*.min.js"
    - "*.min.css"
    - "*.min.html"
    - "*/.git/*"
    - "*/node_modules/*"

# ws-resources.json generation
compress:
  enabled: false
  level: 6
  generateWSResources: true
  customRoutes: {}
  ignorePatterns: []

# Cache-Control headers in ws-resources.json
cache:
  enabled: true
  immutableMaxAge: 31536000
  mutableMaxAge: 300

# Obsidian import configuration
obsidian:
  vaultPath: ""
  attachmentDir: "images"
  convertWikilinks: true
  includeDrafts: false
  frontmatterFormat: "yaml"
  linkStyle: "markdown"
```

## Site Configuration

### `site.kind`

- **Type:** String, `hugo` or `static`
- **Default:** unset (detect)
- **Description:** Says whether walgo builds the site with Hugo or deploys its publish directory as-is

When `site.kind` is unset, walgo treats a site as Hugo when it finds a Hugo config: `hugo.toml`, `hugo.yaml`, `hugo.yml`, `hugo.json`, `config.toml`, or any `hugo.*` or `config.*` under `config/_default/` or `config/production/`. A root `config.yaml`, `config.yml`, or `config.json` counts only next to `archetypes/` or Markdown under `content/`, because those names are common in non-Hugo projects. Anything else is a static site.

Set `site.kind` when detection gets it wrong:

```yaml
site:
  kind: static  # Never run hugo; deploy hugo.publishDir as-is
```

```yaml
site:
  kind: hugo    # Always run hugo, and read a root config.yaml as Hugo config
```

A static site skips Hugo and the optimizer. Any other value is an error when walgo loads `walgo.yaml`.

## Hugo Configuration

Tells walgo where the Hugo site's files are. Hugo reads everything else from its own config file, such as `hugo.toml`.

`walgo build` runs `hugo build --environment production --minify --gc --cleanDestinationDir --destination <publishDir>`. Hugo always minifies its output. Draft pages are left out unless the Hugo config sets `buildDrafts = true`.

### `hugo.version`

- **Type:** String
- **Default:** `""`
- **Description:** Not implemented. Walgo accepts the key and ignores it. Walgo runs whichever `hugo` it finds installed.

### `hugo.baseURL`

- **Type:** String
- **Default:** `""`
- **Description:** Not implemented. Walgo accepts the key and ignores it. Set `baseURL` in the Hugo config instead.

### `hugo.publishDir`

- **Type:** String
- **Default:** the `publishDir` in the site's Hugo config, else `"public"`. A static site defaults to `"public"`.
- **Description:** Output directory for the built site, and the directory walgo deploys

A relative path is relative to the site directory. An absolute path is used as-is. walgo refuses a publish directory that contains the site itself, such as `.`.

```yaml
hugo:
  publishDir: "dist"  # Use dist/ instead of public/
```

### `hugo.contentDir`

- **Type:** String
- **Default:** `"content"`
- **Description:** Content directory where `walgo ai generate` looks for sections and saves the new page. `walgo ai pipeline`, `walgo ai resume`, and the AI content fixer always use `content/`.

Hugo takes its content directory from the Hugo config, so set `contentDir` there to the same value.

```yaml
hugo:
  contentDir: "posts"  # Content in posts/ instead
```

### `hugo.resourceDir`

- **Type:** String
- **Default:** `"resources"`
- **Description:** Not implemented. Walgo accepts the key and ignores it. Hugo takes `resourceDir` from the Hugo config.

## Walrus Configuration

Controls Walrus deployment behavior.

Storage duration is not a `walgo.yaml` key. Pass `--epochs` to `walgo deploy`, `walgo deploy-http`, or `walgo update`, where it defaults to `1`. `walgo projects update --epochs` defaults to the project's previous value. `walgo launch` asks for it. One epoch is 1 day on testnet and 2 weeks on mainnet.

### `walrus.projectID`

- **Type:** String
- **Default:** `""`. `walgo init` writes the placeholder `YOUR_WALRUS_PROJECT_ID`, which walgo treats as unset.
- **Description:** Object ID of the site on Walrus

```yaml
walrus:
  projectID: "0x5c1d...e9a2"
```

**Auto-set:** Walgo writes the object ID here after each on-chain deployment.

`walgo deploy` and `walgo launch` update the site with this ID instead of publishing a new one. `walgo status` uses it when you pass no object ID. `walgo update` uses it when you pass no object ID and `ws-resources.json` has none. `walgo domain` shows it as the object ID to link.

### `walrus.bucketName`

- **Type:** String
- **Default:** `""`
- **Description:** Not implemented. Walgo accepts the key and ignores it.

### `walrus.entrypoint`

- **Type:** String
- **Default:** `"index.html"`
- **Description:** Not implemented. Walgo accepts the key and ignores it.

### `walrus.network`

- **Type:** String
- **Default:** unset. `walgo init` sets it to the active Sui environment when that is `testnet` or `mainnet`.
- **Options:** `"testnet"`, `"mainnet"`
- **Description:** The network the site is deployed on

```yaml
walrus:
  network: "mainnet"  # Use mainnet (requires real SUI)
```

Deployments always go to the active Sui environment. Switch it with `sui client switch --env <network>`. If `walrus.network` is set to a different network, `walgo deploy` and `walgo update` stop with a network mismatch error. `walgo launch` writes the network you pick to this key.

`walgo setup --network` writes the site-builder config, `~/.config/walrus/sites-config.yaml`. It does not change this key.

**Warning:** Mainnet requires real SUI tokens with monetary value!

### `walrus.suinsDomain`

- **Type:** String
- **Default:** `""`
- **Description:** Not implemented. Walgo accepts the key and ignores it. To point a SuiNS name at the site, see `walgo domain`.

## Optimizer Configuration

Controls asset optimization behavior. An optimizer changes a file only when both its `enabled` key and its minify flag are true. The [Optimizer Documentation](OPTIMIZER.md) explains when the optimizer runs.

### Global Optimizer Settings

#### `optimizer.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Enable/disable all optimization, for builds and for `walgo optimize`

```yaml
optimizer:
  enabled: false  # Disable all optimization
```

#### `optimizer.verbose`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Print an error line for each file that fails to optimize

```yaml
optimizer:
  verbose: true  # Print files that fail to optimize
```

#### `optimizer.outputDir`

- **Type:** String
- **Default:** `""`
- **Description:** Not implemented. Walgo accepts the key and ignores it. The optimizer rewrites files in place in the publish directory.

#### `optimizer.skipPatterns`

- **Type:** Array of strings
- **Default:** `["*.min.js", "*.min.css", "*.min.html", "*/.git/*", "*/node_modules/*"]`
- **Description:** File patterns to skip

A pattern without `/` matches the file name. A pattern with `/` also matches the full path.

```yaml
optimizer:
  skipPatterns:
    - "*.min.js"
    - "*.min.css"
    - "third-party.js"
    - "sw.js"
```

Walgo always skips files whose name contains `.min.`, so already minified files need no pattern.

### HTML Optimizer

#### `optimizer.html.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Enable HTML optimization

```yaml
optimizer:
  html:
    enabled: false  # Skip HTML files
```

#### `optimizer.html.minifyHTML`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify HTML. Every other HTML key needs it.

```yaml
optimizer:
  html:
    minifyHTML: true  # Minify HTML
```

#### `optimizer.html.removeComments`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Remove HTML comments

```yaml
optimizer:
  html:
    removeComments: true
```

**Note:** Preserves conditional comments (e.g., IE-specific).

#### `optimizer.html.removeWhitespace`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Collapse whitespace between elements

```yaml
optimizer:
  html:
    removeWhitespace: false  # Preserve whitespace (safer)
```

**Note:** The content of `<pre>` and `<textarea>` is always kept.

#### `optimizer.html.compressInlineCSS`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify CSS in `<style>` tags

```yaml
optimizer:
  html:
    compressInlineCSS: true
```

#### `optimizer.html.compressInlineJS`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify JavaScript in `<script>` tags. Variable names are kept.

```yaml
optimizer:
  html:
    compressInlineJS: true
```

### CSS Optimizer

#### `optimizer.css.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Enable CSS optimization

```yaml
optimizer:
  css:
    enabled: true
```

#### `optimizer.css.minifyCSS`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify CSS files. Minification strips every comment except `/*! ... */` comments, and shortens colors and numbers.

```yaml
optimizer:
  css:
    minifyCSS: true
```

**Examples:**
- `#ffffff` → `#fff`
- `rgb(0, 0, 0)` → `#000`

#### `optimizer.css.removeComments`

- **Type:** Boolean
- **Default:** `true`
- **Description:** No separate effect. CSS minification already strips every comment except `/*! ... */` comments.

#### `optimizer.css.removeUnused`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Not implemented. Walgo accepts the key and ignores it. The `walgo optimize --remove-unused-css` flag is ignored too.

#### `optimizer.css.autoprefixer`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Not implemented. Walgo accepts the key and ignores it.

#### `optimizer.css.compressColors`

- **Type:** Boolean
- **Default:** `true`
- **Description:** No separate effect. CSS minification already shortens colors.

### JavaScript Optimizer

#### `optimizer.js.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Enable JavaScript optimization

```yaml
optimizer:
  js:
    enabled: true
```

#### `optimizer.js.minifyJS`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify JavaScript files. Minification strips comments.

```yaml
optimizer:
  js:
    minifyJS: true
```

**Note:** Preserves license comments (`/*! ... */`).

#### `optimizer.js.removeComments`

- **Type:** Boolean
- **Default:** `true`
- **Description:** No separate effect. JavaScript minification already strips comments.

#### `optimizer.js.obfuscate`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Let the minifier shorten local variable names

```yaml
optimizer:
  js:
    obfuscate: true  # Shorten local variable names
```

**Warning:** If a script breaks after a build, set `obfuscate: false`.

#### `optimizer.js.sourceMaps`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Not implemented. Walgo writes no source maps.

## Compression Configuration

Controls `ws-resources.json`, the Walrus Sites file that sets each file's headers, the site's routes, and the files site-builder skips. Walgo does not Brotli-compress files during a build or a deployment. Run `walgo compress` for that.

### `compress.enabled`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Not implemented. Walgo accepts the key and ignores it.

### `compress.level`

- **Type:** Integer, 0-11
- **Default:** `6`
- **Description:** Not implemented. Walgo accepts the key and ignores it. Set the Brotli level with `walgo compress --level`.

### `compress.generateWSResources`

- **Type:** Boolean
- **Default:** `true` in the `walgo.yaml` that `walgo init` writes. `false` when the key is missing.
- **Description:** Write `ws-resources.json` to the publish directory after each build

```yaml
compress:
  generateWSResources: true
```

`ws-resources.json` gets a `Content-Type` header for each file, the [cache headers](#cache-configuration), a route for each `index.html`, the ignore list, and site metadata. A `404.html` in the publish directory becomes the fallback route.

Walgo writes a new `ws-resources.json` on each build. It copies the object ID of an earlier deployment from the project database, or else from the old file. When the site has a `static/` folder, the Hugo build deletes the old file first, so the object ID carries over only for a site in `walgo projects`. `walgo deploy` reads the object ID from `walrus.projectID` in `walgo.yaml` first.

### `compress.customRoutes`

- **Type:** Map of route pattern to file path
- **Default:** `{}`
- **Description:** Extra routes for `ws-resources.json`

**Status:** Has no effect. Walgo replaces the routes in `ws-resources.json` with the routes it generates from the publish directory, which drops these entries.

### `compress.ignorePatterns`

- **Type:** Array of strings
- **Default:** `[]`
- **Description:** Paths to add to the `ignore` list in `ws-resources.json`. site-builder does not upload matching files.

```yaml
compress:
  ignorePatterns:
    - "/drafts/*"
```

Walgo always ignores `/.DS_Store`, `/Thumbs.db`, `/.gitkeep`, `/.gitignore`, `/.git/*`, `/.well-known/*`, and `.map` files in `/`, `/js/`, `/css/`, and `/assets/`.

## Cache Configuration

Sets the `Cache-Control` headers that walgo writes into `ws-resources.json`. Walgo uses them whenever it writes that file, both in `walgo build` and in `walgo compress --generate-ws-resources`.

### `cache.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Add a `Cache-Control` header for each file

```yaml
cache:
  enabled: false  # No Cache-Control headers
```

JavaScript and CSS files with a second dot in the name, such as `app.3f2a.js`, and font files get `public, max-age=<immutableMaxAge>, immutable`. HTML files get `public, max-age=<mutableMaxAge>, must-revalidate`. Every other file gets `public, max-age=<mutableMaxAge>`.

### `cache.immutableMaxAge`

- **Type:** Integer (seconds)
- **Default:** `31536000` (one year)
- **Description:** `max-age` for those JavaScript, CSS, and font files. `0` keeps the default.

```yaml
cache:
  immutableMaxAge: 2592000  # 30 days
```

### `cache.mutableMaxAge`

- **Type:** Integer (seconds)
- **Default:** `300` (5 minutes)
- **Description:** `max-age` for HTML and every other file. `0` keeps the default.

```yaml
cache:
  mutableMaxAge: 60  # 1 minute
```

## Obsidian Configuration

Controls how `walgo import` converts an Obsidian vault. `walgo import` creates a new site and writes a new `walgo.yaml` before it imports, so an import always starts from the defaults below. To change a setting for an import, pass the flag listed with the key.

### `obsidian.vaultPath`

- **Type:** String
- **Default:** `""`
- **Description:** Not implemented. Walgo accepts the key and ignores it. The vault path is the argument of `walgo import`.

### `obsidian.attachmentDir`

- **Type:** String
- **Default:** `"images"`
- **Flag:** `--attachment-dir`
- **Description:** Directory under `static/` for imported attachments

```yaml
obsidian:
  attachmentDir: "attachments"  # static/attachments/
```

### `obsidian.convertWikilinks`

- **Type:** Boolean
- **Default:** `true`
- **Flag:** `--convert-wikilinks`
- **Description:** Convert `[[links]]` to markdown links

```yaml
obsidian:
  convertWikilinks: true
```

**Example conversion:**
```markdown
# Before
[[My Page]]
[[My Page|Custom Text]]

# After
[My Page](my-page.md)
[Custom Text](my-page.md)
```

### `obsidian.includeDrafts`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Not implemented. Walgo accepts the key and ignores it.

### `obsidian.frontmatterFormat`

- **Type:** String
- **Default:** `"yaml"`
- **Options:** `"yaml"`, `"toml"`, `"json"`
- **Flag:** `--frontmatter-format`
- **Description:** Frontmatter format for imported notes

```yaml
obsidian:
  frontmatterFormat: "toml"  # Use TOML frontmatter
```

### `obsidian.linkStyle`

- **Type:** String
- **Default:** `"markdown"`
- **Options:** `"markdown"`, `"relref"`
- **Flag:** `--link-style`
- **Description:** How converted wikilinks are written

`markdown` writes `[My Page](my-page.md)`. `relref` writes `[My Page]({{< relref "my-page.md" >}})`, and Hugo fails the build when the target page does not exist. Any other value acts as `markdown`.

```yaml
obsidian:
  linkStyle: "relref"  # Use Hugo relref shortcodes
```

## Environment Variables

Walgo does not read `walgo.yaml` keys from the environment. It reads these variables:

| Variable | Effect |
|---|---|
| `WALGO_ASCII` | `1` or `true` prints ASCII instead of emoji. It wins over `WALGO_EMOJI`. |
| `WALGO_EMOJI` | `1` or `true` prints emoji, even in CI or when `TERM` is `dumb` or unset. |
| `WALGO_THEME_DOWNLOAD_TIMEOUT` | Timeout for downloading a theme from GitHub, as a Go duration such as `5m`. The default is `2m`. |

## Configuration Examples

### Minimal Configuration

```yaml
# walgo.yaml - Minimal setup
compress:
  generateWSResources: true
```

Everything else uses defaults. Without `generateWSResources: true`, `walgo build` writes no `ws-resources.json`.

### Development Configuration

```yaml
# walgo.yaml - Development setup
optimizer:
  enabled: false         # Disable optimization for speed

walrus:
  network: "testnet"
```

To build draft pages, set `buildDrafts = true` in the Hugo config. For short-lived test deployments, run `walgo deploy --epochs 1`.

### Production Configuration

```yaml
# walgo.yaml - Production setup
optimizer:
  enabled: true
  verbose: false
  html:
    enabled: true
    minifyHTML: true
    removeComments: true
  css:
    enabled: true
    minifyCSS: true
    compressColors: true
  js:
    enabled: true
    minifyJS: true
    removeComments: true
    obfuscate: false       # Don't risk breaking code

compress:
  generateWSResources: true

walrus:
  network: "testnet"
```

Set the storage duration when you deploy, for example `walgo deploy --epochs 10`.

### Aggressive Optimization

```yaml
# walgo.yaml - Maximum optimization
optimizer:
  enabled: true
  verbose: true

  html:
    enabled: true
    minifyHTML: true
    removeComments: true
    removeWhitespace: true
    compressInlineCSS: true
    compressInlineJS: true

  css:
    enabled: true
    minifyCSS: true
    removeComments: true
    compressColors: true

  js:
    enabled: true
    minifyJS: true
    removeComments: true
    obfuscate: false        # Still risky

  skipPatterns:
    - "sw.js"
```

**Warning:** Test thoroughly! Aggressive optimization can break sites.

### Obsidian-Focused Configuration

`walgo import` writes a new `walgo.yaml`, so set the import options with flags:

```bash
walgo import ~/Documents/MyVault \
  --attachment-dir images \
  --frontmatter-format yaml \
  --link-style markdown
```

## Configuration Validation

### Checking Configuration

```bash
# Check that walgo.yaml exists in the current directory
walgo doctor
```

`walgo doctor` does not print configuration values. Walgo checks `walgo.yaml` each time a command loads it and stops on the errors below. `walgo optimize` is the exception. When it cannot read `walgo.yaml`, it runs with the default optimizer settings, and without a directory argument it optimizes the current directory instead of the publish directory.

### Common Validation Errors

**Invalid YAML syntax:**
```
Error: build failed: failed to parse config file /path/to/site/walgo.yaml: yaml: line 2: did not find expected key
```

**Solution:** Check YAML indentation and syntax.

**Unknown configuration key:**

Walgo ignores a key it does not recognize and prints no warning. A misspelled key, such as `optimiser`, has no effect.

**Solution:** Check key names against this reference.

**Invalid value type:**
```
Error: build failed: failed to parse config file /path/to/site/walgo.yaml: yaml: unmarshal errors:
  line 2: cannot unmarshal !!str `maybe` into bool
```

**Solution:** Use proper types (`true` or `false` for booleans).

**Invalid site kind:**
```
Error: build failed: /path/to/site/walgo.yaml: site.kind is "jekyll"; use "hugo" or "static", or leave it unset to detect
```

**Solution:** Set `site.kind` to `hugo` or `static`, or remove it.

## Best Practices

### 1. Version Control Your Config

```bash
git add walgo.yaml
git commit -m "Add Walgo configuration"
```

### 2. Use Comments

```yaml
# Site is deployed to testnet
walrus:
  network: "testnet"
```

### 3. Don't Commit Secrets

`walgo.yaml` has no secret keys. Walgo stores AI provider keys in `~/.walgo/ai-credentials.yaml`, outside the site. `walgo ai configure` writes that file.

### 4. Test Configuration Changes

```bash
# Build and check the output in the publish directory
walgo build

# Preview the deployment plan without uploading
walgo deploy --dry-run

# Or test with HTTP deployment first
walgo deploy-http \
  --publisher https://publisher.walrus-testnet.walrus.space \
  --aggregator https://aggregator.walrus-testnet.walrus.space
```

### 5. Document Your Settings

Add README with configuration explanation:

```markdown
# Configuration

- `obfuscate: false` - Our scripts break when variable names change
- `publishDir: dist` - Our build writes to dist/
```

## Related Documentation

- [Commands Reference](COMMANDS.md) - All command-line flags
- [Environment Variables](CONFIGURATION.md#environment-variables) - Full env var reference
- [Optimizer Documentation](OPTIMIZER.md) - Detailed optimizer guide
- [Deployment Guide](DEPLOYMENT.md) - Deployment configuration

## Troubleshooting

See [Troubleshooting Guide](TROUBLESHOOTING.md) for configuration-related issues.

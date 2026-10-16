# Optimizer

The optimizer minifies the HTML, CSS, and JavaScript in a built site before walgo deploys it. It uses [tdewolff/minify](https://github.com/tdewolff/minify), the minifier behind `hugo --minify`, and rewrites each file in place in the publish directory. Source files are never touched.

## When it runs

- `walgo build`, and every command that builds first (`deploy`, `launch`, `update`, `serve`, and the desktop app), runs the optimizer after Hugo when `optimizer.enabled` is true.
- Static sites skip it. A site is static when it has no Hugo config or when `walgo.yaml` sets `site.kind: static`; walgo deploys its files exactly as they are.
- `walgo optimize [directory]` runs it on demand. With no argument it targets the site's publish directory and reads `walgo.yaml` from the current directory. With an argument it targets that directory and reads `walgo.yaml` from it, falling back to the defaults below.

The optimizer skips files whose name contains `.min.` and files that match `skipPatterns`.

## Each optimizer needs its minify flag

An optimizer changes a file only when both its `enabled` key and its minify flag are true:

| File types | Runs when |
|---|---|
| `.html`, `.htm` | `html.enabled` and `html.minifyHTML` |
| `.css` | `css.enabled` and `css.minifyCSS` |
| `.js`, `.mjs` | `js.enabled` and `js.minifyJS` |

The other keys refine that pass. On their own, with the minify flag off, they do nothing.

## Configuration

These are the defaults that `walgo init` writes:

```yaml
optimizer:
  enabled: true
  verbose: false
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
```

### Global keys

| Key | Effect |
|---|---|
| `enabled` | Turns the whole optimizer on or off, for builds and for `walgo optimize`. |
| `verbose` | Prints an error line for each file that fails to optimize. |
| `skipPatterns` | Glob patterns to skip. A pattern without `/` matches the file name; a pattern with `/` also matches the full path. |
| `outputDir` | Not implemented. walgo accepts the key and ignores it. |

### HTML keys

HTML minification keeps the document tags, end tags, default attribute values, and special comments such as IE conditional comments, which matches Hugo's own minify settings.

| Key | Effect |
|---|---|
| `minifyHTML` | Minifies HTML. Required for every key below. |
| `removeComments` | Drops ordinary comments. When false, comments stay. |
| `removeWhitespace` | Collapses whitespace between elements. When false, whitespace stays. Content of `<pre>` and `<textarea>` is always kept. |
| `compressInlineCSS` | Also minifies `<style>` blocks. |
| `compressInlineJS` | Also minifies inline `<script>` blocks. Variable names are kept. |

### CSS keys

| Key | Effect |
|---|---|
| `minifyCSS` | Minifies CSS. Minification always strips comments and shortens colors and numbers. |
| `removeComments` | No separate effect; minification already strips comments. |
| `compressColors` | No separate effect; minification already shortens colors. |
| `removeUnused` | Not implemented. walgo accepts the key and ignores it. The `walgo optimize --remove-unused-css` flag is ignored too. |
| `autoprefixer` | Not implemented. walgo accepts the key and ignores it. |

### JavaScript keys

| Key | Effect |
|---|---|
| `minifyJS` | Minifies JavaScript. Minification always strips comments. |
| `removeComments` | No separate effect; minification already strips comments. |
| `obfuscate` | Lets the minifier shorten local variable names. When false, names are kept. |
| `sourceMaps` | Not implemented. walgo writes no source maps. |

The unimplemented keys stay in the schema so existing `walgo.yaml` files keep loading.

## `walgo optimize` flags

| Flag | Effect |
|---|---|
| `--html=false` | Sets `html.enabled` to false for this run. |
| `--css=false` | Sets `css.enabled` to false for this run. |
| `--js=false` | Sets `js.enabled` to false for this run. |
| `-v`, `--verbose` | Sets `verbose` for this run. |
| `--remove-unused-css` | Not implemented; has no effect. |

## Troubleshooting

- A page breaks after minification: set `html.removeWhitespace: false`, rebuild, and compare.
- A script breaks: set `js.obfuscate: false`. If it still breaks, add the file to `skipPatterns`.
- Third-party bundles that ship minified are skipped when their name contains `.min.`. Add other prebuilt files to `skipPatterns`.

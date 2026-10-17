# Troubleshooting Guide

Solutions to common issues when using Walgo.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Build Issues](#build-issues)
- [Deployment Issues](#deployment-issues)
- [Optimization Issues](#optimization-issues)
- [Network Issues](#network-issues)
- [Configuration Issues](#configuration-issues)
- [Wallet Issues](#wallet-issues)
- [Getting More Help](#getting-more-help)

## Installation Issues

### "walgo: command not found"

**Symptoms:**

```bash
$ walgo version
bash: walgo: command not found
```

**Solutions:**

1. **Verify Installation:**

   ```bash
   # Check if walgo is installed
   which walgo
   ls -la $(which walgo)
   ```

2. **Add to PATH:**

   ```bash
   # For user installation
   export PATH="$PATH:$HOME/.local/bin"

   # For system installation
   export PATH="$PATH:/usr/local/bin"

   # Make permanent (add to ~/.bashrc or ~/.zshrc)
   echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Reinstall:**

   ```bash
   # Using install script
   curl -fsSL https://raw.githubusercontent.com/ganbitlabs/walgo/main/install.sh | bash

   # Or using Go
   go install github.com/ganbitlabs/walgo@latest
   ```

---

### "Permission denied" when installing

**Symptoms:**

```bash
$ sudo mv walgo /usr/local/bin/
mv: cannot move 'walgo' to '/usr/local/bin/walgo': Permission denied
```

**Solutions:**

1. **Use sudo:**

   ```bash
   sudo mv walgo /usr/local/bin/
   ```

2. **Install to user directory (no sudo needed):**

   ```bash
   mkdir -p ~/.local/bin
   mv walgo ~/.local/bin/
   export PATH="$PATH:$HOME/.local/bin"
   ```

3. **Check ownership:**
   ```bash
   ls -la /usr/local/bin/
   sudo chown $(whoami) /usr/local/bin/
   ```

---

### macOS: "walgo cannot be opened because the developer cannot be verified"

**Symptoms:**
macOS security warning when running walgo.

**Solutions:**

1. **Remove quarantine attribute:**

   ```bash
   xattr -d com.apple.quarantine /usr/local/bin/walgo
   ```

2. **Or allow in System Preferences:**
   - System Preferences → Security & Privacy
   - Click "Allow Anyway" next to walgo warning

---

### Windows: "Windows protected your PC" SmartScreen warning

**Symptoms:**
Windows SmartScreen blocks walgo.exe.

**Solutions:**

1. **Click "More info"** → "Run anyway"

2. **Or disable SmartScreen (not recommended):**
   - Windows Security → App & browser control
   - Reputation-based protection settings
   - Turn off "Check apps and files"

---

## Build Issues

### "Hugo not found" error

**Symptoms:**

```bash
$ walgo build
Error: build failed: hugo is not installed or not found in PATH
```

**Solutions:**

1. **Install Hugo:**

   ```bash
   # macOS
   brew install hugo

   # Linux
   sudo apt install hugo  # Debian/Ubuntu
   sudo dnf install hugo  # Fedora

   # Windows
   scoop install hugo-extended
   ```

2. **Verify installation:**

   ```bash
   hugo version
   ```

3. **Check PATH:**
   ```bash
   which hugo
   export PATH="$PATH:/usr/local/bin"
   ```

---

### Build fails with theme errors

**Symptoms:**

```bash
$ walgo build
ERROR failed to load modules: module "PaperMod" not found in "/path/to/site/themes/PaperMod"; either add it as a Hugo Module or store it in "/path/to/site/themes".: module does not exist
```

**Solutions:**

1. **Initialize and update Git submodules:**

   ```bash
   git submodule update --init --recursive
   ```

2. **Re-add theme:**

   ```bash
   git submodule add --force https://github.com/adityatelange/hugo-PaperMod.git themes/PaperMod

   # Or download it with walgo (replaces the themes in themes/)
   walgo theme install https://github.com/adityatelange/hugo-PaperMod
   ```

3. **Verify theme in config:**
   ```toml
   # hugo.toml
   theme = "PaperMod"  # Must match directory name
   ```

---

### Build fails with "config file not found"

**Symptoms:**

```bash
$ walgo build
Error: build failed: walgo.yaml not found in /path/to/site
```

Or, when the site has no Hugo config:

```bash
$ walgo build
Error: build failed: /path/to/site is a static site (no Hugo config such as hugo.toml, or site.kind: static in walgo.yaml) and its publish directory /path/to/site/public does not exist
```

**Solutions:**

1. **Check for config file:**

   ```bash
   ls -la | grep -E "hugo\.(toml|yaml|json)|config\.(toml|yaml|json)"
   ```

2. **Create config if missing:**

   ```bash
   # Create hugo.toml
   echo 'baseURL = "http://localhost:1313/"
   languageCode = "en-us"
   title = "My Site"' > hugo.toml
   ```

3. **Run from the site root:**
   `walgo build` reads `walgo.yaml` and the Hugo config from the current directory. The global `--config` flag does not change that.
   ```bash
   cd /path/to/site
   walgo build
   ```

---

### Build produces empty public/ directory

**Symptoms:**
Build succeeds but `public/` is empty or has minimal files.

**Solutions:**

1. **Check for content:**

   ```bash
   ls -la content/
   ```

2. **Create content:**

   ```bash
   hugo new posts/hello.md
   # Edit content/posts/hello.md and set draft: false
   ```

3. **Preview with drafts (for testing):**
   `walgo build` has no drafts flag. Preview drafts with `walgo serve`:
   ```bash
   walgo serve --drafts
   ```

---

## Deployment Issues

### "Insufficient funds" error (on-chain)

**Symptoms:**

`walgo deploy` stops with `Deployment failed:`, followed by the site-builder command and its output. The `Stderr:` part holds site-builder's error about the SUI or WAL balance.

**Solutions:**

1. **Check balance:**

   ```bash
   walgo doctor
   ```

2. **Get testnet SUI and WAL:**

   - Visit `https://faucet.sui.io/?address=<your-address>` (`walgo doctor` prints this link when your balance is empty)
   - Exchange SUI for WAL: `walrus get-wal --context testnet`

3. **Reduce epochs:**

   ```bash
   walgo deploy --epochs 1  # Minimum cost
   ```

4. **For mainnet, buy SUI:**
   - Use exchanges (Binance, Coinbase, etc.)
   - Transfer to your wallet address

---

### "site-builder not found" error

**Symptoms:**

```bash
$ walgo deploy
Error: required tools not found: site-builder

  On-chain deployments need sui, walrus and site-builder.

  To fix:  walgo setup-deps
```

**Solutions:**

1. **Install site-builder:**

   ```bash
   walgo setup-deps
   ```

2. **Verify installation:**

   ```bash
   which site-builder
   site-builder --version
   ```

3. **Manual installation:**
   ```bash
   suiup install site-builder@mainnet
   ```

---

### HTTP deployment fails with network error

**Symptoms:**

```bash
$ walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space --aggregator https://aggregator.walrus-testnet.walrus.space
❌ Error: HTTP deploy failed: ...
```

**Solutions:**

1. **Check internet connection:**

   ```bash
   ping 8.8.8.8
   curl -I https://walrus.site
   ```

2. **Run diagnostics:**

   ```bash
   walgo doctor
   ```

3. **Upload one file at a time:**

   `--mode blobs` sends one request per file and retries each file on transient errors, instead of sending the whole site in one Quilt request.

   ```bash
   walgo deploy-http --mode blobs \
     --publisher https://publisher.walrus-testnet.walrus.space \
     --aggregator https://aggregator.walrus-testnet.walrus.space
   ```

4. **Use other endpoints:**
   Pick another publisher and aggregator from the [public services list](https://docs.wal.app/docs/usage/web-api#public-services):
   ```bash
   walgo deploy-http --publisher <publisher-url> --aggregator <aggregator-url>
   ```

---

### Deployment succeeds but site doesn't load

**Symptoms:**
Site deployed successfully but URL shows error or blank page.

**Solutions:**

1. **Wait a moment:**

   - Propagation takes 10-30 seconds
   - Try refreshing after waiting

2. **Check deployment status:**

   ```bash
   walgo status <object-id>
   ```

3. **Test index.html:**

   ```bash
   ls -la public/index.html
   cat public/index.html | head -20
   ```

4. **Clear browser cache:**

   - Hard reload: `Ctrl+F5` (Windows) or `Cmd+Shift+R` (Mac)

5. **Check browser console:**
   - F12 → Console tab
   - Look for errors

---

### "Object not found" when updating

**Symptoms:**

```bash
$ walgo update 0x7b5a...
Error: Site object not found
```

**Solutions:**

1. **Verify object ID:**

   ```bash
   # Check saved object ID
   cat walgo.yaml | grep projectID
   ```

2. **Check correct network:**
   `walgo update` and `walgo status` use the active Sui environment.

   ```bash
   # If deployed on testnet
   sui client switch --env testnet
   walgo update 0x7b5a...

   # If deployed on mainnet
   sui client switch --env mainnet
   walgo update 0x7b5a...
   ```

3. **Verify object exists:**
   ```bash
   walgo status 0x7b5a...
   ```

---

## Optimization Issues

### JavaScript breaks after optimization

**Symptoms:**
Site works before optimization but JavaScript errors after.

**Solutions:**

1. **Disable JS obfuscation:**

   ```yaml
   # walgo.yaml
   optimizer:
     js:
       obfuscate: false # Set to false
   ```

2. **Turn off JavaScript minification:**

   ```yaml
   # walgo.yaml
   optimizer:
     js:
       minifyJS: false
   ```

3. **Skip specific files:**

   ```yaml
   optimizer:
     skipPatterns:
       - "problematic-script.js"
       - "*.bundle.js"
   ```

4. **Test optimization separately:**

   ```bash
   # Build without optimization (set optimizer.enabled: false in walgo.yaml)
   walgo build

   # Test optimization manually
   walgo optimize public/ --verbose
   ```

   With a directory argument, `walgo optimize` uses the default optimizer settings, because the directory has no `walgo.yaml`.

---

### CSS styling broken after optimization

**Symptoms:**
Site looks different after optimization, styles missing.

**Solutions:**

1. **Turn off CSS minification:**

   ```yaml
   # walgo.yaml
   optimizer:
     css:
       minifyCSS: false
   ```

   `optimizer.css.removeUnused` is not implemented, so it does not remove any CSS.

2. **Check Hugo's CSS minifier:**

   `walgo build` runs `hugo --minify`. Hugo's CSS minifier can break `var()` spacing in multi-value properties, so the `hugo.toml` files walgo creates turn it off:

   ```toml
   # hugo.toml
   [minify]
     disableCSS = true
   ```

3. **Compare before/after:**

   ```bash
   # Build without optimization (set optimizer.enabled: false in walgo.yaml)
   walgo build
   cp -r public public-before

   # Optimize a copy with the default settings
   cp -r public public-after
   walgo optimize public-after

   # Diff the CSS files
   diff public-before/style.css public-after/style.css
   ```

---

### Optimization too aggressive

**Symptoms:**
Site broken after optimization, but unsure which setting caused it.

**Solutions:**

1. **Disable optimization entirely:**

   ```yaml
   optimizer:
     enabled: false
   ```

2. **Enable selectively:**

   ```yaml
   optimizer:
     enabled: true
     html:
       enabled: true # Start with HTML only
     css:
       enabled: false
     js:
       enabled: false
   ```

3. **Test each optimizer:**

   ```bash
   # Test HTML optimization
   walgo optimize public --css=false --js=false

   # Test CSS optimization
   walgo optimize public --html=false --js=false

   # Test JS optimization
   walgo optimize public --html=false --css=false
   ```

   Pass the directory. Without it, `walgo optimize` uses `walgo.yaml`, and does nothing while `optimizer.enabled` is `false`.

---

## Network Issues

### "Cannot connect to Walrus" error

**Symptoms:**

```bash
Error: Failed to connect to Walrus publisher/aggregator
```

**Solutions:**

1. **Check internet connection:**

   ```bash
   ping 8.8.8.8
   curl -I https://walrus.site
   ```

2. **Test Walrus endpoints:**

   ```bash
   curl -I https://publisher.walrus-testnet.walrus.space
   curl -I https://aggregator.walrus-testnet.walrus.space
   ```

3. **Run diagnostics:**

   ```bash
   walgo doctor -v
   ```

4. **Check firewall:**
   - Ensure HTTPS (port 443) is not blocked
   - Check corporate firewall/proxy settings

---

### Site is live on chain but serves nothing (expired storage)

**Symptoms:** the portal returns 404, while `walgo status` still finds the object
and reports expired resources.

**Cause:** a Walrus Site is two things with different lifetimes. The Sui object
never expires; the blobs holding the content do, after the number of epochs paid
for at deploy time (a mainnet epoch is ~2 weeks, a testnet epoch ~1 day).

**Expired storage cannot be extended.** The walrus contract rejects extending an
expired blob (`assert_certified_not_expired`), so there is no renew path once the
deadline passes — the content has to be stored again:

```bash
walgo build                 # regenerate the files locally
walgo update --epochs 10    # re-upload; the site keeps its object ID
walrus burn-blobs --all-expired   # clean up the dead blob objects afterwards
```

Notes:

- Re-uploading costs the full write + storage price. Extending, which only works
  *before* expiry, is cheaper — `walgo status` prints the expiration date so the
  renewal can happen in time.
- The local build is the only source of the content once the blobs are gone.
  Walrus no longer holds it, so a site whose sources are lost cannot be restored.
- `walgo update` defaults to 1 epoch. Pass `--epochs` unless a two-week (mainnet)
  or one-day (testnet) lifetime is what you want.

---

### "Method not found. JSON-RPC on public fullnodes has been deprecated"

**Symptoms:**

```bash
Error: Failed to get client from url: max failovers exceeded
ErrorObject { code: MethodNotFound, message: "Method not found. JSON-RPC on public
fullnodes has been deprecated. Please migrate to gRPC or GraphQL endpoints." }
```

**Cause:** Sui Foundation disabled JSON-RPC on its public fullnodes on 2026-07-31.
Older walrus and site-builder builds only speak JSON-RPC, so every deployment fails.

**Solution:** update the tools. Minimum versions are walrus **1.52.0** and
site-builder **2.12.0**:

```bash
suiup install walrus@mainnet
suiup install site-builder@mainnet
walrus --version && site-builder --version
```

`rpc_url` in `sites-config.yaml` and `rpc_urls` in `client_config.yaml` stay the
same — `https://fullnode.<network>.sui.io:443` serves gRPC on the same address.

See the [JSON-RPC migration guide](https://docs.sui.io/develop/accessing-data/json-rpc-migration).

---

### "RPC error" when deploying

**Symptoms:**

```bash
Error: Failed to connect to Sui RPC node
```

**Solutions:**

1. **Check Sui network status:**

   - Visit [Sui status page](https://status.sui.io)

2. **Check whether the endpoint is serving stale state:**

   ```bash
   # gRPC (what walrus and the sui CLI use)
   sui client object 0x5 --json | grep version
   # GraphQL (independent read path)
   curl -s https://graphql.mainnet.sui.io/graphql -H 'Content-Type: application/json' \
     -d '{"query":"{ object(address:\"0x5\"){ version } }"}'
   ```

   Different versions mean the public fullnode is lagging. Errors like
   `object ... is unavailable for consumption, current version: ...` or
   `could not find WAL coins with sufficient balance` (with a funded wallet)
   come from that lag. Point `rpc_urls` at another provider or wait it out.

3. **Wait and retry:**
   - Network congestion is temporary
   - Try again in a few minutes

---

## Configuration Issues

### "Invalid configuration file" error

**Symptoms:**

```bash
$ walgo build
Error: build failed: failed to parse config file /path/to/site/walgo.yaml: yaml: line 5: mapping values are not allowed in this context
```

**Solutions:**

1. **Validate YAML syntax:**

   ```bash
   # Use online validator: https://www.yamllint.com/
   # Or install yamllint
   yamllint walgo.yaml
   ```

2. **Check indentation:**

   ```yaml
   # BAD - mixed spaces/tabs
   optimizer:
   	enabled: true  # Tab here
     html:          # Spaces here

   # GOOD - consistent spaces
   optimizer:
     enabled: true
     html:
       enabled: true
   ```

3. **Check for special characters:**

   ```yaml
   # BAD - unquoted special chars
   title: My Site: The Best One

   # GOOD - quoted
   title: "My Site: The Best One"
   ```

---

### Configuration not being applied

**Symptoms:**
Changes to `walgo.yaml` don't seem to take effect.

**Solutions:**

1. **Check file location:**

   `walgo build`, `walgo deploy`, and the other site commands read `walgo.yaml` from the current directory. The global `--config` flag and `~/.walgo.yaml` do not change the settings they use.

   ```bash
   # Reports whether walgo.yaml exists in the current directory
   walgo doctor
   ```

2. **Run from the site root:**

   ```bash
   cd /path/to/site
   walgo build
   ```

3. **Check for typos:**

   ```yaml
   # BAD - typo
   optimiser: # British spelling
     enabled: true

   # GOOD
   optimizer: # American spelling
     enabled: true
   ```

---

### Environment variables not working

**Symptoms:**
Environment variables don't override config file.

**Solutions:**

1. **Use walgo.yaml and flags instead:**

   Walgo does not read settings from environment variables. Put settings in `walgo.yaml`, and pass per-run values as flags:

   ```bash
   walgo deploy --epochs 5
   ```

2. **Know the variables walgo does read:**

   - `WALGO_ASCII=1` (or `true`) prints ASCII markers instead of emoji. It wins over `WALGO_EMOJI`
   - `WALGO_EMOJI=1` (or `true`) prints emoji, even in CI or a terminal without emoji support
   - `WALGO_THEME_DOWNLOAD_TIMEOUT` sets the theme download timeout as a Go duration, such as `5m` (default `2m`)

3. **Verify variable is set:**

   ```bash
   env | grep WALGO
   ```

---

## Wallet Issues

### "Wallet not configured" error

**Symptoms:**

`walgo deploy` stops with `could not determine the active Sui network`, or with `Deployment failed:` and a wallet error from site-builder in the `Stderr:` part.

**Solutions:**

1. **Set up the Sui wallet:**

   ```bash
   sui client
   ```

2. **Run setup:**

   `walgo setup` writes `~/.config/walrus/sites-config.yaml`, which tells site-builder where the wallet is.

   ```bash
   walgo setup --network testnet
   ```

3. **Verify wallet:**

   ```bash
   walgo doctor
   ```

4. **Check Sui CLI:**
   ```bash
   sui client active-address
   ```

---

### "Invalid recovery phrase" when importing wallet

**Symptoms:**
Recovery phrase rejected during wallet import.

**Solutions:**

1. **Check word count:**

   - Must be 12 or 24 words
   - Verify no typos

2. **Check word list:**

   - Words must be from BIP39 word list
   - Use correct spelling

3. **Create new wallet instead:**
   ```bash
   sui client new-address ed25519
   # This creates a new address and prints its recovery phrase
   ```

---

### "Transaction failed" error

**Symptoms:**

```bash
Error: Transaction failed with error: ...
```

**Solutions:**

1. **Check balance:**

   ```bash
   walgo doctor
   ```

2. **Increase gas budget:**

   `walgo deploy` has no gas flag. Raise `gas_budget` for your network in `~/.config/walrus/sites-config.yaml`:

   ```yaml
   contexts:
     testnet:
       general:
         gas_budget: 1000000000
   ```

3. **Wait and retry:**

   - Network congestion
   - Try again in a few minutes

4. **Check transaction on explorer:**
   - Visit Sui Explorer
   - Search for transaction ID
   - See detailed error

---

## Obsidian Import Issues

### "Vault not found" error

**Symptoms:**

```bash
$ walgo import ~/Documents/Vault
Error: vault path does not exist: /Users/yourname/Documents/Vault
```

**Solutions:**

1. **Check path:**

   ```bash
   ls -la ~/Documents/Vault

   # Use absolute path
   walgo import /Users/yourname/Documents/Vault
   ```

2. **Check permissions:**
   ```bash
   ls -ld ~/Documents/Vault
   # Should show read permissions
   ```

---

### Wikilinks not converting

**Symptoms:**
Imported files still contain `[[wikilinks]]` instead of markdown links.

**Solutions:**

1. **Enable conversion:**

   Conversion is on by default. Check that you did not pass `--convert-wikilinks=false`. `walgo import` creates a new site with a fresh `walgo.yaml`, so `obsidian` settings in another site's `walgo.yaml` do not apply.

2. **Or use flag:**
   ```bash
   walgo import ~/Vault --convert-wikilinks
   ```

---

### Attachments not copying

**Symptoms:**
Images/PDFs missing after import.

**Solutions:**

1. **Check attachment directory:**

   Attachments go to `static/images/` in the new site. `--attachment-dir` names a different directory under `static/`, and walgo creates it:

   ```bash
   walgo import ~/Vault --attachment-dir files
   ```

2. **Check the copied files:**

   ```bash
   ls <site-name>/static/images
   ```

3. **Check file paths in Obsidian:**
   - Must use relative paths
   - Must be in vault directory

---

## Getting More Help

### Using walgo doctor

```bash
# Basic diagnostics
walgo doctor

# Verbose output
walgo doctor -v

# Auto-fix issues
walgo doctor --fix-all
```

### Checking Logs

```bash
# Enable verbose output
walgo deploy --verbose
```

`deploy`, `update`, `doctor`, `optimize`, `compress`, `ai pipeline`, and `ai resume` print more with `--verbose` (`-v`). `deploy-http` and `ai plan` accept the flag but print nothing extra. There is no global verbose flag or environment variable.

### Common Diagnostic Commands

```bash
# Check installation
walgo version
which walgo

# Check dependencies
hugo version
site-builder --version
sui --version

# Check configuration
walgo doctor -v

# Check wallet
sui client active-address
sui client balance

# Check network
ping walrus.site
curl -I https://walrus.site
```

### Reporting Issues

When reporting issues on GitHub, include:

1. **Walgo version:**

   ```bash
   walgo version
   ```

2. **System info:**

   ```bash
   uname -a  # macOS/Linux
   # or
   systeminfo  # Windows
   ```

3. **Command and error:**

   ```bash
   # Exact command you ran
   walgo deploy --epochs 5

   # Full error output
   Error: ...
   ```

4. **Verbose output:**

   ```bash
   walgo deploy --verbose   # or the --verbose flag of the command that failed
   ```

5. **Configuration (without secrets):**
   ```yaml
   # Contents of walgo.yaml (remove any API keys)
   ```

### Support Channels

- **Documentation:** [https://github.com/ganbitlabs/walgo/tree/main/docs](https://github.com/ganbitlabs/walgo/tree/main/docs)
- **GitHub Issues:** [https://github.com/ganbitlabs/walgo/issues](https://github.com/ganbitlabs/walgo/issues)
- **GitHub Discussions:** [https://github.com/ganbitlabs/walgo/discussions](https://github.com/ganbitlabs/walgo/discussions)
- **Feedback Form:** [https://tally.so/r/447YBO](https://tally.so/r/447YBO)
- **Walrus Discord:** [https://discord.gg/walrus](https://discord.gg/walrus)
- **Sui Discord:** [https://discord.gg/sui](https://discord.gg/sui)

### Before Asking for Help

1. **Search existing issues:**

   - Someone may have had the same problem
   - Solution might already exist

2. **Read documentation:**

   - Check relevant guides
   - Review configuration reference

3. **Run diagnostics:**

   ```bash
   walgo doctor -v
   ```

4. **Try verbose mode:**

   ```bash
   walgo deploy --verbose   # or the --verbose flag of the command that failed
   ```

5. **Isolate the issue:**
   - Test with minimal example
   - Disable optimizations
   - Try HTTP deployment first

## Related Documentation

- [Getting Started Guide](GETTING_STARTED.md)
- [Installation Guide](INSTALLATION.md)
- [Configuration Reference](CONFIGURATION.md)
- [Deployment Guide](DEPLOYMENT.md)
- [Commands Reference](COMMANDS.md)

# AI-Powered Content Generation

## Overview

Walgo includes powerful AI features to help you generate and update Hugo content automatically. Use AI to:

- **Generate new blog posts** with proper Hugo formatting
- **Create pages** (About, Contact, Services, etc.)
- **Update existing content** with AI-powered edits
- **Maintain proper Hugo frontmatter** and Markdown structure

**Supported AI Providers:**

- **OpenAI**: GPT-4, GPT-3.5 Turbo, and other OpenAI models
- **OpenRouter**: Access to multiple models including Claude, GPT-4, Gemini, and more

## Quick Start

### 1. Configure AI Provider

First, set up your AI provider credentials:

```bash
walgo ai configure
```

You'll be prompted to:

1. Select provider (OpenAI or OpenRouter)
2. Enter your API key
3. Optionally specify a custom base URL
4. Enter a model name (press Enter for `gpt-4` on OpenAI or `openai/gpt-4` on OpenRouter)

**Getting API Keys:**

- **OpenAI**: Get your API key at [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- **OpenRouter**: Get your API key at [openrouter.ai/keys](https://openrouter.ai/keys)

**Credentials Storage:**

- Credentials are stored securely in `~/.walgo/ai-credentials.yaml`
- File permissions are set to 0600 (owner read/write only)
- Every site on your machine uses the same credentials
- `walgo.yaml` holds no AI settings

### 2. Generate Content

Generate new blog posts or pages from inside your site directory:

```bash
# Interactive mode
walgo ai generate

# Generate without building the site afterwards
walgo ai generate --no-build

# Generate, build, and start the development server
walgo ai generate --serve
```

Walgo lists the content sections it finds in `content/` and asks for your instructions. The AI picks the content type and the filename from your instructions. After saving the file, Walgo fixes its frontmatter and builds the site.

### 3. AI Pipeline (Create Complete Site)

Generate an entire site with AI:

```bash
walgo ai pipeline
```

This command:

1. Asks for the site name, site type, description, and target audience
2. Creates a `<site-name>/` directory with `walgo.yaml`, `hugo.toml`, archetypes, and the theme. `<site-name>` is the site name in lowercase, with spaces and underscores turned into hyphens and other symbols except dots dropped. `My Site!` becomes `my-site/`.
3. Creates a site structure plan (JSON) and saves it to `<site-name>/.walgo/plan.json`
4. Generates the pages, several at a time when the plan has more than three
5. Adds a menu to `hugo.toml`, fixes the content for the theme, and builds the site
6. Saves the site as a draft project in `walgo projects`

If generation is interrupted, run `walgo ai resume` from inside `<site-name>/`.

**Workflow:**

```
🤖 AI Site Pipeline

✓ Using openai (gpt-4)

Site name: my-tech-blog

Site type:
  1) Blog
  2) Docs
Select [1]: 1

Describe your site (1-2 sentences): A personal tech blog

Target audience: Developers and tech enthusiasts
...
📊 Results:
   Total pages: 6
   Completed:   6
   Skipped:     0
   Failed:      0
   Duration:    1m12s
...
   ✓ Created draft project

🎉 Site generated successfully!
```

**Supported Site Types:**

- Blog - Blog with the Ananke theme
- Docs - Documentation site with the Book theme

For a Biolink or Whitepaper site, use `walgo quickstart`, which creates the site from the theme's example content without AI.

**Flags:**

- `--parallel` - `auto`, `sequential`, or `parallel` (default: `auto`)
- `--concurrent` - Maximum pages generated at once (default: 5)
- `--rpm` - Page requests sent to the AI per minute, after a first burst of up to `--concurrent` requests (default: 30). The planning request and HTTP-level retries do not count against it.
- `--dry-run` - Generate each page with the AI, then discard it instead of writing it to `content/`. Every other step still writes to disk: the site folder, `walgo.yaml`, `hugo.toml`, the theme, the plan, the content fixes, the build, and the draft project. The plan marks the discarded pages as done, so `walgo ai resume` does not write them later.
- `-v, --verbose` - Show verbose output

### 4. AI Plan (Create Site Plan Only)

Create a site plan without generating content:

```bash
walgo ai plan
```

This command:

- Asks the same questions as `walgo ai pipeline`
- Creates a site structure plan using AI
- Creates a `<site-name>/` directory with `walgo.yaml`, using the site name exactly as typed
- Saves plan to `<site-name>/.walgo/plan.json`
- Does not generate any content
- Does not install a theme or write `hugo.toml`
- Plan can be reviewed before generation

**Run `walgo ai resume`** from inside `<site-name>/` to generate content from the plan.

### 5. AI Resume (Resume from Plan)

Resume content generation from an existing plan:

```bash
walgo ai resume
```

This command:

- Loads `.walgo/plan.json` from the current directory
- Continues generating remaining pages
- Adds a menu to `hugo.toml` and applies the content fixes
- Creates draft project entry
- Does not build the site. In a folder that `walgo ai pipeline` created, run `walgo build` afterwards.
- A folder from `walgo ai plan` has no Hugo config and no theme, and `walgo ai resume` adds neither. `walgo build` treats that folder as a static site and fails. For a site you can build, use `walgo ai pipeline`.

It takes the same `--parallel`, `--concurrent`, `--rpm`, `--dry-run`, and `--verbose` flags as `walgo ai pipeline`.

**Use Case:** When `walgo ai pipeline` was interrupted, you can resume from where it left off.

### 6. Automatic Content Fixes

Walgo runs its content fixer after `walgo ai generate`, `walgo ai update`, `walgo ai pipeline`, and `walgo ai resume`. The fixer checks every Markdown file in `content/`:

- Fixes YAML quoting and a malformed frontmatter start
- Removes duplicate H1 headings (title is already in frontmatter)
- Adds frontmatter fields the theme's archetypes expect
- Sets draft to false

### 7. Example Generate Session

```
🤖 AI Content Generator (openai: gpt-4)

ℹ️  Current content structure:
  * posts/ (2 files)
      - first-post.md
      - second-post.md

  ✓ Default type: posts

💡 What content do you want to create?
  Examples:
  - Create a blog post about blockchain technology for beginners
  - Write a tutorial on deploying Hugo sites to Walrus
  - Generate an about page for my portfolio

Your instructions: Write a blog post about deploying static sites to Walrus with Hugo

🔄 Generating content (AI will determine type and filename)...

✅ Content generated successfully!
   Type: posts
   File: deploying-to-walrus.md
   Path: /home/you/my-blog/content/posts/deploying-to-walrus.md

🔄 Fixing YAML frontmatter...
✓ YAML frontmatter validated

🔄 Building site...
```

### 8. Update Existing Content

Use AI to update existing Hugo files:

```bash
walgo ai update content/posts/my-post.md
```

**Example Session:**

```
🤖 AI Content Updater (openai: gpt-4)
📄 File: content/posts/my-post.md

Describe what changes you want to make: Add a new section about performance optimization and update the conclusion

🔄 Updating content...

✅ Content updated!

Save changes to file? [Y/n]: Y

✅ File updated: content/posts/my-post.md

🔄 Fixing YAML frontmatter...
✓ YAML frontmatter validated
```

After saving, Walgo runs the content fixer and rebuilds the site.

## Configuration

### AI Configuration

AI settings do not go in `walgo.yaml`. `walgo ai configure` saves the provider, API key, base URL, and model to `~/.walgo/ai-credentials.yaml`, and every site on your machine uses them.

Show the saved settings, with the API key masked:

```bash
walgo ai get
```

Remove them:

```bash
walgo ai remove            # Remove all credentials
walgo ai remove openai     # Remove only OpenAI credentials
```

To change the model, run `walgo ai configure` again for the same provider.

### Available Models

**OpenAI Models:**

- `gpt-4` - Most capable, best for complex content
- `gpt-4-turbo` - Faster, cost-effective alternative
- `gpt-3.5-turbo` - Fast and economical

**OpenRouter Models:**

- `anthropic/claude-3-opus` - Anthropic's most capable model
- `anthropic/claude-3-sonnet` - Balanced performance
- `openai/gpt-4` - GPT-4 via OpenRouter
- `google/gemini-pro` - Google's Gemini Pro
- Many more available at [openrouter.ai/models](https://openrouter.ai/models)

### Credentials Management

**View credentials location:**

```bash
ls ~/.walgo/ai-credentials.yaml
```

**Credentials file structure:**

```yaml
providers:
  openai:
    provider: openai
    api_key: sk-...
    base_url: https://api.openai.com/v1
    model: gpt-4
  openrouter:
    provider: openrouter
    api_key: sk-or-...
    base_url: https://openrouter.ai/api/v1
    model: openai/gpt-4
```

**Security:**

- Never commit `ai-credentials.yaml` to version control
- File is stored in your home directory (`~/.walgo/`)
- File permissions are restrictive (0600)
- API keys are never stored in project files

## Features

### Content Generation

AI-generated content includes:

**Proper Hugo Frontmatter:**

```yaml
---
title: "Your Content Title"
date: 2025-01-15T10:30:00+00:00
draft: false
description: "SEO-friendly description"
tags: ["hugo", "static-sites", "walrus"]
categories: ["tutorials"]
---
```

**Well-Structured Markdown:**

- Clear heading hierarchy
- Proper formatting
- SEO optimization
- Engaging introductions and conclusions
- Code blocks and examples when relevant

**Content Types:**

The AI picks the content type from your instructions and the sections in `content/`.

1. **Blog Posts**

   - Date-based content
   - Tags and categories
   - SEO-optimized
   - Engaging narratives

2. **Pages**
   - Standalone content
   - No dates/categories
   - Professional structure
   - Informative and concise

### Content Updates

When updating existing content, AI will:

- Preserve original frontmatter (unless changes requested)
- Maintain existing style and tone
- Keep proper Hugo formatting
- Update dates if content significantly changes
- Apply requested modifications accurately

### AI System Prompts

Walgo uses specialized system prompts that understand:

- Hugo content file structure
- YAML frontmatter requirements
- Markdown best practices
- SEO optimization
- Content type differences (posts vs pages)

## Workflows

### Complete Content Creation Workflow

```bash
# 1. Configure AI (one-time setup)
walgo ai configure

# 2. Generate content (this also builds the site)
walgo ai generate

# 3. Review generated content
cat content/posts/your-new-post.md

# 4. Edit if needed (manual or AI)
# Manual: edit with your preferred editor
# AI: walgo ai update content/posts/your-new-post.md

# 5. Preview
walgo serve

# 6. Deploy
walgo launch
```

### Bulk Content Generation

Generate multiple posts quickly:

```bash
# Generate first post
walgo ai generate --no-build
# Instruction: "Write a blog post about Hugo basics"

# Generate second post
walgo ai generate --no-build
# Instruction: "Write a blog post about Walrus Sites benefits"

# Generate third post, then build once
walgo ai generate
# Instruction: "Write a tutorial on deploying with Walgo"
```

### Content Refresh Workflow

Update all existing content:

```bash
# Update each post
walgo ai update content/posts/old-post-1.md
# Instruction: "Add current best practices and update examples"

walgo ai update content/posts/old-post-2.md
# Instruction: "Improve SEO and add more details"

# Build and deploy
walgo build
walgo launch
```

## Advanced Usage

### Custom Base URLs

Use custom API endpoints:

```bash
walgo ai configure
# When prompted for base URL, enter custom endpoint
# Example: http://localhost:8000/v1 (for local models)
```

### Multiple Providers

Configure multiple providers and switch between them:

1. Configure OpenAI:

   ```bash
   walgo ai configure
   # Select OpenAI, enter key
   ```

2. Configure OpenRouter:

   ```bash
   walgo ai configure
   # Select OpenRouter, enter key
   ```

3. Walgo uses OpenAI when both providers are configured. To use OpenRouter, remove the OpenAI credentials:
   ```bash
   walgo ai remove openai
   ```

### Content Templates

AI understands Hugo content patterns. You can request specific structures:

```
Instruction: "Create a tutorial-style blog post about X with:
- Introduction explaining why it matters
- Prerequisites section
- Step-by-step instructions with code examples
- Troubleshooting section
- Conclusion with next steps"
```

## Best Practices

### 1. Be Specific with Instructions

**Good:**

```
Write a comprehensive blog post about deploying static sites to Walrus.
Include sections on: benefits of decentralized hosting, step-by-step
deployment guide, cost comparison, and best practices.
```

**Less Effective:**

```
Write about Walrus
```

### 2. Provide Context

Put context in your instructions:

- Target audience information
- Tone preferences (technical, casual, professional)
- Specific examples to include
- Length preferences

### 3. Review AI Output

Always review and edit AI-generated content:

- Check technical accuracy
- Verify links and references
- Adjust tone and style
- Add personal insights
- Update dates and metadata

### 4. Iterate with Updates

If first generation isn't perfect:

```bash
# Generate initial content
walgo ai generate

# Update with refinements
walgo ai update content/posts/file.md
# Instruction: "Make the introduction more engaging and add code examples"
```

### 5. Check Draft Status Before You Deploy

- Generated content has `draft: false`, so the next build publishes it
- The content fixer changes `draft: true` to `draft: false` in the YAML frontmatter of Markdown files under `content/`, including pages you did not generate
- Review content before you deploy
- To hold a page back, set `draft: true` after your last `walgo ai` command

## Troubleshooting

### "no AI credentials found - run 'walgo ai configure' first"

**Solution:**

```bash
walgo ai configure
```

### "No credentials found for provider"

**Solution:**

```bash
# Reconfigure credentials
walgo ai configure

# Or manually check credentials file
cat ~/.walgo/ai-credentials.yaml
```

### "API request failed"

**Possible causes:**

1. Invalid API key - verify your key is correct
2. Insufficient credits/quota - check your provider account
3. Network issues - check internet connection
4. Rate limiting - wait and try again

**Solution:**

```bash
# Reconfigure with valid credentials
walgo ai configure
```

### "failed to parse credentials file"

**Solution:**

```bash
# Check the credentials file
cat ~/.walgo/ai-credentials.yaml

# If it is damaged, remove it and reconfigure
rm ~/.walgo/ai-credentials.yaml
walgo ai configure
```

### Generated Content Has Issues

**Common issues:**

- Missing frontmatter fields - update instructions to be more specific
- Wrong date format - regenerate or manually fix
- Incorrect content type - name the type in your instructions, such as "a blog post" or "an about page"

**Solution:**

```bash
# Regenerate with better instructions
walgo ai generate

# Or update the problematic file
walgo ai update content/posts/file.md
# Instruction: "Fix frontmatter date format to ISO 8601"
```

## Costs

### OpenAI Pricing

Approximate costs (as of 2025):

- GPT-4: ~$0.03 per 1K tokens (input), ~$0.06 per 1K tokens (output)
- GPT-3.5 Turbo: ~$0.0015 per 1K tokens (input), ~$0.002 per 1K tokens (output)

**Typical content generation:**

- Blog post (500-1000 words): $0.10-$0.30 with GPT-4
- Short update: $0.05-$0.10 with GPT-4

### OpenRouter Pricing

Varies by model:

- Claude 3 Opus: ~$15 per 1M input tokens
- GPT-4: Similar to OpenAI pricing
- Check current pricing: [openrouter.ai/models](https://openrouter.ai/models)

**Benefits of OpenRouter:**

- Access to multiple providers with one API key
- Automatic fallback between models
- Unified billing

## Security & Privacy

### Credentials

- API keys stored locally in `~/.walgo/ai-credentials.yaml`
- File permissions: 0600 (owner read/write only)
- Never stored in project files or version control
- Never transmitted except to configured AI provider

### Content Privacy

When you generate content:

- Your instructions are sent to the AI provider
- Generated content is returned to your local machine
- Content is stored locally in your Hugo site
- No data is stored on Walgo servers (there are none!)

**Important:**

- Review your AI provider's privacy policy
- Don't include sensitive information in prompts
- Be aware that prompts may be used for model training (provider-dependent)

## Examples

### Example 1: Generate Tutorial Post

```bash
walgo ai generate
```

**Instruction:**

```
Create a detailed tutorial on setting up a Hugo blog with the Ananke theme.
Include installation steps, configuration examples, and how to create your
first post. Target audience: developers new to Hugo.
```

**Result:**
A complete blog post with:

- Clear installation instructions
- Configuration examples with code blocks
- Step-by-step first post creation
- Proper frontmatter and SEO

### Example 2: Create About Page

```bash
walgo ai generate
```

**Instruction:**

```
Create an About page for a software developer's portfolio site. Include
sections for: background, skills, experience, and contact information.
Keep it professional and concise.
```

**Result:**
A well-structured About page with appropriate frontmatter and professional content.

### Example 3: Update Existing Post

```bash
walgo ai update content/posts/walrus-intro.md
```

**Instruction:**

```
Add a new section called "Cost Comparison" comparing Walrus storage costs
to traditional cloud hosting. Update the conclusion to reference this new section.
```

**Result:**
Original post updated with new section, maintaining existing structure and style.

### Example 4: Refresh Old Content

```bash
walgo ai update content/posts/old-tutorial.md
```

**Instruction:**

```
This tutorial is from 2023. Update it to reflect 2025 best practices,
update any deprecated commands, and improve code examples. Keep the
same overall structure and tone.
```

**Result:**
Refreshed content with current information while preserving original intent.

## FAQ

**Q: Which AI provider should I use?**
A:

- **OpenAI**: Best for straightforward use, well-documented, reliable
- **OpenRouter**: Best for accessing multiple models (Claude, GPT-4, etc.) with one API key

**Q: Can I use local AI models?**
A: Yes! Configure a custom base URL pointing to your local model server (must be OpenAI API-compatible).

**Q: Will AI always generate perfect content?**
A: No. AI-generated content should always be reviewed and edited. Think of it as a first draft.

**Q: How much does it cost?**
A: Varies by provider and model. Typical blog post: $0.05-$0.30. See Costs section above.

**Q: Can AI update my Hugo theme files?**
A: The AI features are designed for content (markdown files), not theme files. Manual editing recommended for themes.

**Q: Is my API key secure?**
A: Yes. It's stored locally with restrictive permissions and never committed to version control.

**Q: Can I generate content in other languages?**
A: Yes! Simply write your instructions in your target language, and AI will generate content in that language.

**Q: What if I don't like the generated content?**
A: You can:

1. Regenerate with better instructions
2. Use `walgo ai update` to refine it
3. Manually edit the file
4. Delete it and start over

**Q: Can I batch process multiple files?**
A: Currently, each command processes one file. For batch updates, write a shell script that calls `walgo ai update` for each file.

## Related Commands

- `walgo init` - Create new Hugo site
- `walgo new` - Create new content (traditional way)
- `walgo build` - Build site
- `walgo serve` - Preview site locally
- `walgo launch` - Deploy to Walrus

## Support

For issues or questions:

- GitHub Issues: [github.com/ganbitlabs/walgo/issues](https://github.com/ganbitlabs/walgo/issues)
- Check provider documentation: [OpenAI](https://platform.openai.com/docs) | [OpenRouter](https://openrouter.ai/docs)

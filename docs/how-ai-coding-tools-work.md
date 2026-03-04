# How AI Coding Tools Actually Work Under the Hood

Everything you need to understand about CLAUDE.md, instructions.md, GEMINI.md, MCP servers, skills, plugins, extensions, and how they all connect — explained from the ground up.

---

## The One Thing Nobody Explains

Every AI coding tool — Claude Code, Codex CLI, Gemini CLI, Cursor — works the same way at the deepest level. They all:

1. Take your message
2. Wrap it with instructions, tool definitions, and context
3. Send it to an API
4. Get a response back
5. Show it to you

The difference between them is just **packaging**. Different file names, different formats, different config paths — but the exact same architecture. Once you understand one, you understand all of them.

---

## How CLAUDE.md, instructions.md, and GEMINI.md Work

This is the part most people don't understand.

**CLAUDE.md gets injected into EVERY message you send.** Not once at the start of a session. Not once per conversation. Every single API call.

When you open Claude Code in a project, it:

1. Looks for `CLAUDE.md` in `~/.claude/CLAUDE.md` (your global instructions)
2. Looks for `CLAUDE.md` in parent directories above your project
3. Looks for `CLAUDE.md` in your project root
4. Concatenates them all together
5. Prepends them to the system prompt on **every API call**

So if your CLAUDE.md says "always use bun instead of npm", that instruction is sent to Claude every single time you send a message. If you write 50 messages in a session, that line gets sent 50 times. It never "forgets" your CLAUDE.md — because it re-reads it every time.

**That's why CLAUDE.md should be concise.** Every line costs tokens on every single message. A 500-line CLAUDE.md means 500 lines of tokens consumed on every single request, eating into your context window and your budget.

**Codex CLI does the exact same thing** with `~/.codex/instructions.md`. Gemini CLI does it with `GEMINI.md`. Cursor does it with `.cursorrules`. Different file names, identical mechanism: read the file, inject it into the prompt, send it with every message.

---

## What Actually Gets Sent to the API

Every time you send a message in Claude Code, here's what is actually transmitted to the Claude API:

```
API Call = {
  system: [
    "You are Claude Code, Anthropic's CLI...",    // Built-in system prompt (thousands of lines)
    "Contents of ~/.claude/CLAUDE.md...",          // Your global instructions
    "Contents of ./CLAUDE.md...",                  // Your project instructions
    "Contents of auto-memory files...",            // Things Claude remembered from past sessions
    "Available tools: [                            // MCP server tool definitions
      { name: 'search_repos',
        description: 'Search GitHub repositories',
        parameters: { query: string } },
      { name: 'create_issue',
        description: 'Create a GitHub issue',
        parameters: { title: string, body: string } },
      ...every tool from every MCP server...
    ]"
  ],
  messages: [
    { role: 'user', content: 'your first message' },
    { role: 'assistant', content: 'Claude response' },
    { role: 'user', content: 'your second message' },
    { role: 'assistant', content: 'Claude response' },
    ...the entire conversation history up to this point...
    { role: 'user', content: 'your latest message' }
  ]
}
```

**Every. Single. Message.** The CLAUDE.md, the MCP tool list, the memory, the full conversation history — it's all sent every time. The model has no persistent memory between API calls. Each call is stateless. The only reason Claude "remembers" your conversation is because the client (Claude Code) re-sends the entire history with each new message.

This is why:
- **Context windows matter** — you have a finite budget (200K tokens for Claude) and everything (system prompt + instructions + tools + conversation) has to fit
- **Long conversations get compressed** — when you approach the limit, Claude Code summarizes older messages to free up space
- **CLAUDE.md should be concise** — it's repeated in every single call
- **Too many MCP servers slow things down** — each server's tool definitions consume tokens

Codex CLI sends the same structure to OpenAI's API. Gemini CLI sends it to Google's API. The payload format differs slightly, but the architecture is identical: system instructions + tool definitions + conversation history, sent fresh with every request.

---

## What's Inside Each Tool's Home Directory

### `~/.claude/` — Claude Code

| File / Directory | Purpose |
|---|---|
| `settings.json` | Your preferences: which model to use (opus/sonnet/haiku), permission mode, MCP server definitions, effort level |
| `CLAUDE.md` | Global instructions injected into every prompt across all projects |
| `projects/` | Per-project auto-memory — things Claude remembers between sessions for specific repos |
| `plans/` | Saved implementation plans from Claude Code's plan mode |
| `tasks/` | Background task state (lock files, progress watermarks) |
| `plugins/` | Installed skills and plugins — packaged instruction sets with reference docs |
| `cache/` | Update changelog, version info |

### `~/.codex/` — OpenAI Codex CLI

| File / Directory | Purpose |
|---|---|
| `config.toml` | Model choice, MCP server definitions, project trust levels — note: **TOML format**, not JSON |
| `instructions.md` | Global instructions (same concept as CLAUDE.md) injected into every prompt |
| `vendor_imports/skills/` | Codex's skill system — curated skill packs like screenshot, sora video generation, security analysis |

### `~/.gemini/` — Google Gemini CLI

| File / Directory | Purpose |
|---|---|
| `settings.json` | MCP servers, UI theme, approval mode, session retention settings |
| `GEMINI.md` | Global instructions (same concept as CLAUDE.md) |
| `google_accounts.json` | OAuth credentials for Google authentication |
| `extensions/` | Gemini's version of plugins — these are full applications, some with React UI components |
| `extensions/github/` | Example: a complete GitHub MCP server with Go backend code, React frontend, and full UI |

### Project-Level Files

These live in your project directory (next to your code):

| File | Tool | Purpose |
|---|---|---|
| `CLAUDE.md` | Claude Code | Project-specific instructions for Claude |
| `AGENTS.md` or `agents/` | Claude Code / Codex | Defines specialized sub-agents with focused instructions |
| `codex.md` | Codex CLI | Project-specific instructions (Codex also reads AGENTS.md) |
| `GEMINI.md` | Gemini CLI | Project-specific instructions for Gemini |
| `.cursorrules` | Cursor | Project-specific instructions for Cursor |
| `.continue/` | Continue | MCP server configs and settings for Continue extension |

They all serve the same purpose: "when working in this project, follow these extra rules." The only difference is the file name each tool looks for.

---

## The Naming Confusion: Skills vs Tools vs MCP vs Extensions

Every AI coding tool invented their own names for the same concepts. Here's the translation table:

| Concept | Claude Code | Codex CLI | Gemini CLI | Cursor | What it actually is |
|---|---|---|---|---|---|
| The AI brain | Claude API | OpenAI API | Gemini API | Multiple | The LLM that reads your prompt and generates responses |
| External capabilities | MCP Servers | MCP Servers | Extensions | MCP Servers | Separate processes running on your machine that the AI can call to do things (search GitHub, read files, query databases) |
| Custom instructions | CLAUDE.md | instructions.md | GEMINI.md | .cursorrules | A text file whose contents are injected into every single prompt |
| Reusable workflows | Skills / Plugins | Skills | Extensions | — | Packaged bundles of instructions + reference docs + tool configs that teach the AI a specific workflow |
| Per-project config | CLAUDE.md (project) | codex.md | GEMINI.md (project) | .cursorrules | Same as custom instructions but scoped to one project |
| Config directory | ~/.claude/ | ~/.codex/ | ~/.gemini/ | ~/.cursor/ | Where settings, credentials, and state are stored |

**They're all the same concepts with different names.** An MCP server in Claude Code does the exact same thing as an extension in Gemini CLI — it gives the AI the ability to call an external program. A skill in Codex is the same idea as a plugin in Claude Code — a packaged set of instructions and tools.

---

## How MCP Servers Actually Work

MCP (Model Context Protocol) is the standard that all these tools are converging on. Here's what actually happens when you use an MCP server:

```
You type: "Create a GitHub issue for this bug"
    |
    v
Claude Code reads ~/.claude/settings.json
    |
    v
Sees "github" MCP server: { command: "npx", args: ["-y", "@modelcontextprotocol/server-github"] }
    |
    v
Launches that process on YOUR machine (npx runs the GitHub MCP server locally)
    |
    v
The MCP server tells Claude Code: "I have these tools: search_repos, create_issue, list_prs, ..."
    |
    v
Claude Code sends your message to Claude API with those tools listed
    |
    v
Claude decides: "I should use the create_issue tool"
    |
    v
Claude Code calls the local MCP process: create_issue(title="Bug: ...", body="...")
    |
    v
The MCP process calls the GitHub API using YOUR token
    |
    v
GitHub responds: "Issue #42 created"
    |
    v
Result sent back to Claude, who tells you: "I created issue #42"
```

Key things to understand:

- **MCP servers run on YOUR machine**, not in the cloud. They're local processes.
- **The AI model never sees your tokens or credentials directly.** The MCP server handles authentication.
- **The model just sees tool names and descriptions.** It doesn't know or care whether the tool was configured manually, through a marketplace button, or through mcpup. The config file is just a way to tell the client app "launch this process and expose its tools to the AI."
- **MCP is a protocol, not a product.** Any AI tool can support it. That's why Claude Code, Codex, Cursor, and others all use MCP — it's an open standard.

---

## Why This Matters for mcpup

Look at what your own machine has right now:

- **Claude Code** (`~/.claude/settings.json`): GitHub server using `npx2` (broken command), Playwright enabled
- **Codex CLI** (`~/.codex/config.toml`): GitHub using `npx2` (same bug), Playwright enabled, completely different file format (TOML)
- **Gemini CLI** (`~/.gemini/settings.json`): GitHub using proper `npx` command, has Brave Search that the others don't have

Three tools. Three files. Three formats. Already out of sync. The GitHub server is broken in two of them but works in the third. Brave Search exists in one but not the others.

This is the exact problem mcpup solves. One canonical config, synced to all clients, all formats handled automatically. When each tool adds their own marketplace or server browser, it makes the fragmentation worse — because each marketplace only writes to its own config file. mcpup is the bridge that keeps them all in sync.

---

## The Bottom Line

Every AI coding tool works the same way under the hood:

1. **Read instruction files** (CLAUDE.md / instructions.md / GEMINI.md) and inject them into every prompt
2. **Read MCP server configs** from a settings file, launch those processes, expose their tools to the AI
3. **Send everything** (system prompt + instructions + tools + full conversation history) to an API on every single message
4. **Get a response** and display it

The differences are just packaging: different file names, different config formats (JSON vs TOML vs YAML), different directory paths, different terminology (skills vs plugins vs extensions). The architecture is identical.

Understanding this means you understand how all of them work — and why a tool like mcpup that bridges the gaps between them is valuable.

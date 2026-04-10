# MCP servers (Cursor)

The repo root [`.mcp.json`](../.mcp.json) configures **Model Context Protocol** servers for Cursor and compatible clients. The default file ships with an empty `mcpServers` object so nothing runs until you add entries.

## Add a server

Edit `.mcp.json` and add a named server under `mcpServers`. Common patterns:

**stdio (Node)**

```json
{
  "mcpServers": {
    "fetch": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-fetch"]
    }
  }
}
```

**stdio (filesystem)** — replace the path with directories you want exposed:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/root"]
    }
  }
}
```

**Remote (HTTP/SSE)** — if your server documents a URL-based transport, use the shape your client expects (see Cursor MCP docs).

## Examples on disk

Copy from [`.mcp.example.json`](../.mcp.example.json) into `.mcp.json` and adjust names, commands, and paths.

## Product note

CentrAI Agent’s **Go** runtime MCP **client** is roadmap work ([docs/V2.md](../docs/V2.md)). This file is for **editor** MCP tooling in this repository.

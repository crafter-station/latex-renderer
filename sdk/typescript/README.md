# latex-renderer-sdk

TypeScript SDK for the LaTeX Renderer API.

[Latex Renderer Source Code](https://github.com/crafter-station/latex-renderer)

## Installation

```bash
npm install latex-renderer-sdk
```

## Usage

```typescript
import { LatexRenderer } from "latex-renderer-sdk";

const client = new LatexRenderer({
  apiKey: "your-api-key",
  baseUrl: "https://your-api-url.com", // optional, defaults to http://localhost:8080
});
```

### Render LaTeX to HTML

```typescript
const html = await client.renderHTML(
  "\\documentclass{article}\\begin{document}Hello $E=mc^2$\\end{document}",
);
```

### Render LaTeX to PDF

```typescript
import { writeFile } from "fs/promises";

const pdf = await client.renderPDF(
  "\\documentclass{article}\\begin{document}Hello World\\end{document}",
);

await writeFile("output.pdf", pdf);
```

### Error handling

```typescript
import {
  LatexRenderer,
  RenderError,
  AuthenticationError,
  ConnectionError,
} from "latex-renderer-sdk";

try {
  const html = await client.renderHTML(latex);
} catch (error) {
  if (error instanceof RenderError) {
    console.error("LaTeX compilation failed:", error.detail);
  } else if (error instanceof AuthenticationError) {
    console.error("Invalid API key:", error.message);
  } else if (error instanceof ConnectionError) {
    console.error("Network error:", error.message);
  }
}
```

### Request cancellation

```typescript
const controller = new AbortController();
setTimeout(() => controller.abort(), 5000);

const html = await client.renderHTML(latex, { signal: controller.signal });
```

## Configuration

| Option    | Type     | Default                  | Description             |
| --------- | -------- | ------------------------ | ----------------------- |
| `apiKey`  | `string` | _(required)_             | API authentication key  |
| `baseUrl` | `string` | `http://localhost:8080`  | API base URL            |
| `timeout` | `number` | `30000`                  | Request timeout (ms)    |

## Publishing to npm

### First time setup

1. Create a **Granular Access Token** at https://www.npmjs.com/settings/YOUR_USERNAME/tokens
   - Click "Generate New Token" > "Granular Access Token"
   - Set permissions: **Read and Write** for packages
2. Login with the token:

```bash
npm login
```

Or set it directly:

```bash
npm config set //registry.npmjs.org/:_authToken=YOUR_TOKEN
```

### First publish

```bash
npm run publish:first
```

### Update and republish

```bash
npm run publish:patch   # 1.0.0 → 1.0.1 (bug fixes)
npm run publish:minor   # 1.0.0 → 1.1.0 (new features)
npm run publish:major   # 1.0.0 → 2.0.0 (breaking changes)
```

## Requirements

Node.js >= 18.0.0

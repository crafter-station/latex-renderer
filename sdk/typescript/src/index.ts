import type { LatexRendererConfig, RenderOptions } from "./types.js";
import {
  LatexRendererError,
  AuthenticationError,
  RenderError,
  APIError,
  ConnectionError,
} from "./errors.js";

export {
  LatexRendererError,
  AuthenticationError,
  RenderError,
  APIError,
  ConnectionError,
} from "./errors.js";
export type { LatexRendererConfig, RenderOptions } from "./types.js";

export class LatexRenderer {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly timeout: number;

  constructor(config: LatexRendererConfig) {
    if (!config.apiKey) {
      throw new LatexRendererError("apiKey is required");
    }
    this.apiKey = config.apiKey;
    this.baseUrl = (config.baseUrl ?? "http://localhost:8080").replace(
      /\/$/,
      "",
    );
    this.timeout = config.timeout ?? 30_000;
  }

  async renderHTML(latex: string, options?: RenderOptions): Promise<string> {
    const response = await this.request("/render", latex, options);
    return response.text();
  }

  async renderPDF(
    latex: string,
    options?: RenderOptions,
  ): Promise<Uint8Array> {
    const response = await this.request("/render/pdf", latex, options);
    const buffer = await response.arrayBuffer();
    return new Uint8Array(buffer);
  }

  private async request(
    endpoint: string,
    body: string,
    options?: RenderOptions,
  ): Promise<Response> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "text/plain",
        },
        body,
        signal: options?.signal ?? controller.signal,
      });

      if (!response.ok) {
        await this.handleError(response);
      }

      return response;
    } catch (error) {
      if (error instanceof LatexRendererError) throw error;

      if (error instanceof Error && error.name === "AbortError") {
        throw new ConnectionError("Request timed out or was cancelled", error);
      }

      throw new ConnectionError(
        error instanceof Error ? error.message : "Unknown network error",
        error instanceof Error ? error : undefined,
      );
    } finally {
      clearTimeout(timeoutId);
    }
  }

  private async handleError(response: Response): Promise<never> {
    let json: { error?: string; detail?: string } | undefined;

    try {
      json = await response.json();
    } catch {
      throw new APIError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status,
      );
    }

    const message = json?.error ?? "Unknown error";

    if (response.status === 401) {
      throw new AuthenticationError(message);
    }

    if (
      response.status === 400 &&
      (message === "latex render failed" || message === "pdf render failed")
    ) {
      throw new RenderError(message, json?.detail);
    }

    throw new APIError(message, response.status);
  }
}

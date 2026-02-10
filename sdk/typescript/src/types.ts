export interface LatexRendererConfig {
  apiKey: string;
  baseUrl?: string;
  timeout?: number;
}

export interface RenderOptions {
  signal?: AbortSignal;
}

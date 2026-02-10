export class LatexRendererError extends Error {
  public readonly statusCode?: number;

  constructor(message: string, statusCode?: number) {
    super(message);
    this.name = "LatexRendererError";
    this.statusCode = statusCode;
  }
}

export class AuthenticationError extends LatexRendererError {
  constructor(message: string) {
    super(message, 401);
    this.name = "AuthenticationError";
  }
}

export class RenderError extends LatexRendererError {
  public readonly detail?: string;

  constructor(message: string, detail?: string) {
    super(message, 400);
    this.name = "RenderError";
    this.detail = detail;
  }
}

export class APIError extends LatexRendererError {
  constructor(message: string, statusCode: number) {
    super(message, statusCode);
    this.name = "APIError";
  }
}

export class ConnectionError extends LatexRendererError {
  public readonly cause?: Error;

  constructor(message: string, cause?: Error) {
    super(message);
    this.name = "ConnectionError";
    this.cause = cause;
  }
}

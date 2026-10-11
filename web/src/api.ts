import type { Problem } from "./types";

export class ApiError extends Error {
  problem: Problem;
  constructor(problem: Problem) {
    super(problem.code);
    this.problem = problem;
  }
}

export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`/api${path}`, {
      ...options,
      headers: {
        ...(options.body instanceof FormData
          ? {}
          : { "Content-Type": "application/json" }),
        ...options.headers,
      },
    });
  } catch {
    throw new ApiError({ code: "network" });
  }
  if (!response.ok) {
    let problem: Problem;
    try {
      problem = (await response.json()) as Problem;
    } catch {
      problem = { code: "internal" };
    }
    throw new ApiError(problem);
  }
  return response.json() as Promise<T>;
}

export function json(method: string, body: unknown): RequestInit {
  return { method, body: JSON.stringify(body) };
}
export function problemOf(error: unknown): Problem {
  return error instanceof ApiError ? error.problem : { code: "internal" };
}

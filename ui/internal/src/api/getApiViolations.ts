import { AxiosError } from "axios";

type ApiViolations = Record<string, string>;

type ApiResponseData = {
  message?: string;
} & ApiViolations;

export function getApiViolations(error: unknown): ApiViolations | null {
  if (!(error instanceof AxiosError)) {
    return null;
  }

  const responseData = error.response?.data;
  if (!responseData || typeof responseData !== "object") {
    return null;
  }

  const violations: ApiViolations = {};

  for (const [field, value] of Object.entries(responseData as ApiResponseData)) {
    if (field === "message") {
      continue;
    }

    if (typeof value === "string") {
      violations[field] = value;
    }
  }

  if (Object.keys(violations).length === 0) {
    return null;
  }

  return violations;
}

export function getApiMessage(error: unknown): string | null {
  if (!(error instanceof AxiosError)) {
    return null;
  }

  const message = error.response?.data?.message;
  if (typeof message === "string" && message.length > 0) {
    return message;
  }

  return null;
}

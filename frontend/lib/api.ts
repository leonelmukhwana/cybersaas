const API_URL =
  process.env.NEXT_PUBLIC_API_URL ||
  "http://localhost:8080";

interface ApiOptions extends RequestInit {
  token?: string;
}

function getStoredToken(): string | null {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    const stored = localStorage.getItem(
      "cybersaas-auth"
    );

    if (!stored) {
      return null;
    }

    const parsed = JSON.parse(stored);

    if (
      parsed &&
      typeof parsed === "object" &&
      "state" in parsed &&
      parsed.state &&
      typeof parsed.state === "object" &&
      typeof parsed.state.token === "string"
    ) {
      return parsed.state.token;
    }

    return null;
  } catch {
    return null;
  }
}

export async function api<T>(
  endpoint: string,
  options: ApiOptions = {}
): Promise<T> {
  const {
    token: providedToken,
    headers,
    ...fetchOptions
  } = options;

  const token =
    providedToken || getStoredToken();

  const response = await fetch(
    `${API_URL}${endpoint}`,
    {
      ...fetchOptions,
      headers: {
        "Content-Type": "application/json",

        ...(token
          ? {
              Authorization: `Bearer ${token}`,
            }
          : {}),

        ...headers,
      },
    }
  );

  const rawResponse =
    await response.text();

  let data: unknown = null;

  if (rawResponse) {
    try {
      data = JSON.parse(rawResponse);
    } catch {
      data = rawResponse;
    }
  }

  console.log("API RESPONSE:", {
    endpoint,
    status: response.status,
    data,
  });

  if (!response.ok) {
    const message =
      typeof data === "object" &&
      data !== null &&
      "error" in data &&
      typeof data.error === "string"
        ? data.error
        : typeof data === "object" &&
            data !== null &&
            "message" in data &&
            typeof data.message === "string"
          ? data.message
          : typeof data === "string" &&
              data.length > 0
            ? data
            : `Request failed with status ${response.status}.`;

    throw new Error(message);
  }

  return data as T;
}
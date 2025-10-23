import { http, HttpResponse, type HttpHandler } from "msw";
import type { APIKey } from "../";
import { apiKeys, findIndexByName, shortKeyFrom } from "./state";

function randomRawKey(): string {
  const bytes = new Uint8Array(24);
  crypto.getRandomValues(bytes);
  return Array.from(bytes).map((b) => b.toString(16).padStart(2, "0")).join("");
}

// Create handlers for a given base prefix, e.g. "/mocked-api-keys"
export function createHandlers(basePrefix: string): HttpHandler[] {
  const base = basePrefix.replace(/\/+$/, "");

  return [
    // GET /api-keys → limit to 10 items
    http.get(`${base}/api-keys`, () => {
      return HttpResponse.json({ api_keys: apiKeys.api_keys.slice(0, 10) });
    }),

    // Create
    http.post(`${base}/api-keys`, async ({ request }) => {
      const body = (await request.json()) as { name?: string };
      const name = (body.name || "").trim();

      if (!/^[A-Za-z0-9._-]{3,35}$/.test(name)) {
        return HttpResponse.json({ error: "invalid name" }, { status: 400 });
      }
      if (findIndexByName(name) !== -1) {
        return HttpResponse.json({ error: "name already exists" }, { status: 409 });
      }

      const raw = randomRawKey();
      const now = new Date().toISOString();
      const created: APIKey = {
        name,
        short_key: shortKeyFrom(raw),
        created_at: now,
        last_used_at: null,
        disabled_at: null,
      };
      apiKeys.api_keys.unshift(created);

      // Return both camel and snake variants if your schema differs; harmless for UI not using it yet
      return HttpResponse.json({ raw_api_key: raw, api_key: created });
    }),

    // Get by name
    http.get(`${base}/api-keys/:name`, ({ params }) => {
      const idx = findIndexByName(String(params.name || ""));
      if (idx === -1) return HttpResponse.json({ error: "not found" }, { status: 404 });
      return HttpResponse.json({ api_key: apiKeys.api_keys[idx] });
    }),

    // Delete
    http.delete(`${base}/api-keys/:name`, ({ params }) => {
      const idx = findIndexByName(String(params.name || ""));
      if (idx === -1) return HttpResponse.json({ error: "not found" }, { status: 404 });
      apiKeys.api_keys.splice(idx, 1);
      return new HttpResponse(null, { status: 204 });
    }),

    // Disable
    http.post(`${base}/api-keys/:name/disable`, ({ params }) => {
      const idx = findIndexByName(String(params.name || ""));
      if (idx === -1) return HttpResponse.json({ error: "not found" }, { status: 404 });
      apiKeys.api_keys[idx].disabled_at = new Date().toISOString();
      return HttpResponse.json({ api_key: apiKeys.api_keys[idx] });
    }),

    // Enable
    http.post(`${base}/api-keys/:name/enable`, ({ params }) => {
      const idx = findIndexByName(String(params.name || ""));
      if (idx === -1) return HttpResponse.json({ error: "not found" }, { status: 404 });
      apiKeys.api_keys[idx].disabled_at = null;
      return HttpResponse.json({ api_key: apiKeys.api_keys[idx] });
    }),
  ];
}
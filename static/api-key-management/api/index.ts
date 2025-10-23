import createClient from "openapi-fetch";
import type { paths, components } from "./schema";

export type APIKey = components["schemas"]["APIKey"];
export type CreateAPIKeyBody = components["schemas"]["CreateAPIKeyBody"];
export type CreateAPIKeyData = components["schemas"]["CreateAPIKeyData"];
export type GetAPIKeysData = components["schemas"]["GetAPIKeysData"];
export type GetAPIKeyData = components["schemas"]["GetAPIKeyData"];
export type DisableAPIKeyData = components["schemas"]["DisableAPIKeyData"];
export type EnableAPIKeyData = components["schemas"]["EnableAPIKeyData"];

export function createAPIClient(baseURL: string) {
  const baseUrl = (baseURL || "").replace(/\/+$/, "");
  const client = createClient<paths>({ baseUrl });

  return {
    // Raw typed client (GET/POST/etc) if you ever need it
    raw: client,

    // GET /api-keys → APIKey[]
    async getAPIKeys(): Promise<GetAPIKeysData> {
      const { data, error } = await client.GET("/api-keys");
      if (error) throw error;
      return data;
    },

    // POST /api-keys { name } → { raw_api_key, api_key }
    async createAPIKey(body: CreateAPIKeyBody): Promise<CreateAPIKeyData> {
      const { data, error } = await client.POST("/api-keys", { body });
      if (error) throw error;
      return data;
    },

    // GET /api-keys/{name} → APIKey | null
    async getAPIKey(name: string): Promise<APIKey | null> {
      const { data, error } = await client.GET("/api-keys/{name}", {
        params: { path: { name } },
      });
      if (error) throw error;
      return data?.api_key ?? null;
    },

    // DELETE /api-keys/{name} → void (204)
    async deleteAPIKey(name: string): Promise<void> {
      const { error } = await client.DELETE("/api-keys/{name}", {
        params: { path: { name } },
      });
      if (error) throw error;
    },

    // POST /api-keys/{name}/disable → APIKey
    async disableAPIKey(name: string): Promise<APIKey> {
      const { data, error } = await client.POST("/api-keys/{name}/disable", {
        params: { path: { name } },
      });
      if (error) throw error;
      return data.api_key!;
    },

    // POST /api-keys/{name}/enable → APIKey
    async enableAPIKey(name: string): Promise<APIKey> {
      const { data, error } = await client.POST("/api-keys/{name}/enable", {
        params: { path: { name } },
      });
      if (error) throw error;
      return (data as EnableAPIKeyData).api_key!;
    },
  };
}

export type ApiKeysClient = ReturnType<typeof createAPIClient>;
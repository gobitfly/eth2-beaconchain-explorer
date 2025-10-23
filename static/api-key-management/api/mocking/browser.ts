import { setupWorker } from "msw/browser";
import { createHandlers } from "./handlers";

// Start MSW for the given base prefix (e.g., "/mocked-api-keys" or "http://localhost:8088/mocked-api-keys")
export async function startMocks(basePrefix: string) {
  const worker = setupWorker(...createHandlers(basePrefix));
  await worker.start({
    onUnhandledRequest: "bypass",
    serviceWorker: { url: "/mockServiceWorker.js" }, // created by `msw init ./static`
  });
  return worker;
}
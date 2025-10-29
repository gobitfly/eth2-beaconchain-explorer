import { createAPIClient } from './api';
import { initApiKeysTable } from "./list-keys";
import { useAPIKeyCreation } from "./create-keys";
import { startMocks } from "./api/mocking/browser";

declare global {
  interface Window {
    __APIKEYMANAGEMENT_CONFIG__: {
      apiBaseUrl: string;
    };
  }
}

const APiBaseUrl = window.__APIKEYMANAGEMENT_CONFIG__.apiBaseUrl;

const ApiClient = createAPIClient(APiBaseUrl);

function initializeApiKeyManagement() {
  const apiKeysTables = initApiKeysTable(ApiClient);

  useAPIKeyCreation(ApiClient, {
    onCreated: () => {
      apiKeysTables?.ajax.reload()
    },
  });
  
  document.querySelector("#simmulateReload")?.addEventListener("click", () => {
    apiKeysTables?.ajax.reload();
  });
}

(async () => {
  // Start MSW unconditionally (temporary)
  await startMocks(APiBaseUrl);

  initializeApiKeyManagement();

})();

// document.addEventListener("DOMContentLoaded", () => {
//   initializeApiKeyManagement();
// });
import { createAPIClient } from './api';
import { initApiKeysTable } from "./list";
import { useAPIKeyCreation } from "./create";
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
  const dt = initApiKeysTable(ApiClient);

  useAPIKeyCreation(ApiClient, {
    onCreated: () => {
      dt?.ajax.reload()
    },
  });
  
  document.querySelector("#simmulateReload")?.addEventListener("click", () => {
    dt?.ajax.reload();
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
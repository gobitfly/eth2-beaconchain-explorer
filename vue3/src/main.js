import { createPinia } from "pinia"
import { createSSRApp } from "vue"
import App from "./app.vue"
import createRouter from "@/router"

import tooltip from "@/directives/tooltip.js"

export function createApp() {
  const app = createSSRApp(App)
  app.directive("tooltip", tooltip)
  const store = createPinia()
  app.use(store)
  const router = createRouter()
  app.use(router)
  return { app, router, store }
}

import { renderToString } from "vue/server-renderer"
import { createApp } from "./main"
import { useRoute } from "vue-router"
import { computed } from "vue"

export async function render(url, manifest) {
  const { app, router, store } = createApp()
  try {
    await router.push(url)
    await router.isReady()

    const to = router.currentRoute

    const matchedRoute = to.value.matched
    if (to.value.matched.length === 0) {
      return ""
    }

    const matchedComponents = []
    matchedRoute.map((route) => {
      matchedComponents.push(...Object.values(route.components))
    })

    const asyncDataFuncs = matchedComponents.map((component) => {
      const asyncData = component.asyncData || null
      if (asyncData) {
        const config = {
          route: to,
        }
        return asyncData(config)
      }
    })

    await Promise.all(asyncDataFuncs)

    const ctx = {}
    const html = await renderToString(app, ctx)
    const preloadLinks = renderPreloadLinks(ctx.modules, manifest)
    const teleports = renderTeleports(ctx.teleports)
    const state = JSON.stringify(store.state.value)

    return [html, state, preloadLinks, teleports]
  } catch (error) {
    console.log(error)
  }
}

function renderPreloadLinks(modules, manifest) {
  let links = ""
  const seen = new Set()
  modules.forEach((id) => {
    const files = manifest[id]
    if (files) {
      files.forEach((file) => {
        if (!seen.has(file)) {
          seen.add(file)
          links += renderPreloadLink(file)
        }
      })
    }
  })
  return links
}

function renderPreloadLink(file) {
  if (file.endsWith(".js")) {
    return `<link rel="modulepreload" crossorigin href="${file}">`
  } else if (file.endsWith(".css")) {
    return `<link rel="stylesheet" href="${file}">`
  } else {
    return ""
  }
}

function renderTeleports(teleports) {
  if (!teleports) return ""
  return Object.entries(teleports).reduce((all, [key, value]) => {
    if (key.startsWith("#el-popper-container-")) {
      return `${all}<div id="${key.slice(1)}">${value}</div>`
    }
    return all
  }, teleports.body || "")
}

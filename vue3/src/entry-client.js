import { createApp } from "./main"

import { useRoute } from "vue-router"
import { computed } from "vue"

import HighchartsVue from "highcharts-vue"

const { app, router, store } = createApp()

app.use(HighchartsVue)

router.beforeResolve((to, from, next) => {
  let diffed = false
  const matched = router.resolve(to).matched
  const prevMatched = router.resolve(from).matched

  console.log(from, to)

  if (from && !from.name) {
    return next()
  } else {
    window.document.title = to.meta.title || "beaconcha.in"
  }

  const activated = matched.filter((c, i) => {
    return diffed || (diffed = prevMatched[i] !== c)
  })

  if (!activated.length) {
    return next()
  }

  const matchedComponents = []
  matched.map((route) => {
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
  try {
    Promise.all(asyncDataFuncs).then(() => {
      next()
    })
  } catch (err) {
    next(err)
  }
})

if (window.__INITIAL_STATE__) {
  console.log(window.__INITIAL_STATE__)
  store.state.value = JSON.parse(JSON.stringify(window.__INITIAL_STATE__))
}

router.isReady().then(() => {
  app.mount("#app")
})

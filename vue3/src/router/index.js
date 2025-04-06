import { createRouter, createWebHistory, createMemoryHistory } from "vue-router"

export default function () {
  const routerHistory = import.meta.env.SSR === false ? createWebHistory() : createMemoryHistory()

  return createRouter({
    history: routerHistory,
    routes: [
      {
        path: "/",
        name: "index",
        component: () => import("@/views/index.vue"),
        meta: {
          title: "Open Source Ethereum (ETH) Testnet Explorer - beaconcha.in - 2023",
        },
      },
      {
        path: "/:pathMatch(.*)*",
        name: "catchall",
        component: () => import("@/views/catchall.vue"),
        meta: {
          title: "Open Source Ethereum (ETH) Testnet Explorer - beaconcha.in - 2023",
        },
      },
    ],
  })
}

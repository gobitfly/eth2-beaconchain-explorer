import fs from "node:fs"
import path from "node:path"
import express from "express"
import Axios from "axios"
import { fileURLToPath } from "node:url"
import adapter from "axios/lib/adapters/http.js"
import serveStatic from "serve-static"

Axios.defaults.adapter = adapter

const BACKEND_URL = "http://localhost:8080"

const isTest = process.env.NODE_ENV === "test" || !!process.env.VITE_TEST_BUILD
const isProduction = process.env.NODE_ENV === "production"
export async function createServer(root = process.cwd(), isProd = isProduction) {
  const __dirname = path.dirname(fileURLToPath(import.meta.url))
  const resolve = (p) => path.resolve(__dirname, p)
  const indexProd = isProd ? fs.readFileSync(resolve("dist/client/index.html"), "utf-8") : ""
  const manifest = isProd ? JSON.parse(fs.readFileSync(resolve("dist/client/ssr-manifest.json"), "utf-8")) : {}

  const app = express()

  let vite
  if (!isProd) {
    vite = await (
      await import("vite")
    ).createServer({
      root,
      logLevel: isTest ? "error" : "info",
      server: {
        middlewareMode: true,
        watch: {
          usePolling: true,
          interval: 100,
        },
      },
      appType: "custom",
    })
    // use vite's connect instance as middleware
    app.use(vite.middlewares)
  } else {
    app.use((await import("compression")).default())
    app.use(
      (await import("serve-static")).default(resolve("dist/client"), {
        index: false,
      })
    )
  }

  app.use("/", serveStatic(path.join(__dirname, "../static/")))

  app.use("/mock/getPageData", async (req, res, next) => {
    try {
      const { data } = await Axios.get(BACKEND_URL + "/index/pageData", { responseType: "stream" })
      res.set("Access-Control-Allow-Origin", "*")
      data.pipe(res)
    } catch (err) {
      console.log("error", err)
      return next(err)
    }
  })

  app.use("/mock/indexData", async (req, res, next) => {
    try {
      const { data } = await Axios.get(BACKEND_URL + "/index/data", { responseType: "stream" })
      res.set("Access-Control-Allow-Origin", "*")
      data.pipe(res)
    } catch (err) {
      console.log("error", err)
      return next(err)
    }
  })

  app.use("/mock/launchMetrics", async (req, res, next) => {
    try {
      const { data } = await Axios.get(BACKEND_URL + "/launchMetrics", { responseType: "stream" })
      res.set("Access-Control-Allow-Origin", "*")
      data.pipe(res)
    } catch (err) {
      console.log("error", err)
      return next(err)
    }
  })

  app.use("/mock/gitcoinfeed", async (req, res, next) => {
    try {
      const { data } = await Axios.get(BACKEND_URL + "/gitcoinfeed", { responseType: "stream" })
      res.set("Access-Control-Allow-Origin", "*")
      data.pipe(res)
    } catch (err) {
      console.log("error", err)
      return next(err)
    }
  })

  app.use("/mock/latestState", async (req, res, next) => {
    try {
      const { data } = await Axios.get(BACKEND_URL + "/latestState", { responseType: "stream" })
      res.set("Access-Control-Allow-Origin", "*")
      data.pipe(res)
    } catch (err) {
      console.log("error", err)
      return next(err)
    }
  })

  app.use("*", async (req, res) => {
    try {
      const url = req.originalUrl

      let template, render
      if (!isProd) {
        // always read fresh template in dev
        template = fs.readFileSync(resolve("index.html"), "utf-8")
        template = await vite.transformIndexHtml(url, template)
        render = (await vite.ssrLoadModule("/src/entry-server")).render
      } else {
        template = indexProd
        render = (await import("./dist/server/entry-server.js")).render
      }

      const [appHtml, state, links, teleports] = await render(url, manifest)

      console.log(url)

      const html = template
        .replace(`<!--preload-links-->`, links)
        .replace(`"<pinia-store>"`, state)
        .replace(`<!--app-html-->`, appHtml)
        .replace("<!--title-->", "Index")
        .replace(/(\n|\r\n)\s*<!--app-teleports-->/, teleports)

      res.status(200).set({ "Content-Type": "text/html" }).end(html)
    } catch (e) {
      vite && vite.ssrFixStacktrace(e)
      console.log(e.stack)
      res.status(500).end(e.stack)
    }
  })

  return { app, vite }
}

if (!isTest) {
  createServer().then(({ app }) =>
    app.listen(8888, () => {
      console.log("http://localhost:8888")
    })
  )
}

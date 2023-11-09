import path from "node:path"
import { resolve } from "path"
import { defineConfig, loadEnv, searchForWorkspaceRoot } from "vite"
import vuePlugin from "@vitejs/plugin-vue"
import vueJsx from "@vitejs/plugin-vue-jsx"
// https://www.npmjs.com/package/vite-svg-loader
import svgLoader from "vite-svg-loader"

const base = "/"

// preserve this to test loading __filename & __dirname in ESM as Vite polyfills them.
// if Vite incorrectly load this file, node.js would error out.
globalThis.__vite_test_filename = __filename
globalThis.__vite_test_dirname = __dirname

export default defineConfig(({ command, ssrBuild }) => ({
  base,
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
      "#": resolve(__dirname, ".."),
    },
  },
  plugins: [
    vuePlugin(),
    vueJsx(),
    svgLoader({
      defaultImport: "component", // or 'url' or 'component'
      svgoConfig: {
        multipass: true,
      },
    }),
  ],
  server: {
    fs: {
      // Allow serving files from one level up to the project root
      allow: [
        // search up for workspace root
        searchForWorkspaceRoot(process.cwd()),
        // your custom rules
        "..",
      ],
    },
  },
  build: {
    minify: false,
  },
  ssr: {
    noExternal: [],
  },
  optimizeDeps: {},
}))

import react from "@vitejs/plugin-react"
import fs from "node:fs"
import path from "node:path"
import { defineConfig, loadEnv } from "vite"

const srcDir = path.resolve(import.meta.dirname, "src")

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, import.meta.dirname, ["REACT_APP_", "VITE_"])
  const httpTarget = process.env.MC_PROXY_HTTP || "http://localhost:8080"
  const wsTarget = process.env.MC_PROXY_WEBSOCKET || "ws://localhost:8080"

  const processEnv = {
    "process.env.NODE_ENV": JSON.stringify(mode === "production" ? "production" : "development"),
    "process.env.PUBLIC_URL": JSON.stringify(""),
  }
  for (const [key, value] of Object.entries(env)) {
    processEnv[`process.env.${key}`] = JSON.stringify(value)
  }
  if (process.env.REACT_APP_IS_DEV_ENV !== undefined) {
    processEnv["process.env.REACT_APP_IS_DEV_ENV"] = JSON.stringify(process.env.REACT_APP_IS_DEV_ENV)
  }

  return {
    plugins: [
      react({
        include: "**/*.{jsx,js}",
      }),
    ],
    envPrefix: ["REACT_APP_", "VITE_"],
    define: processEnv,
    publicDir: "public",
    esbuild: {
      loader: "jsx",
      include: /src\/.*\.jsx?$/,
      exclude: [],
    },
    optimizeDeps: {
      esbuildOptions: {
        plugins: [
          {
            name: "load-js-files-as-jsx",
            setup(build) {
              build.onLoad({ filter: /\/src\/.*\.js$/ }, async (args) => ({
                loader: "jsx",
                contents: await fs.promises.readFile(args.path, "utf8"),
              }))
            },
          },
        ],
      },
    },
    css: {
      preprocessorOptions: {
        scss: {
          loadPaths: [srcDir],
        },
      },
    },
    server: {
      port: 3000,
      strictPort: true,
      proxy: {
        "/api/ws": {
          target: wsTarget,
          changeOrigin: true,
          ws: true,
        },
        "/api": {
          target: httpTarget,
          changeOrigin: true,
          cookieDomainRewrite: "localhost",
        },
        "/secure_share": {
          target: httpTarget,
          changeOrigin: true,
        },
        "/insecure_share": {
          target: httpTarget,
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: "build",
      emptyOutDir: true,
      sourcemap: false,
      chunkSizeWarningLimit: 2000,
    },
  }
})

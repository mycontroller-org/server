const { createProxyMiddleware } = require("http-proxy-middleware")

module.exports = function (app) {
  const wsTarget = getEnvironmentValue("MC_PROXY_WEBSOCKET", "ws://localhost:8080")
  const httpTarget = getEnvironmentValue("MC_PROXY_HTTP", "http://localhost:8080")
  const logLevel = getEnvironmentValue("MC_PROXY_LOG_LEVEL", "debug")

  app.use(
    "/api/ws",
    createProxyMiddleware({
      target: wsTarget,
      changeOrigin: true,
      ws: true,
      logLevel: logLevel,
    })
  )
  app.use(
    "/api",
    createProxyMiddleware({
      target: httpTarget,
      changeOrigin: true,
      logLevel: logLevel,
      cookieDomainRewrite: 'localhost',
    })
  )
  app.use(
    "/secure_share",
    createProxyMiddleware({
      target: httpTarget,
      changeOrigin: true,
      logLevel: logLevel,
    })
  )
  app.use(
    "/insecure_share",
    createProxyMiddleware({
      target: httpTarget,
      changeOrigin: true,
      logLevel: logLevel,
    })
  )
}

const getEnvironmentValue = (varName, defaultValue) => {
  const value = process.env[varName]
  if (value !== undefined) {
    return value
  }
  return defaultValue
}

import FaroSourceMapUploaderPlugin from '@grafana/faro-webpack-plugin'
import path from 'path'

const token = process.env.FARO_SOURCEMAP_TOKEN

const gitSha =
  process.env.BUILDKITE_COMMIT || process.env.GIT_SHA || 'static-id'

/** @type {import('next').NextConfig} */
const nextConfig = {
  // output: 'export',
  output: 'standalone', // Creates a minimal server for Docker
  basePath: '/v2',
  deploymentId: gitSha,

  // Set workspace root to avoid Next.js inference warnings
  // This is needed because we're in a monorepo with lockfiles at multiple levels
  // Even though Turbopack isn't enabled, this config helps Next.js correctly identify
  // the workspace root for dependency resolution and module resolution
  turbopack: {
    root: process.cwd(),
  },

  // Enable source maps in production
  productionBrowserSourceMaps: true,

  pageExtensions: ['js', 'jsx', 'ts', 'tsx'], // https://nextjs.org/docs/pages/api-reference/config/next-config-js/pageExtensions

  outputFileTracingIncludes: {
    '/**/*': ['./src/app/**/*.tsx'],
  },

  // Do not set process env here. They are set at build time, not runtime.
  // Kube and docker will set the env vars at runtime in production.
  // register within instrumentation.js
  env: {
    GIT_SHA: gitSha,
  },

  generateBuildId: async () => {
    return gitSha
  },

  // assetPrefix: isProd ? "/v2" : undefined,

  // Needed for SOC2 compliance
  // https://www.iothreat.com/blog/server-leaks-information-via-x-powered-by-http-response-header-field-s
  poweredByHeader: false,

  devIndicators: process.env.BUILDKITE ? false : { position: 'bottom-right' },

  async headers() {
    return [
      {
        // Disable caching for Monaco editor assets.
        // Monaco's worker initialization is sensitive to how cached assets are served.
        // Cached assets can load in different timing/order, causing race conditions
        // that lead to initialization hangs. Disabling cache ensures consistent,
        // predictable loading behavior.
        source: '/v2/monaco/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0',
          },
          {
            key: 'Pragma',
            value: 'no-cache',
          },
          {
            key: 'Expires',
            value: '0',
          },
        ],
      },
    ]
  },

  webpack(config, { dev, isServer, webpack }) {
    const isBuildKite = !!process.env.BUILDKITE
    if (!dev && !isServer && isBuildKite) {
      config.plugins.push(
        new FaroSourceMapUploaderPlugin({
          appName: 'union-v2-app',
          endpoint:
            'https://faro-api-prod-us-central-0.grafana.net/faro/api/v1',
          apiKey: token,
          appId: 2436,
          stackId: 400599,
          // nextjs: true prepends "_next/" to source map file paths, but with
          // basePath "/v2" assets are served at "/v2/_next/static/chunks/...".
          // We handle the path fixup ourselves below.
          nextjs: false,
          bundleId: gitSha,
          outputPath: path.join(process.cwd(), '.next'),
          outputFiles: /\.js\.map$/,
          recursive: true,
        }),
      )

      // Patch source map `file` fields to match production URLs.
      // Stack traces show "https://host/v2/_next/static/chunks/foo.js", so the
      // `file` field must be "v2/_next/static/chunks/foo.js" for Grafana to match.
      // We derive the correct path from the asset's output location rather than
      // guessing from the original `file` value (which also lacks the static/chunks/ dir).
      config.plugins.push({
        apply(compiler) {
          compiler.hooks.compilation.tap('FaroSourceMapBasePath', (compilation) => {
            compilation.hooks.processAssets.tap(
              {
                name: 'FaroSourceMapBasePath',
                stage: webpack.Compilation.PROCESS_ASSETS_STAGE_SUMMARIZE,
              },
              (assets) => {
                const { RawSource } = compiler.webpack.sources
                Object.keys(assets).forEach((assetFilename) => {
                  if (!assetFilename.endsWith('.map')) return
                  const asset = compilation.getAsset(assetFilename)
                  const content = asset?.source.source().toString()
                  if (!content) return
                  const sourceMap = JSON.parse(content)
                  // assetFilename: "static/chunks/foo.js.map"
                  // production URL: "/v2/_next/static/chunks/foo.js"
                  const rawBasePath = nextConfig.basePath || ''
                  const normalizedBasePath = rawBasePath.replace(/^\/+|\/+$/g, '')
                  const basePathPrefix = normalizedBasePath ? `${normalizedBasePath}/` : ''
                  sourceMap.file = `${basePathPrefix}_next/${assetFilename.replace(/\.map$/, '')}`
                  compilation.updateAsset(assetFilename, new RawSource(JSON.stringify(sourceMap)))
                })
              },
            )
          })
        },
      })
    }

    return config
  },
}

export default nextConfig

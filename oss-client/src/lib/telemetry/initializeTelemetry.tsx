'use client'
import {
  ErrorsInstrumentation,
  getWebInstrumentations,
  initializeFaro,
  LogLevel,
  ReactIntegration,
} from '@grafana/faro-react'
import { TracingInstrumentation } from '@grafana/faro-web-tracing'
import { getLocation, isBrowser as getIsBrowser } from '../windowUtils'

const isProd = process.env.NODE_ENV === 'production'
const isBrowser = getIsBrowser()

export const initializeTelemetry = isBrowser
  ? initializeFaro({
      ...(isProd && {
        url: 'https://faro-collector-prod-us-central-0.grafana.net/collect/51d1e26901a45e7fe01d13c2a01080a9',
      }),

      ...(!isProd && {
        // Disable all network transports in development so nothing is sent to Grafana
        transports: [],
      }),

      app: {
        name: 'union-v2-app',
        environment: process.env.UNION_PLATFORM_ENV,
        release: process.env.GIT_SHA,
        version: process.env.GIT_SHA,
      },
      isolate: true,

      ignoreUrls: [
        // exclude NewRelic calls
        /.*mt.auryc.com.*/,
      ],

      instrumentations: [
        // Mandatory, overwriting the instrumentations array would cause the default instrumentations to be omitted
        ...getWebInstrumentations({
          captureConsole: true,
          enablePerformanceInstrumentation: true,
        }),
        new TracingInstrumentation(),
        new ErrorsInstrumentation(),
        new ReactIntegration(),
      ],
      consoleInstrumentation: {
        consoleErrorAsLog: true,
        disabledLevels: [LogLevel.DEBUG, LogLevel.TRACE],
      },

      eventDomain: getLocation().hostname,
      trackResources: true,
    })
  : undefined

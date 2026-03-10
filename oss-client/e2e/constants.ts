export const BASE_URL = `https://${process.env.DEV_HOST || `localhost.${process.env.HOST}`}:${process.env.DEV_PORT || 8080}/v2`
export const E2E_PROJECT_ORG = `domain/${process.env.TEST_DOMAIN}/project/${process.env.TEST_PROJECT}`
export const HOST = `https://${process.env.HOST}`
export const E2E_PROJECT = process.env.TEST_PROJECT || ''
export const E2E_ORG = process.env.TEST_DOMAIN || ''

// https://nextjs.org/docs/app/api-reference/file-conventions/instrumentation
// https://github.com/vercel/next.js/discussions/16205#discussioncomment-9360516
import { makeEnvPublic } from 'next-runtime-env'

export const register = async () => {
  makeEnvPublic([
    'NODE_ENV',
    'GTM_ENV',
    'GTM_PREVIEW',
    'GIT_SHA',
    'UNION_PLATFORM_ENV',
    'UNION_PLATFORM_REGION',
    'UNION_ORG_OVERRIDE',
  ])
}

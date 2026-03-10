import { Button } from '@/components/Button'
import { DOCS_BYOC_USER_GUIDE_URL } from '@/lib/constants'

export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-black">
      <div className="flex flex-col items-center text-center">
        <p className="mb-2 text-[18px] leading-[28px] font-medium text-white">
          404
        </p>
        <h1 className="mb-3 text-center text-[48px] leading-[48px] font-bold tracking-[-0.025em] text-white">
          Page not found
        </h1>
        <p className="mb-8 text-lg">Sorry, this page cannot be found</p>
        <div className="flex gap-4">
          <Button
            className="px-6 py-2 text-base font-semibold"
            outline
            size="lg"
            href={DOCS_BYOC_USER_GUIDE_URL}
          >
            Read the docs
          </Button>
          <Button
            className="px-6 py-2 text-base font-semibold"
            plain
            size="lg"
            href="mailto:support@union.ai"
          >
            Contact support
          </Button>
        </div>
      </div>
    </div>
  )
}

import { Button } from '@/components/Button'
import { useLayoutStore } from './state/LayoutStore'
import { PaneOpenLeft } from '@/components/icons/PaneOpenLeft'
import { PaneCloseLeft } from '@/components/icons/PaneCloseLeft'

export const LayoutToggle = () => {
  const mode = useLayoutStore((s) => s.mode)
  const setMode = useLayoutStore((s) => s.setMode)
  return (
    <span className="absolute top-0 left-4 z-50">
      {mode === 'default' ? (
        <Button
          aria-label="Hide action log"
          className="!px-1 !py-0"
          plain
          onClick={() => setMode('no-action-log')}
        >
          <PaneCloseLeft className="text-(--system-gray-1) dark:text-(--system-gray-6)" />
        </Button>
      ) : (
        <Button
          aria-label="Show action log"
          className="!px-1 !py-0"
          plain
          onClick={() => setMode('default')}
        >
          <PaneOpenLeft className="size-4 text-(--system-gray-1) dark:text-(--system-gray-6)" />
        </Button>
      )}
    </span>
  )
}

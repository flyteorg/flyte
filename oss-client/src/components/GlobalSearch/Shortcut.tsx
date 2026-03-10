export const Shortcut = ({ children }: { children: React.ReactNode }) => {
  return (
    <span className="flex h-6 items-center rounded-sm border-1 border-(--system-gray-3) px-1 text-sm text-(--system-gray-6)">
      {children}
    </span>
  )
}

export const OpenSearchShortCut = ({ className }: { className?: string }) => {
  const isMac =
    typeof navigator !== 'undefined' &&
    typeof navigator.platform === 'string' &&
    navigator.platform.toUpperCase().includes('MAC')
  return (
    <div
      className={`absolute right-3 flex items-center gap-1 ${className ?? 'top-3'}`}
    >
      {isMac ? (
        <>
          <Shortcut>⌘</Shortcut>
          <Shortcut>K</Shortcut>
        </>
      ) : (
        <>
          <Shortcut>Ctrl</Shortcut>
          <Shortcut>K</Shortcut>
        </>
      )}
    </div>
  )
}

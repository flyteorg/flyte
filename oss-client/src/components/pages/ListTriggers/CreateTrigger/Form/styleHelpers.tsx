import clsx from 'clsx'

export const HelpText = ({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) => (
  <div
    className={clsx('text-2xs font-medium text-(--system-gray-5)', className)}
  >
    {children}
  </div>
)

import clsx from 'clsx'

export function DescriptionList({
  className,
  ...props
}: React.ComponentPropsWithoutRef<'dl'>) {
  return (
    <dl
      {...props}
      className={clsx(
        className,
        'grid grid-cols-1 text-base/6 sm:grid-cols-2 sm:text-sm/6',
      )}
    />
  )
}

export function DescriptionTerm({
  className,
  ...props
}: React.ComponentPropsWithoutRef<'dt'>) {
  return (
    <dt
      {...props}
      className={clsx(
        className,
        'col-start-1 pt-3 text-zinc-500 sm:py-2 dark:text-zinc-400',
      )}
    />
  )
}

export function DescriptionDetails({
  className,
  ...props
}: React.ComponentPropsWithoutRef<'dd'>) {
  return (
    <dd
      {...props}
      className={clsx(
        className,
        'pt-1 pb-3 text-zinc-950 sm:py-2 dark:text-white',
      )}
    />
  )
}

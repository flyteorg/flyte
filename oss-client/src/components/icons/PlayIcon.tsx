export function PlayIcon(
  props: React.ComponentPropsWithoutRef<'svg'>,
) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="12"
      height="12"
      viewBox="0 0 12 12"
      fill="none"
      {...props}
    >
      <path d="M4 9.5V2.5L9.5 6L4 9.5Z" fill={props.fill ?? '#919191'} />
    </svg>
  )
}

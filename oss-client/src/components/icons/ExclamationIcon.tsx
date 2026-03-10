export const ExclamationIcon = (
  props: React.ComponentPropsWithoutRef<'svg'>,
) => (
  <svg
    width="2"
    height="6"
    viewBox="0 0 2 6"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    {...props}
  >
    <path
      d="M1 1L1 4"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
    />
    <circle cx="1" cy="5" r="1" fill="currentColor" />
  </svg>
)

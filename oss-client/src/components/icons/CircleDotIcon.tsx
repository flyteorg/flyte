export const CircleDotIcon = (props: React.ComponentPropsWithoutRef<'svg'>) => (
  <svg
    width="10"
    height="10"
    viewBox="0 0 10 10"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    {...props}
  >
    <circle cx="5" cy="5" r="4" stroke="currentColor" strokeWidth="1.5" />
    <circle cx="5" cy="5" r="2" fill="currentColor" />
  </svg>
)

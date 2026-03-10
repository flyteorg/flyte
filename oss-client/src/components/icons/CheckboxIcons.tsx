export const CheckedBoxIcon = (
  props: React.ComponentPropsWithoutRef<'svg'>,
) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="17"
    height="17"
    viewBox="0 0 17 17"
    fill="none"
    {...props}
  >
    <rect
      x="1.0625"
      y="0.734375"
      width="15"
      height="15"
      rx="3.5"
      stroke="#434343"
    />
    <path
      d="M5.5625 8.53477L7.3625 10.3348L11.5625 6.13477"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

export const UncheckedBoxIcon = (
  props: React.ComponentPropsWithoutRef<'svg'>,
) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="17"
    viewBox="0 0 16 17"
    fill="none"
    {...props}
  >
    <rect
      x="0.5"
      y="0.90625"
      width="15"
      height="15"
      rx="3.5"
      stroke="#434343"
    />
  </svg>
)

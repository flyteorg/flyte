export default function CacheWrittenIcon(
  props: React.ComponentPropsWithoutRef<'svg'>,
) {
  return (
    <svg viewBox="0 0 17 16" aria-hidden="true" {...props}>
      <defs>
        <mask id="cut-plus">
          <rect width="100%" height="100%" fill="white" />
          <path
            d="M8.40039 5.52734V10.4716"
            stroke="black"
            strokeWidth="1.5"
            strokeLinecap="round"
          />
          <path
            d="M5.92773 8L10.872 8"
            stroke="black"
            strokeWidth="1.5"
            strokeLinecap="round"
          />
        </mask>
      </defs>
      <path
        d="M8.40039 15C12.3768 15 14.4004 13.6942 14.4004 12.0833V3.91667C14.4004 2.30667 12.3724 1 8.40039 1C4.42839 1 2.40039 2.30667 2.40039 3.91667V12.0833C2.40039 13.6942 4.42394 15 8.40039 15Z"
        fill="currentColor"
        mask="url(#cut-plus)"
      />
    </svg>
  )
}

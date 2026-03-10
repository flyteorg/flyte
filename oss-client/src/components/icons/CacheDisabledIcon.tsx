export default function CacheDisabledIcon(
  props: React.ComponentPropsWithoutRef<'svg'>,
) {
  return (
    <svg viewBox="0 0 17 16" aria-hidden="true" {...props}>
      <defs>
        <mask id="cut-line">
          <rect width="100%" height="100%" fill="white" />
          <path
            d="M13.6523 0.620117L1.15064 13.1218"
            stroke="black"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </mask>
      </defs>
      <path
        d="M8.40039 15C12.3768 15 14.4004 13.6942 14.4004 12.0833V3.91667C14.4004 2.30667 12.3724 1 8.40039 1C4.42839 1 2.40039 2.30667 2.40039 3.91667V12.0833C2.40039 13.6942 4.42394 15 8.40039 15Z"
        fill="currentColor"
        mask="url(#cut-line)"
      />
      <path
        d="M14.6523 1.74805L2.15064 14.2498"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

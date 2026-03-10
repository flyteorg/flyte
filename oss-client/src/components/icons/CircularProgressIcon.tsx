export function CircularProgressIcon({
  isStatic,
  ...props
}: React.ComponentPropsWithoutRef<'svg'> & { isStatic: boolean }) {
  return (
    <div className="flex items-center justify-center">
      <div className="relative">
        <svg
          width="40"
          height="40"
          viewBox="0 0 40 40"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          {...props}
        >
          <circle
            cx="20"
            cy="20"
            r="14"
            stroke="currentColor"
            strokeWidth="3"
            strokeLinecap="round"
            // transform the arc so the gap is on top-left quarter
            strokeDasharray={isStatic ? '66 88' : '44 88'}
            strokeDashoffset={isStatic ? '-22' : '0'}
            transform={isStatic ? 'rotate(-180 20 20)' : undefined}
            className={isStatic ? '' : 'animate-spin-constant'}
          />
        </svg>

        <style jsx>{`
          @keyframes spin-constant {
            0% {
              transform: rotate(0deg);
            }
            100% {
              transform: rotate(360deg);
            }
          }

          .animate-spin-constant {
            animation: spin-constant 1s linear infinite;
            transform-origin: center;
          }
        `}</style>
      </div>
    </div>
  )
}

export const Dot = ({ color }: { color: string }) => (
  <div
    className="h-1.5 w-1.5"
    style={{
      background: color,
      borderRadius: '100%',
    }}
  />
)

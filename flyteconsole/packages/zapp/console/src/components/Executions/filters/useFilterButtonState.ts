import { useState } from 'react';

/** Basic common state to be used with a `FilterPopoverButton` */
export function useFilterButtonState() {
  const [open, setOpen] = useState(false);
  return {
    open,
    setOpen,
    onClick: () => setOpen(!open),
  };
}

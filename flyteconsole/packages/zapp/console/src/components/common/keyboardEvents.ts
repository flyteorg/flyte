/** Creates an event handler that will call the specified callback when
 * the event is the result of pressing the Escape key.
 */
export function escapeKeyListener(callback: () => void) {
  return function (event: React.KeyboardEvent<HTMLTextAreaElement | HTMLInputElement>) {
    if (event.defaultPrevented) {
      return;
    }

    const key = event.key || event.keyCode;

    if (key === 'Escape' || key === 'Esc' || key === 27) {
      callback();
    }
  };
}

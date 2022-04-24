// Canvas will be created on first use and cached for better performance
let textMeasureCanvas: HTMLCanvasElement;

/** Uses a canvas to measure the needed width to render a string using a given
 * font definition.
 */
export function measureText(fontDefinition: string, text: string) {
  if (!textMeasureCanvas) {
    textMeasureCanvas = document.createElement('canvas');
  }
  const context = textMeasureCanvas.getContext('2d');
  if (!context) {
    throw new Error('Unable to create canvas context for text measurement');
  }

  context.font = fontDefinition;
  return context.measureText(text);
}

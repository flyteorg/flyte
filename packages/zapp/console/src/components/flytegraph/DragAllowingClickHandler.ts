import * as React from 'react';

export type DragFilteringClickHandlerListener = (event: React.MouseEvent) => void;

/** A helper class to filter click events on an element if the mousedown event
 * results in a drag further than the specified threshold. The provided listener
 * will be called with the final mouseup event if the click has not been filtered
 */
export class DragFilteringClickHandler {
  private deltaX = 0;
  private deltaY = 0;
  private xPos = 0;
  private yPos = 0;
  private dragging = false;

  constructor(
    private listener: DragFilteringClickHandlerListener,
    private dragThresholdPx: number = 2,
  ) {}

  public onMouseDown = (event: React.MouseEvent) => {
    this.dragging = true;
    this.xPos = event.clientX;
    this.yPos = event.clientY;
    this.deltaX = 0;
    this.deltaY = 0;
  };

  public onMouseUp = (event: React.MouseEvent) => {
    if (!this.dragging) {
      return;
    }

    this.dragging = false;
    if (this.deltaX < this.dragThresholdPx && this.deltaY < this.dragThresholdPx) {
      this.listener(event);
    }
  };

  public onMouseMove = (event: React.MouseEvent) => {
    if (!this.dragging) {
      return;
    }
    this.deltaX += Math.abs(event.clientX - this.xPos);
    this.deltaY += Math.abs(event.clientY - this.yPos);
    this.xPos = event.clientX;
    this.yPos = event.clientY;
  };
}

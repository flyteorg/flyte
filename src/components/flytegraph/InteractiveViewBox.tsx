import * as React from 'react';

const viewBoxStyles = {
  display: 'flex',
  height: '100%',
  width: '100%',
};

interface InteractiveViewBoxChildrenProps {
  /** An SVG viewbox string, suitable for use in a `<svg />` element */
  viewBox: string;
}

interface InteractiveViewBoxProps {
  /** A single function component which will accept the current viewBox
   * string and render arbitrary content
   */
  children: (props: InteractiveViewBoxChildrenProps) => React.ReactNode;
  /** The natural width of the content */
  height: number;
  /** The natural height of the content */
  width: number;
}

interface ViewBoxRect {
  naturalHeight: number;
  naturalWidth: number;
  height: number;
  width: number;
  x: number;
  y: number;
}

interface DragData {
  xPos: number;
  yPos: number;
}

/** The smallest ratio difference between actual and natural dimensions */
const minScale = 0.1;
/** The percentage increment/decrement mapped to each pixel on the mouse wheel */
const scaleMultiplier = 0.001;
/** The ratio of screen pixels to viewport pixels used when dragging. Values
 * greater than 1 mean a faster / more sensitive drag.
 */
const dragMultiplier = 2;
const defaultDragData: DragData = { xPos: 0, yPos: 0 };

/** Translates a ViewBoxRect by a given x/y, clamping it within the original
 * (natural) dimensions
 */
function translateViewBox(
  viewBox: ViewBoxRect,
  translateX: number,
  translateY: number,
): ViewBoxRect {
  const x = Math.min(Math.max(0, viewBox.x - translateX), viewBox.naturalWidth - viewBox.width);
  const y = Math.min(Math.max(0, viewBox.y - translateY), viewBox.naturalHeight - viewBox.height);
  return { ...viewBox, x, y };
}

/** Scales a ViewBoxRect, clamping it within the original dimensions (when
 * zooming out) and the smallest allowed viewBox (when zooming in) */
function scaleViewBox(viewBox: ViewBoxRect, scaleDelta: number): ViewBoxRect {
  const width =
    scaleDelta < 1
      ? Math.max(minScale * viewBox.naturalWidth, viewBox.width * scaleDelta)
      : Math.min(viewBox.naturalWidth, viewBox.width * scaleDelta);
  const height =
    scaleDelta < 1
      ? Math.max(minScale * viewBox.naturalHeight, viewBox.height * scaleDelta)
      : Math.min(viewBox.naturalHeight, viewBox.height * scaleDelta);
  return { ...viewBox, height, width };
}

class InteractiveViewBoxImpl extends React.Component<InteractiveViewBoxProps> {
  private viewBox: ViewBoxRect;
  private updating = false;
  private dragData: DragData = defaultDragData;
  private viewboxRef: React.RefObject<HTMLDivElement>;

  constructor(props: InteractiveViewBoxProps) {
    super(props);
    this.viewBox = {
      naturalWidth: props.width,
      naturalHeight: props.height,
      width: props.width,
      height: props.height,
      x: 0,
      y: 0,
    };
    this.viewboxRef = React.createRef();
  }

  public componentDidMount() {
    if (!this.viewboxRef.current) {
      return;
    }

    this.viewboxRef.current.addEventListener('wheel', this.onWheel, {
      capture: true,
    });
  }

  public componentWillUnmount() {
    this.detachDrag();
    if (this.viewboxRef.current) {
      this.viewboxRef.current.removeEventListener('wheel', this.onWheel, {
        capture: true,
      });
    }
  }

  private onWheel = (event: WheelEvent) => {
    event.preventDefault();
    // Make the zooming less twitchy
    if (Math.abs(event.deltaY) < 2) {
      return;
    }

    const scaleDelta = 1.0 + scaleMultiplier * event.deltaY;
    const scaled = scaleViewBox(this.viewBox, scaleDelta);

    // To keep the zoom centered, we need to translate by half the change
    // in width/height
    const xDiff = (this.viewBox.width - scaled.width) / 2;
    const yDiff = (this.viewBox.height - scaled.height) / 2;
    const translated = translateViewBox(scaled, -xDiff, -yDiff);

    this.updateViewBox(translated);
  };

  private onMouseDown = (event: React.MouseEvent) => {
    this.attachDrag(event);
  };

  private onMouseUp = () => {
    this.detachDrag();
  };

  private attachDrag = (event: React.MouseEvent) => {
    const { clientX: xStart, clientY: yStart } = event;
    this.dragData = {
      xPos: xStart,
      yPos: yStart,
    };

    document.addEventListener('mousemove', this.onMouseMove, {
      capture: true,
    });
    document.addEventListener('mouseup', this.onMouseUp, { capture: true });
  };

  private detachDrag = () => {
    document.removeEventListener('mousemove', this.onMouseMove, {
      capture: true,
    });
    document.removeEventListener('mouseup', this.onMouseUp, {
      capture: true,
    });

    this.dragData = defaultDragData;
  };

  private onMouseMove = (event: MouseEvent) => {
    const deltaX = (event.clientX - this.dragData.xPos) * dragMultiplier;
    const deltaY = (event.clientY - this.dragData.yPos) * dragMultiplier;

    this.dragData = { xPos: event.clientX, yPos: event.clientY };
    this.updateViewBox(translateViewBox(this.viewBox, deltaX, deltaY));
  };

  private updateViewBox(viewBox: ViewBoxRect) {
    this.viewBox = viewBox;
    if (!this.updating) {
      this.updating = true;
      window.requestAnimationFrame(() => {
        this.forceUpdate();
        this.updating = false;
      });
    }
  }

  public render() {
    const { x, y, width, height } = this.viewBox;
    const viewBox = `${x},${y},${width},${height}`;
    return (
      <div onMouseDownCapture={this.onMouseDown} style={viewBoxStyles} ref={this.viewboxRef}>
        {this.props.children({ viewBox })}
      </div>
    );
  }
}

/** A Wrapper component for SVGs which provides zoom / pan mouse interactions
 * and feeds the resulting `viewBox` string to its child component. This
 * component will fill stretch to fill the available *content* space of its
 * parent. So it's best used inside a flex child with `flex-grow: 1`, or an
 * element with an explicit width/height.
 */
export const InteractiveViewBox: React.FC<InteractiveViewBoxProps> = (props) => (
  // Using key to force a re-mount of the viewbox if the desired width/height
  // changed.
  <InteractiveViewBoxImpl {...props} key={`${props.width}_${props.height}`} />
);

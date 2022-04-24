import * as React from 'react';

/** This is the definition for an arrowhead-style point marker. It should be
 * inserted in the `<defs>` section of the parent `<svg>` component in order to
 * be available to any path/line components rendered as children. The `id`
 * property is configurable and should match the value used in `url(#)`-style
 * strings (e.g the `markerMid` property in NodeLinks)
 */
export const Arrowhead: React.FC<{ id: string; fill: string }> = ({ fill, id }) => {
  return (
    <marker
      id={id}
      viewBox="0 0 10 15"
      refX="5"
      refY="7.5"
      markerWidth="7"
      markerHeight="10.5"
      orient="auto"
    >
      <path d="M 0 0 L 10 7.5 L 0 15 z" fill={fill} />
    </marker>
  );
};

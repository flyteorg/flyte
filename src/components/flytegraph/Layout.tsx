import { isEqual } from 'lodash';
import memoizeOne, { EqualityFn } from 'memoize-one';
import * as React from 'react';
import * as shallowEqual from 'shallowequal';

import { layoutGraph } from './layoutUtils';
import { LayoutProps } from './types';

type LayoutGraphArgs = Parameters<typeof layoutGraph>;
function layoutArgsAreEqual(newArgs: LayoutGraphArgs, oldArgs: LayoutGraphArgs) {
  const [newData, newConfig] = newArgs;
  const [oldData, oldConfig] = oldArgs;
  /* Nodes can be very deep structures, so we don't want to do a deep compare.
   * For the purposes of layout, we just want to check that the array of
   * references to nodes hasn't changed. Changing any value in the config will
   * result in re-computing a layout, so we do a deep comparison there.
   */
  return shallowEqual(newData, oldData) && isEqual(newConfig, oldConfig);
}

/** Performs layout of graph nodes and passes them to a child component for
 * rendering
 */
export class Layout<T> extends React.Component<LayoutProps<T>> {
  // We memoize the layout result to avoid recomputing it unless the
  // input props actually change
  layoutGraph = memoizeOne<typeof layoutGraph>(
    (data, config) => layoutGraph(data, config),
    layoutArgsAreEqual as EqualityFn,
  );

  public render() {
    const { config, data } = this.props;
    return this.props.children(this.layoutGraph(data, config));
  }
}

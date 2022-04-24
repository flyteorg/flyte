import { measureText } from 'components/flytegraph/layoutUtils';
import { Node } from 'components/flytegraph/Node';
import { NodeRendererProps } from 'components/flytegraph/types';
import { taskColors } from 'components/Theme/constants';
import { DAGNode } from 'models/Graph/types';
import { TaskType } from 'models/Task/constants';
import * as React from 'react';

const textWidths: Dictionary<number> = {};
function getTextWidthForLabel(label: string, fontSize: number) {
  const key = `${fontSize}:${label}`;
  if (textWidths[key] != null) {
    return textWidths[key];
  }
  const computed = measureText(fontSize, label).width;
  textWidths[key] = computed;
  return computed;
}

interface InputOutputNodeRendererProps extends NodeRendererProps<DAGNode> {
  label: string;
}

/** Special case renderer for the start/end nodes in a graph */
export const InputOutputNodeRenderer: React.FC<InputOutputNodeRendererProps> = (props) => {
  const { node, config, label } = props;
  const fillColor = taskColors[TaskType.UNKNOWN];
  const textWidth = getTextWidthForLabel(label, config.fontSize);

  return (
    <Node
      {...props}
      config={{ ...config, fillColor, cornerRounding: 0 }}
      node={{ ...node, textWidth, id: label }}
    />
  );
};

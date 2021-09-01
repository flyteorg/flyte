import { Typography } from '@material-ui/core';
import * as classnames from 'classnames';
import { getCacheKey } from 'components/Cache/utils';
import { useTheme } from 'components/Theme/useTheme';
import { Admin } from 'flyteidl';
import * as React from 'react';
import { NodeExecutionGroup } from '../types';
import { NodeExecutionRow } from './NodeExecutionRow';
import { useExecutionTableStyles } from './styles';
import { calculateNodeExecutionRowLeftSpacing } from './utils';

export interface NodeExecutionChildrenProps {
    abortMetadata?: Admin.IAbortMetadata;
    childGroups: NodeExecutionGroup[];
    level: number;
}

/** Renders a nested list of row items for children of a NodeExecution */
export const NodeExecutionChildren: React.FC<NodeExecutionChildrenProps> = ({
    abortMetadata,
    childGroups,
    level
}) => {
    const showNames = childGroups.length > 1;
    const tableStyles = useExecutionTableStyles();
    const theme = useTheme();
    const childGroupLabelStyle = {
        // The label is aligned with the parent above, so remove one level of spacing
        marginLeft: `${calculateNodeExecutionRowLeftSpacing(
            level - 1,
            theme.spacing
        )}px`
    };
    return (
        <>
            {childGroups.map(({ name, nodeExecutions }, groupIndex) => {
                const rows = nodeExecutions.map((nodeExecution, index) => (
                    <NodeExecutionRow
                        abortMetadata={abortMetadata}
                        key={getCacheKey(nodeExecution.id)}
                        index={index}
                        execution={nodeExecution}
                        level={level}
                    />
                ));
                const key = `group-${name}`;
                return showNames ? (
                    <div key={key} role="list">
                        <div
                            className={classnames(
                                { [tableStyles.borderTop]: groupIndex > 0 },
                                tableStyles.borderBottom,
                                tableStyles.childGroupLabel
                            )}
                            title={name}
                            style={childGroupLabelStyle}
                        >
                            <Typography
                                variant="overline"
                                color="textSecondary"
                            >
                                {name}
                            </Typography>
                        </div>
                        <div>{rows}</div>
                    </div>
                ) : (
                    <div key={key} role="list">
                        {rows}
                    </div>
                );
            })}
        </>
    );
};

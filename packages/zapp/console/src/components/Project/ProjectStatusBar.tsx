import * as React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import classNames from 'classnames';
import { Link } from 'react-router-dom';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import { barChartColors } from 'components/common/constants';
import { useCommonStyles } from 'components/common/styles';

const useStyles = makeStyles(() => ({
  barContainer: {
    display: 'block',
  },
  barItem: {
    display: 'inline-block',
    marginRight: 2,
    borderRadius: 2,
    width: 10,
    height: 12,
    backgroundColor: barChartColors.default,
  },
  successBarItem: {
    backgroundColor: barChartColors.success,
  },
  failedBarItem: {
    backgroundColor: barChartColors.failure,
  },
}));

interface ProjectStatusBarProps {
  items: WorkflowExecutionPhase[];
  paths: string[];
}

/**
 * Renders status of executions
 * @param items
 * @constructor
 */
const ProjectStatusBar: React.FC<ProjectStatusBarProps> = ({ items, paths }) => {
  const styles = useStyles();
  const commonStyles = useCommonStyles();

  return (
    <div className={styles.barContainer}>
      {items.map((item, idx) => {
        return (
          <Link
            className={commonStyles.linkUnstyled}
            to={paths[idx] ?? '#'}
            key={`bar-item-${idx}`}
          >
            <div
              className={classNames(styles.barItem, {
                [styles.successBarItem]: item === WorkflowExecutionPhase.SUCCEEDED,
                [styles.failedBarItem]: item >= WorkflowExecutionPhase.FAILED,
              })}
            />
          </Link>
        );
      })}
    </div>
  );
};

export default ProjectStatusBar;

import { makeStyles, Theme, useTheme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
  rowsContainer: {
    margin: `${theme.spacing(1)}px 0`,
  },
  rowContainer: {
    display: 'flex',
    marginTop: theme.spacing(0.5),
  },
  groupLabel: {
    marginRight: theme.spacing(2),
  },
}));

export interface DetailsItem {
  name: string;
  content: string | JSX.Element;
}

interface DetailsGroupProps {
  className?: string;
  items: DetailsItem[];
  /** Width of the labels in *grid units* */
  labelWidthGridUnits?: number;
}

/** Renders a list of detail items in a consistent style. Each item is shown in its own row with a label
 * to the left and the provided content to the right.
 */
export const DetailsGroup: React.FC<DetailsGroupProps> = ({
  className,
  items,
  labelWidthGridUnits = 14,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const theme = useTheme<Theme>();
  const width = `${theme.spacing(labelWidthGridUnits)}px`;
  const style = {
    width,
    minWidth: width,
  };
  return (
    <section className={classnames(styles.rowsContainer, className)}>
      {items.map(({ name, content }) => (
        <div key={name} className={styles.rowContainer}>
          <div className={classnames(styles.groupLabel, commonStyles.textMuted)} style={style}>
            {name}
          </div>
          <div>{content}</div>
        </div>
      ))}
    </section>
  );
};

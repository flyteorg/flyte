import { makeStyles, Theme } from '@material-ui/core/styles';
import { Variant } from '@material-ui/core/styles/createTypography';
import { SvgIconProps } from '@material-ui/core/SvgIcon';
import Typography from '@material-ui/core/Typography';
import classnames from 'classnames';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    alignItems: 'center',
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
    justifyContent: 'center',
    textAlign: 'center',
    width: '100%',
  },
  section: {
    marginBottom: theme.spacing(2),
  },
}));

type SizeValue = 'small' | 'medium' | 'large';
const iconSizes: Record<SizeValue, number> = {
  small: 40,
  medium: 64,
  large: 80,
};

const titleVariants: Record<SizeValue, Variant> = {
  small: 'body1',
  medium: 'h5',
  large: 'h4',
};

const descriptionVariants: Record<SizeValue, Variant> = {
  small: 'body2',
  medium: 'body1',
  large: 'body1',
};

export interface NonIdealStateProps {
  /** Optional className to be applied to the root component */
  className?: string;
  /** Optional description to be shown as body text for this state */
  description?: React.ReactNode;
  /** A component which renders the icon for this state */
  icon?: React.ComponentType<SvgIconProps>;
  /** A size variant to use. Defaults to 'medium' */
  size?: SizeValue;
  /** The primary message for this state, shown as header text */
  title: React.ReactNode;
}

/** A standard container for displaying a non-ideal state, such as an error,
 * a missing record, a lack of data based on a filter, or a degraded state which
 * requires user interaction to improve (ex. asking the user to make a selection).
 */
export const NonIdealState: React.FC<NonIdealStateProps> = ({
  children,
  className,
  description,
  icon: IconComponent,
  size = 'medium',
  title,
}) => {
  const styles = useStyles();
  const icon = IconComponent ? (
    <div style={{ fontSize: iconSizes[size] }}>
      <IconComponent color="action" fontSize="inherit" />
    </div>
  ) : null;

  return (
    <div className={classnames(styles.container, className)}>
      {icon}
      <Typography className={styles.section} variant={titleVariants[size]}>
        {title}
      </Typography>
      {description && (
        <Typography className={styles.section} variant={descriptionVariants[size]}>
          {description}
        </Typography>
      )}
      {children}
    </div>
  );
};

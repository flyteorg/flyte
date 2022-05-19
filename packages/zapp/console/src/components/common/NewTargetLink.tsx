import Link, { LinkProps } from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import OpenInNew from '@material-ui/icons/OpenInNew';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';

const useStyles = makeStyles({
  externalBlockContainer: {
    alignItems: 'center',
    display: 'flex',
  },
});

interface NewTargetLinkProps extends LinkProps {
  /** If set to true, will show an icon next to the link hinting to the user that they will be leaving the site */
  external?: boolean;
  /** If set to true, will be rendered as a <span> to preserve inline behavior */
  inline?: boolean;
}

/** Renders a link which will be opened in a new tab while also including the required props to avoid
 * linter errors. Can be configured to show a special icon for external links as a hint to the user.
 */
export const NewTargetLink: React.FC<NewTargetLinkProps> = (props) => {
  const { className, children, external = false, inline = false, ...otherProps } = props;
  const commonStyles = useCommonStyles();
  const styles = useStyles();

  const icon = external ? <OpenInNew className={commonStyles.iconRight} /> : null;

  return (
    <Link {...otherProps} target="_blank" rel="noopener noreferrer">
      {inline ? (
        <span>
          {children}
          {icon}
        </span>
      ) : (
        <div className={classnames(styles.externalBlockContainer, className)}>
          {children}
          {icon}
        </div>
      )}
    </Link>
  );
};

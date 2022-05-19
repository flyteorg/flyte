import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { Identifier } from 'models/Common/types';
import { NewTargetLink } from 'components/common/NewTargetLink';
import { versionDetailsUrlGenerator } from 'components/Entities/generators';
import t from '../strings';

const useStyles = makeStyles((theme: Theme) => ({
  link: {
    marginBottom: theme.spacing(2),
  },
}));

interface TaskVersionDetailsLinkProps {
  id: Identifier;
}

export const TaskVersionDetailsLink: React.FC<TaskVersionDetailsLinkProps> = ({ id }) => {
  const styles = useStyles();
  return (
    <NewTargetLink className={styles.link} href={versionDetailsUrlGenerator(id)} external={true}>
      {t('details_task')}
    </NewTargetLink>
  );
};

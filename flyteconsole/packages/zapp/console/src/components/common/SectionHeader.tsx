import { makeStyles, Theme } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    marginBottom: theme.spacing(1),
  },
}));

export interface SectionHeaderProps {
  title: string;
  subtitle?: string | null;
}
export const SectionHeader: React.FC<SectionHeaderProps> = ({ title, subtitle }) => {
  const styles = useStyles();
  return (
    <header className={styles.container}>
      <Typography variant="h6">{title}</Typography>
      {!!subtitle && <Typography variant="body2">{subtitle}</Typography>}
    </header>
  );
};

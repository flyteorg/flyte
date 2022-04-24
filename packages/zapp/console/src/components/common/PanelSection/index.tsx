import * as React from 'react';
import { makeStyles } from '@material-ui/core/styles';

const useStyle = makeStyles((theme) => ({
  detailsPanelCard: {
    borderBottom: `1px solid ${theme.palette.divider}`,
  },
  detailsPanelCardContent: {
    padding: `${theme.spacing(2)}px ${theme.spacing(3)}px`,
  },
}));

interface PanelSectionProps {
  children: React.ReactNode;
}

export const PanelSection = (props: PanelSectionProps) => {
  const commonStyles = useStyle();
  return (
    <div className={commonStyles.detailsPanelCard}>
      <div className={commonStyles.detailsPanelCardContent}>{props.children}</div>
    </div>
  );
};

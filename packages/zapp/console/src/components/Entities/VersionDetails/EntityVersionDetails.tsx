import * as React from 'react';
import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { contentMarginGridUnits } from 'common/layout';
import { WaitForData } from 'components/common/WaitForData';
import { useTaskTemplate } from 'components/hooks/useTask';
import { ResourceIdentifier, Identifier } from 'models/Common/types';
import { DumpJSON } from 'components/common/DumpJSON';
import { Row } from '../Row';
import EnvVarsTable from './EnvVarsTable';
import t, { patternKey } from '../strings';
import { entityStrings } from '../constants';

const useStyles = makeStyles((theme: Theme) => ({
  header: {
    marginBottom: theme.spacing(1),
    marginLeft: theme.spacing(contentMarginGridUnits),
  },
  table: {
    marginLeft: theme.spacing(contentMarginGridUnits),
  },
  divider: {
    borderBottom: `1px solid ${theme.palette.divider}`,
    marginBottom: theme.spacing(1),
  },
}));

export interface EntityExecutionsProps {
  id: ResourceIdentifier;
}

/** The tab/page content for viewing a workflow's executions */
export const EntityVersionDetails: React.FC<EntityExecutionsProps> = ({ id }) => {
  const styles = useStyles();

  // NOTE: need to be generic for supporting other type like workflow, etc.
  const templateState = useTaskTemplate(id as Identifier);

  const template = templateState?.value?.closure?.compiledTask?.template;
  const envVars = template?.container?.env;
  const image = template?.container?.image;

  return (
    <>
      <Typography className={styles.header} variant="h3">
        {t(patternKey('details', entityStrings[id.resourceType]))}
      </Typography>
      <div className={styles.divider} />
      <WaitForData {...templateState}>
        <div className={styles.table}>
          {image && (
            <Row title={t('imageFieldName')}>
              <Typography>{image}</Typography>
            </Row>
          )}
          {envVars && (
            <Row title={t('envVarsFieldName')}>
              <EnvVarsTable rows={envVars} />
            </Row>
          )}
          {template && (
            <Row title={t('commandsFieldName')}>
              <DumpJSON value={template} />
            </Row>
          )}
        </div>
      </WaitForData>
    </>
  );
};

import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { WaitForData } from 'components/common/WaitForData';
import { useNamedEntity } from 'components/hooks/useNamedEntity';
import { NamedEntityMetadata, ResourceIdentifier, Variable } from 'models/Common/types';
import * as React from 'react';
import reactLoadingSkeleton from 'react-loading-skeleton';
import { ReactJsonViewWrapper } from 'components/common/ReactJsonView';
import { useEntityVersions } from 'components/hooks/Entity/useEntityVersions';
import { executionSortFields } from 'models/Execution/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { TaskClosure } from 'models/Task/types';
import { executionFilterGenerator } from './generators';
import { Row } from './Row';
import t, { patternKey } from './strings';
import { entityStrings, entitySections } from './constants';

const Skeleton = reactLoadingSkeleton;

const useStyles = makeStyles((theme: Theme) => ({
  header: {
    marginBottom: theme.spacing(1),
  },
  description: {
    marginTop: theme.spacing(1),
  },
  divider: {
    borderBottom: `1px solid ${theme.palette.divider}`,
    marginBottom: theme.spacing(1),
  },
}));

const InputsAndOuputs: React.FC<{
  id: ResourceIdentifier;
}> = ({ id }) => {
  const sort = {
    key: executionSortFields.createdAt,
    direction: SortDirection.DESCENDING,
  };

  const baseFilters = executionFilterGenerator[id.resourceType](id);

  // to render the input and output,
  // need to fetch the latest version and get the input and ouptut data
  const versions = useEntityVersions(
    { ...id, version: '' },
    {
      sort,
      filter: baseFilters,
      limit: 1,
    },
  );

  let inputs: Record<string, Variable> | undefined;
  let outputs: Record<string, Variable> | undefined;

  if ((versions?.value?.[0]?.closure as TaskClosure)?.compiledTask?.template) {
    const template = (versions?.value?.[0]?.closure as TaskClosure)?.compiledTask?.template;
    inputs = template?.interface?.inputs?.variables;
    outputs = template?.interface?.outputs?.variables;
  }

  return (
    <WaitForData {...versions}>
      {inputs && (
        <Row title={t('inputsFieldName')}>
          <ReactJsonViewWrapper src={inputs} shouldCollapse={(field) => !field?.name} />
        </Row>
      )}
      {outputs && (
        <Row title={t('outputsFieldName')}>
          <ReactJsonViewWrapper src={outputs} shouldCollapse={(field) => !field?.name} />
        </Row>
      )}
    </WaitForData>
  );
};

/** Fetches and renders the description for a given Entity (LaunchPlan,Workflow,Task) ID */
export const EntityDescription: React.FC<{
  id: ResourceIdentifier;
}> = ({ id }) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const namedEntity = useNamedEntity(id);
  const { metadata = {} as NamedEntityMetadata } = namedEntity.value;
  const hasDescription = !!metadata.description;
  const sections = entitySections[id.resourceType];

  return (
    <>
      <Typography className={styles.header} variant="h3">
        {t('basicInformation')}
      </Typography>
      <div className={styles.divider} />
      <Typography variant="body2" component="span" className={styles.description}>
        <WaitForData {...namedEntity} spinnerVariant="none" loadingComponent={Skeleton}>
          <Row title={t('description')}>
            <span
              className={classnames({
                [commonStyles.hintText]: !hasDescription,
              })}
            >
              {hasDescription
                ? metadata.description
                : t(patternKey('noDescription', entityStrings[id.resourceType]))}
            </span>
          </Row>
        </WaitForData>
        {sections?.descriptionInputsAndOutputs && <InputsAndOuputs id={id} />}
      </Typography>
    </>
  );
};

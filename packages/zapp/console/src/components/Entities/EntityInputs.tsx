import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import CheckIcon from '@material-ui/icons/Check';
import { useLaunchPlans } from 'components/hooks/useLaunchPlans';
import { formatType, getInputDefintionForLiteralType } from 'components/Launch/LaunchForm/utils';
import { FilterOperationName } from 'models/AdminEntity/types';
import { ResourceIdentifier } from 'models/Common/types';
import { LaunchPlanClosure, LaunchPlanSpec } from 'models/Launch/types';
import * as React from 'react';
import t from './strings';
import { transformLiterals } from '../Literals/helpers';

const useStyles = makeStyles((theme: Theme) => ({
  header: {
    marginBottom: theme.spacing(1),
  },
  divider: {
    borderBottom: `1px solid ${theme.palette.divider}`,
    marginBottom: theme.spacing(1),
  },
  rowContainer: {
    display: 'flex',
    marginTop: theme.spacing(3),
  },
  firstColumnContainer: {
    width: '60%',
    marginRight: theme.spacing(3),
  },
  secondColumnContainer: {
    width: '40%',
  },
  configs: {
    listStyleType: 'none',
    paddingInlineStart: 0,
  },
  config: {
    display: 'flex',
  },
  configName: {
    color: theme.palette.grey[400],
    marginRight: theme.spacing(2),
    minWidth: '95px',
  },
  configValue: {
    color: '#333',
    fontSize: '14px',
  },
  headCell: {
    color: theme.palette.grey[400],
  },
  noInputs: {
    color: theme.palette.grey[400],
  },
}));

interface Input {
  name: string;
  type?: string;
  required?: boolean;
  defaultValue?: string;
}

/** Fetches and renders the expected & fixed inputs for a given Entity (LaunchPlan) ID */
export const EntityInputs: React.FC<{
  id: ResourceIdentifier;
}> = ({ id }) => {
  const styles = useStyles();

  const launchPlanState = useLaunchPlans(
    { project: id.project, domain: id.domain },
    {
      limit: 1,
      filter: [
        {
          key: 'launch_plan.name',
          operation: FilterOperationName.EQ,
          value: id.name,
        },
      ],
    },
  );

  const closure = launchPlanState?.value?.length
    ? launchPlanState.value[0].closure
    : ({} as LaunchPlanClosure);

  const spec = launchPlanState?.value?.length
    ? launchPlanState.value[0].spec
    : ({} as LaunchPlanSpec);

  const expectedInputs = React.useMemo<Input[]>(() => {
    const results = [] as Input[];
    Object.keys(closure?.expectedInputs?.parameters ?? {}).forEach((name) => {
      const parameter = closure?.expectedInputs.parameters[name];
      if (parameter?.var?.type) {
        const typeDefinition = getInputDefintionForLiteralType(parameter.var.type);
        results.push({
          name,
          type: formatType(typeDefinition),
          required: !!parameter.required,
          defaultValue: parameter.default?.value,
        });
      }
    });
    return results;
  }, [closure]);

  const fixedInputs = React.useMemo<Input[]>(() => {
    const inputsMap = transformLiterals(spec?.fixedInputs?.literals ?? {});
    return Object.keys(inputsMap).map((name) => ({ name, defaultValue: inputsMap[name] }));
  }, [spec]);

  const configs = React.useMemo(
    () => [
      { name: t('configType'), value: 'single (csv)' },
      {
        name: t('configUrl'),
        value:
          'https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv',
      },
      { name: t('configSeed'), value: '7' },
      { name: t('configTestSplitRatio'), value: '0.33' },
    ],
    [],
  );

  return (
    <>
      <Typography className={styles.header} variant="h3">
        {t('launchPlanLatest')}
      </Typography>
      <div className={styles.divider} />
      <div className={styles.rowContainer}>
        <div className={styles.firstColumnContainer}>
          <Typography className={styles.header} variant="h4">
            {t('expectedInputs')}
          </Typography>
          {expectedInputs.length ? (
            <TableContainer component={Paper}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>
                      <Typography className={styles.headCell} variant="h4">
                        {t('inputsName')}
                      </Typography>
                    </TableCell>
                    <TableCell className={styles.headCell}>
                      <Typography className={styles.headCell} variant="h4">
                        {t('inputsType')}
                      </Typography>
                    </TableCell>
                    <TableCell className={styles.headCell} align="center">
                      <Typography className={styles.headCell} variant="h4">
                        {t('inputsRequired')}
                      </Typography>
                    </TableCell>
                    <TableCell className={styles.headCell}>
                      <Typography className={styles.headCell} variant="h4">
                        {t('inputsDefault')}
                      </Typography>
                    </TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {expectedInputs.map(({ name, type, required, defaultValue }) => (
                    <TableRow key={name}>
                      <TableCell>{name}</TableCell>
                      <TableCell>{type}</TableCell>
                      <TableCell align="center">
                        {required ? <CheckIcon fontSize="small" /> : ''}
                      </TableCell>
                      <TableCell>{defaultValue || '-'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            <p className={styles.noInputs}>{t('noExpectedInputs')}</p>
          )}
        </div>
        <div className={styles.secondColumnContainer}>
          <Typography className={styles.header} variant="h4">
            {t('fixedInputs')}
          </Typography>
          {fixedInputs.length ? (
            <TableContainer component={Paper}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>
                      <Typography className={styles.headCell} variant="h4">
                        {t('inputsName')}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography className={styles.headCell} variant="h4">
                        {t('inputsDefault')}
                      </Typography>
                    </TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {fixedInputs.map(({ name, defaultValue }) => (
                    <TableRow key={name}>
                      <TableCell>{name}</TableCell>
                      <TableCell>{defaultValue || '-'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            <p className={styles.noInputs}>{t('noFixedInputs')}</p>
          )}
        </div>
      </div>
      {/* <div className={styles.rowContainer}>
        <div className={styles.firstColumnContainer}>
          <Typography className={styles.header} variant="h4">
            {t('configuration')}
          </Typography>
          <ul className={styles.configs}>
            {configs.map(({ name, value }) => (
              <li className={styles.config} key={name}>
                <span className={styles.configName}><Typography variant="body">{name}:</Typography></span>
                <span className={styles.configValue}><Typography variant="body">{value}</Typography></span>
              </li>
            ))}
          </ul>
        </div>
        <div className={styles.secondColumnContainer}></div>
      </div> */}
    </>
  );
};

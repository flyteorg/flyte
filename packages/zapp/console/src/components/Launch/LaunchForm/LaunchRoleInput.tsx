import { TextField, Typography } from '@material-ui/core';
import { NewTargetLink } from 'components/common/NewTargetLink';
import * as React from 'react';
import { AuthRoleStrings } from './constants';
import { makeStringChangeHandler } from './handlers';
import { useStyles } from './styles';
import { LaunchRoleInputRef, LaunchRoles, AuthRoleTypes } from './types';

const docLink =
  'https://github.com/flyteorg/flyteidl/blob/f4809c1a52073ec80de41be109bccc6331cdbac3/protos/flyteidl/core/security.proto#L111';

export interface LaunchRoleInputProps {
  initialValue?: any;
  showErrors: boolean;
}
const RoleDescription = () => (
  <>
    <Typography variant="body2">
      Enter values for
      <NewTargetLink inline={true} href={docLink}>
        &nbsp;Security Context&nbsp;
      </NewTargetLink>
      to assume for this execution.
    </Typography>
  </>
);

export const LaunchRoleInputImpl: React.RefForwardingComponent<
  LaunchRoleInputRef,
  LaunchRoleInputProps
> = (props, ref) => {
  const [state, setState] = React.useState({
    [AuthRoleTypes.IAM]: '',
    [AuthRoleTypes.k8]: '',
  });

  React.useEffect(() => {
    /* Note: if errors are present in other form elements don't reload new values */
    if (!props.showErrors) {
      if (props.initialValue?.securityContext) {
        setState((state) => ({
          ...state,
          [AuthRoleTypes.IAM]: props.initialValue.securityContext.runAs?.iamRole || '',
          [AuthRoleTypes.k8]: props.initialValue.securityContext.runAs?.k8sServiceAccount || '',
        }));
      } else if (props.initialValue) {
        setState((state) => ({
          ...state,
          [AuthRoleTypes.IAM]: props.initialValue?.assumableIamRole || '',
          [AuthRoleTypes.k8]: props.initialValue?.kubernetesServiceAccount || '',
        }));
      }
    }
  }, [props]);

  const getValue = () => {
    const authRole = {};
    const securityContext = { runAs: {} };

    if (state[AuthRoleTypes.k8].length > 0) {
      authRole['kubernetesServiceAccount'] = state[AuthRoleTypes.k8];
      securityContext['runAs']['k8sServiceAccount'] = state[AuthRoleTypes.k8];
    }

    if (state[AuthRoleTypes.IAM].length > 0) {
      authRole['assumableIamRole'] = state[AuthRoleTypes.IAM];
      securityContext['runAs']['iamRole'] = state[AuthRoleTypes.IAM];
    }
    return { authRole, securityContext } as LaunchRoles;
  };

  const handleInputChange = (value, type) => {
    setState((state) => ({
      ...state,
      [type]: value,
    }));
  };

  React.useImperativeHandle(ref, () => ({
    getValue,
    validate: () => true,
  }));

  const containerStyle: React.CSSProperties = {
    padding: '0 .5rem',
  };
  const inputContainerStyle: React.CSSProperties = {
    margin: '.5rem 0',
  };
  const styles = useStyles();

  return (
    <section>
      <header className={styles.sectionHeader}>
        <Typography variant="h6">Security Context</Typography>
        <RoleDescription />
      </header>
      <section style={containerStyle}>
        <div style={inputContainerStyle}>
          <Typography variant="body1">{AuthRoleStrings[AuthRoleTypes.IAM].label}</Typography>
        </div>
        <TextField
          id={`launchform-role-${AuthRoleTypes.IAM}`}
          helperText={AuthRoleStrings[AuthRoleTypes.IAM].helperText}
          fullWidth={true}
          label={AuthRoleStrings[AuthRoleTypes.IAM].inputLabel}
          value={state[AuthRoleTypes.IAM]}
          variant="outlined"
          onChange={makeStringChangeHandler(handleInputChange, AuthRoleTypes.IAM)}
        />
        <div style={inputContainerStyle}>
          <Typography variant="body1">{AuthRoleStrings[AuthRoleTypes.k8].label}</Typography>
        </div>
        <TextField
          id={`launchform-role-${AuthRoleTypes.k8}`}
          helperText={AuthRoleStrings[AuthRoleTypes.k8].helperText}
          fullWidth={true}
          label={AuthRoleStrings[AuthRoleTypes.k8].inputLabel}
          value={state[AuthRoleTypes.k8]}
          variant="outlined"
          onChange={makeStringChangeHandler(handleInputChange, AuthRoleTypes.k8)}
        />
      </section>
    </section>
  );
};

export const LaunchRoleInput = React.forwardRef(LaunchRoleInputImpl);

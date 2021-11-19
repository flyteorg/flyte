import {
    FormControl,
    FormControlLabel,
    FormLabel,
    Radio,
    RadioGroup,
    TextField,
    Typography
} from '@material-ui/core';
import { log } from 'common/log';
import { NewTargetLink } from 'components/common/NewTargetLink';
import { useDebouncedValue } from 'components/hooks/useDebouncedValue';
import { Admin } from 'flyteidl';
import * as React from 'react';
import { launchInputDebouncDelay, roleTypes } from './constants';
import { makeStringChangeHandler } from './handlers';
import { useInputValueCacheContext } from './inputValueCache';
import { useStyles } from './styles';
import { InputValueMap, LaunchRoleInputRef, RoleType } from './types';

const roleHeader = 'Role';
const roleDocLinkUrl =
    'https://github.com/flyteorg/flyteidl/blob/3789005a1372221eba28fa20d8386e44b32388f5/protos/flyteidl/admin/common.proto#L241';
const roleTypeLabel = 'type';
const roleInputId = 'launch-auth-role';
const defaultRoleTypeValue = roleTypes.iamRole;

export interface LaunchRoleInputProps {
    initialValue?: Admin.IAuthRole;
    showErrors: boolean;
}

interface LaunchRoleInputState {
    error?: string;
    roleType: RoleType;
    roleString?: string;
    getValue(): Admin.IAuthRole;
    onChangeRoleString(newValue: string): void;
    onChangeRoleType(newValue: string): void;
    validate(): boolean;
}

function getRoleTypeByValue(value: string): RoleType | undefined {
    return Object.values(roleTypes).find(
        ({ value: roleTypeValue }) => value === roleTypeValue
    );
}

const roleTypeCacheKey = '__roleType';
const roleStringCacheKey = '__roleString';

interface AuthRoleInitialValues {
    roleType: RoleType;
    roleString: string;
}

function getInitialValues(
    cache: InputValueMap,
    initialValue?: Admin.IAuthRole
): AuthRoleInitialValues {
    let roleType: RoleType | undefined;
    let roleString: string | undefined;

    // Prefer cached value first, since that is user input
    if (cache.has(roleTypeCacheKey)) {
        const cachedValue = `${cache.get(roleTypeCacheKey)}`;
        roleType = getRoleTypeByValue(cachedValue);
        if (roleType === undefined) {
            log.error(`Unexepected cached role type: ${cachedValue}`);
        }
    }
    if (cache.has(roleStringCacheKey)) {
        roleString = cache.get(roleStringCacheKey)?.toString();
    }

    // After trying cache, check for an initial value and populate either
    // field from the initial value if no cached value was passed.
    if (initialValue != null) {
        const initialRoleType = Object.values(roleTypes).find(
            rt => initialValue[rt.value]
        );
        if (initialRoleType != null && roleType == null) {
            roleType = initialRoleType;
        }
        if (initialRoleType != null && roleString == null) {
            roleString = initialValue[initialRoleType.value]?.toString();
        }
    }

    return {
        roleType: roleType ?? defaultRoleTypeValue,
        roleString: roleString ?? ''
    };
}

export function useRoleInputState(
    props: LaunchRoleInputProps
): LaunchRoleInputState {
    const inputValueCache = useInputValueCacheContext();
    const initialValues = getInitialValues(inputValueCache, props.initialValue);

    const [error, setError] = React.useState<string>();
    const [roleString, setRoleString] = React.useState<string>(
        initialValues.roleString
    );

    const [roleType, setRoleType] = React.useState<RoleType>(
        initialValues.roleType
    );

    const validationValue = useDebouncedValue(
        roleString,
        launchInputDebouncDelay
    );

    const getValue = () => ({ [roleType.value]: roleString });
    const validate = () => {
        if (roleString == null || roleString.length === 0) {
            setError('Value is required');
            return false;
        }
        setError(undefined);
        return true;
    };

    const onChangeRoleString = (value: string) => {
        inputValueCache.set(roleStringCacheKey, value);
        setRoleString(value);
    };

    const onChangeRoleType = (value: string) => {
        const newRoleType = getRoleTypeByValue(value);
        if (newRoleType === undefined) {
            throw new Error(`Unexpected role type value: ${value}`);
        }
        inputValueCache.set(roleTypeCacheKey, value);
        setRoleType(newRoleType);
    };

    React.useEffect(() => {
        validate();
    }, [validationValue]);

    return {
        error,
        getValue,
        onChangeRoleString,
        onChangeRoleType,
        roleType,
        roleString,
        validate
    };
}

const RoleDescription = () => (
    <>
        <Typography variant="body2">
            Enter a
            <NewTargetLink inline={true} href={roleDocLinkUrl}>
                &nbsp;role&nbsp;
            </NewTargetLink>
            to assume for this execution.
        </Typography>
    </>
);

export const LaunchRoleInputImpl: React.RefForwardingComponent<
    LaunchRoleInputRef,
    LaunchRoleInputProps
> = (props, ref) => {
    const styles = useStyles();
    const {
        error,
        getValue,
        roleType,
        roleString = '',
        onChangeRoleString,
        onChangeRoleType,
        validate
    } = useRoleInputState(props);
    const hasError = props.showErrors && !!error;
    const helperText = hasError ? error : roleType.helperText;

    React.useImperativeHandle(ref, () => ({
        getValue,
        validate
    }));

    return (
        <section>
            <header className={styles.sectionHeader}>
                <Typography variant="h6">{roleHeader}</Typography>
                <RoleDescription />
            </header>
            <FormControl component="fieldset">
                <FormLabel component="legend">{roleTypeLabel}</FormLabel>
                <RadioGroup
                    aria-label="roleType"
                    name="roleType"
                    value={roleType.value}
                    row={true}
                    onChange={makeStringChangeHandler(onChangeRoleType)}
                >
                    {Object.values(roleTypes).map(({ label, value }) => (
                        <FormControlLabel
                            key={value}
                            value={value}
                            control={<Radio />}
                            label={label}
                        />
                    ))}
                </RadioGroup>
            </FormControl>
            <div className={styles.formControl}>
                <TextField
                    error={hasError}
                    id={roleInputId}
                    helperText={helperText}
                    fullWidth={true}
                    label={roleType.inputLabel}
                    onChange={makeStringChangeHandler(onChangeRoleString)}
                    value={roleString}
                    variant="outlined"
                />
            </div>
        </section>
    );
};

/** Renders controls for selecting an AuthRole type and inputting a value for it. */
export const LaunchRoleInput = React.forwardRef(LaunchRoleInputImpl);

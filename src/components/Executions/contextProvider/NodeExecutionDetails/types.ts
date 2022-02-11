import { NodeExecutionDisplayType } from 'components/Executions/types';
import { Core } from 'flyteidl';

export const UNKNOWN_DETAILS = {
    displayId: 'unknownNode',
    displayType: NodeExecutionDisplayType.Unknown
};

export function isIdEqual(lhs: Core.IIdentifier, rhs: Core.IIdentifier) {
    return (
        lhs.resourceType === rhs.resourceType &&
        lhs.project === rhs.project &&
        lhs.domain === rhs.domain &&
        lhs.name === rhs.name &&
        lhs.version === rhs.version
    );
}

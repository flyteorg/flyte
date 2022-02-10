import classnames from 'classnames';
import { sortedObjectEntries } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import {
    ProtobufListValue,
    ProtobufStruct,
    ProtobufValue
} from 'models/Common/types';
import * as React from 'react';
import { htmlEntities } from '../constants';
import { PrintList } from '../PrintList';
import { PrintValue } from '../PrintValue';
import { useLiteralStyles } from '../styles';
import { ValueLabel } from '../ValueLabel';
import { NoneTypeValue } from './NoneTypeValue';

function renderListItem(value: any, index: number) {
    return (
        <RenderedProtobufValue label={index} value={value as ProtobufValue} />
    );
}

const RenderedProtobufValue: React.FC<{
    label: React.ReactNode;
    value: ProtobufValue;
}> = ({ label, value }) => {
    switch (value.kind) {
        case 'nullValue':
            return <PrintValue label={label} value={<NoneTypeValue />} />;
        case 'listValue': {
            const list = value.listValue as ProtobufListValue;
            return (
                <>
                    <ValueLabel label={label} />
                    <PrintList
                        values={list.values}
                        renderValue={renderListItem}
                    />
                </>
            );
        }
        case 'structValue':
            return (
                <>
                    <ValueLabel label={label} />
                    <ProtobufStructValue
                        struct={value.structValue as ProtobufStruct}
                    />
                </>
            );
        default:
            return <PrintValue label={label} value={`${value[value.kind]}`} />;
    }
};

/** Renders a `ProtobufStructValue` using appropriate sub-components for each
 * possible type */
export const ProtobufStructValue: React.FC<{
    struct: ProtobufStruct;
}> = ({ struct }) => {
    const commonStyles = useCommonStyles();
    const literalStyles = useLiteralStyles();
    const { fields } = struct;
    const mapContent = Object.keys(fields).length ? (
        <ul
            className={classnames(
                literalStyles.nestedContainer,
                commonStyles.textMonospace,
                commonStyles.listUnstyled
            )}
        >
            {sortedObjectEntries(fields).map(([key, value]) => (
                <li key={key}>
                    <RenderedProtobufValue label={key} value={value} />
                </li>
            ))}
        </ul>
    ) : null;
    return (
        <>
            <span>{htmlEntities.leftCurlyBrace}</span>
            {mapContent}
            <span>{htmlEntities.rightCurlyBrace}</span>
        </>
    );
};

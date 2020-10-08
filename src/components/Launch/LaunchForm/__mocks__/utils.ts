import { Core } from 'flyteidl';
import { BlobDimensionality } from 'models';

export function primitiveLiteral(primitive: Core.IPrimitive): Core.ILiteral {
    return { scalar: { primitive } };
}

export function blobLiteral({
    uri,
    format,
    dimensionality
}: {
    uri?: string;
    format?: string;
    dimensionality?: BlobDimensionality;
}): Core.ILiteral {
    return {
        scalar: {
            blob: { uri, metadata: { type: { format, dimensionality } } }
        }
    };
}

import { WaitForData } from 'components/common';
import { useRemoteLiteralMap } from 'components/hooks';
import { Literal, LiteralMap, UrlBlob } from 'models';
import * as React from 'react';
import { maxBlobDownloadSizeBytes } from './constants';
import { LiteralMapViewer } from './LiteralMapViewer';

const DownloadAndRenderBlob: React.FC<{ url: string }> = ({ url }) => {
    const data = useRemoteLiteralMap(url);
    return (
        <WaitForData {...data}>
            <LiteralMapViewer map={data.value} />
        </WaitForData>
    );
};

const BlobTooLarge: React.FC<{ url: string }> = ({ url }) => (
    <p>
        This data is too large to view. <a href={url}>Download</a>
    </p>
);

/** Given a UrlBlob which represents a LiteralMap stored at a particular
 * address, this component will initiate a fetch of the data and then render it
 * using a `LiteralMapViewer`. Special behaviors are used for blobs missing
 * a URL or which are too large to view in the UI. For the latter case, a direct
 * download link is provided. If `map` is defined, use it instead of fetching.
 */
export const RemoteLiteralMapViewer: React.FC<{
    blob: UrlBlob;
    map?: LiteralMap;
}> = ({ blob, map }) => {
    if (!blob.url || !blob.bytes) {
        return (
            <p>
                <em>No data is available.</em>
            </p>
        );
    }

    if (map !== undefined) {
        return <LiteralMapViewer map={map} />;
    }

    return blob.bytes.gt(maxBlobDownloadSizeBytes) ? (
        <BlobTooLarge url={blob.url} />
    ) : (
        <DownloadAndRenderBlob url={blob.url} />
    );
};

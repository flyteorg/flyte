import { useDownloadLink } from 'components/hooks/useDataProxy';
import { WaitForData } from 'components/common/WaitForData';
import * as React from 'react';
import { Core } from 'flyteidl';

/** Fetches and renders the deck data for a given `nodeExecutionId` */
export const ExecutionNodeDeck: React.FC<{ nodeExecutionId: Core.NodeExecutionIdentifier }> = ({
  nodeExecutionId,
}) => {
  const downloadLink = useDownloadLink(nodeExecutionId);

  return (
    <WaitForData {...downloadLink}>
      <iframe title="deck" height="100%" src={downloadLink?.value?.signedUrl?.[0]} />
    </WaitForData>
  );
};

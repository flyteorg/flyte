import { useDownloadLocation } from 'components/hooks/useDataProxy';
import { WaitForData } from 'components/common/WaitForData';
import * as React from 'react';

/** Fetches and renders the deck data for a given `deckUri` */
export const ExecutionNodeDeck: React.FC<{ deckUri: string }> = ({ deckUri }) => {
  const downloadLocation = useDownloadLocation(deckUri);

  return (
    <WaitForData {...downloadLocation}>
      <iframe title="deck" height="100%" src={downloadLocation.value.signedUrl} />
    </WaitForData>
  );
};

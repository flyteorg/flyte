import { useMemo } from 'react';
import { SearchableVersion } from './types';
import { versionsToSearchableSelectorOptions } from './utils';

export function useVersionSelectorOptions(versions: SearchableVersion[]) {
  return useMemo(() => {
    const options = versionsToSearchableSelectorOptions(versions);
    if (options.length > 0) {
      options[0].description = 'latest';
    }
    return options;
  }, [versions]);
}

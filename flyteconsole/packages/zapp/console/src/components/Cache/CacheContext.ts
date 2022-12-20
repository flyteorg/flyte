import * as React from 'react';
import { createCache } from './createCache';

/** Provides a caching layer for data-fetching hooks. A new cache "boundary"
 * can be created by inserting a `CacheContext.Provider` in the component tree
 */
export const CacheContext = React.createContext(createCache());

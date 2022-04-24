import { LocalCacheItem, useLocalCache } from 'basics/LocalCache';

export function useOnlyMineSelectedValue(settingKey: string): boolean {
  const [OnlyMineSettingValue] = useLocalCache(LocalCacheItem.OnlyMineSetting);
  const [onlyMineToggleValue] = useLocalCache(LocalCacheItem.OnlyMineToggle);
  const selectedValue = OnlyMineSettingValue[settingKey];

  // only apply the selected value, when the only mine toggle is on
  return onlyMineToggleValue && selectedValue;
}

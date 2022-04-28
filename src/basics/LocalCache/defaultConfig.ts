import { defaultSelectedValues } from './onlyMineDefaultConfig';

export enum LocalCacheItem {
  // Test flag is created only for unit-tests
  TestUndefined = 'test-undefined',
  TestSettingBool = 'test-setting-bool',
  TestObject = 'test-object',

  // Production flags
  ShowWorkflowVersions = 'flyte.show-workflow-versions',
  ShowDomainSettings = 'flyte.show-domain-settings',

  // Test Only Mine
  OnlyMineToggle = 'flyte.only-mine-toggle',
  OnlyMineSetting = 'flyte.only-mine-setting',
}

type LocalCacheConfig = { [k: string]: string };

/*
 * THe default value could be present as any simple type or as a valid JSON object
 * with all field names wrapped in double quotes
 */
export const defaultLocalCacheConfig: LocalCacheConfig = {
  // Test
  'test-setting-bool': 'false',
  'test-object': '{"name":"Stella","age":"125"}',

  // Production
  'flyte.show-workflow-versions': 'true',
  'flyte.show-domain-settings': 'true',

  // Test Only Mine
  'flyte.only-mine-toggle': 'true',
  'flyte.only-mine-setting': JSON.stringify(defaultSelectedValues),
};

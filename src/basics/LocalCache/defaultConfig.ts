export enum LocalCacheItem {
  // Test flag is created only for unit-tests
  TestUndefined = 'test-undefined',
  TestSettingBool = 'test-setting-bool',
  TestObject = 'test-object',

  // Production flags
  ShowWorkflowVersions = 'flyte.show-workflow-versions',
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

  // Feature flags - for prod testing
  'ff.timeline-view': 'false',
};

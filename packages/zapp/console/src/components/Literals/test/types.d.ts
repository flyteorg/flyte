export type TestCase<K> = {
  value: K;
  expected: any;
};

export type TestCaseList<K> = {
  [key: string]: TestCase<K>;
};

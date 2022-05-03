module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['../../../script/test/jest-setup.ts'],
  testPathIgnorePatterns: ['dist', 'node_modules', 'lib'],
  moduleFileExtensions: ['ts', 'tsx', 'js', 'json', 'node'],
  transform: {
    '^.+\\.(j|t)sx?$': 'ts-jest',
  },

  moduleNameMapper: {
    '@flyteconsole/(.*)': ['<rootDir>/packages/plugins/$1/src', '<rootDir>/packages/zapp/$1/src'],
  },

  coveragePathIgnorePatterns: ['mocks', 'src/index'],
};

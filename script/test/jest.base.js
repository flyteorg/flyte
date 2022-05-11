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
    // In this case <rootDir> will correspond to package/zapp/console (and so) locations
    '@flyteconsole/(.*)': [
      '<rootDir>/../../basics/$1/src',
      '<rootDir>/../../composites/$1/src',
      '<rootDir>/../../plugins/$1/src',
      '<rootDir>/../../zapp/$1/src',
    ],
  },

  coveragePathIgnorePatterns: ['mocks', '.mock', 'src/index'],
};

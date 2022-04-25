module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  testPathIgnorePatterns: ['__stories__', '.storybook', 'node_modules', 'dist', 'build'],
  setupFilesAfterEnv: ['<rootDir>/src/test/setupTests.ts'],
  clearMocks: true,
  verbose: false,

  moduleFileExtensions: ['ts', 'tsx', 'js', 'json', 'node'],
  modulePaths: ['<rootDir>/src'],
  roots: ['<rootDir>/src'],
  transformIgnorePatterns: ['<rootDir>/node_modules/(?!@flyteorg/flyteidl)', 'protobufjs/minimal'],
  transform: {
    '^.+\\.(j|t)sx?$': 'ts-jest',
  },
  moduleNameMapper: {
    '@flyteconsole/(.*)': [
      '<rootDir>/../../../packages/plugins/$1/src',
      '<rootDir>/../../../packages/zapp/$1/src',
    ],
    '\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$':
      '<rootDir>/assetsTransformer.js',
  },

  coverageDirectory: '../../../.coverage',
  collectCoverageFrom: ['**/*.{ts,tsx}', '!**/*/*.stories.{ts,tsx}', '!**/*/*.mocks.{ts,tsx}'],
  coveragePathIgnorePatterns: [
    '__stories__',
    '__mocks__',
    '<rootDir>/.storybook',
    '<rootDir>/node_modules',
    '<rootDir>/dist',
    '<rootDir>/build',
    '<rootDir>/src/tsd',
    '<rootDir>/src/server.ts',
    '<rootDir>/.eslintrc.js',
    '\\.config.js$',
  ],
  coverageReporters: ['text', 'json', 'html'],
};

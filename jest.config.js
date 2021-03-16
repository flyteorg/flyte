module.exports = {
    globals: {
        'ts-jest': {
            babelConfig: true
        }
    },
    verbose: false,
    moduleFileExtensions: ['ts', 'tsx', 'js', 'json', 'node'],
    modulePaths: ['<rootDir>/src'],
    roots: ['<rootDir>/src'],
    testPathIgnorePatterns: [
        '__stories__',
        '<rootDir>/.storybook',
        '<rootDir>/node_modules',
        '<rootDir>/dist',
        '<rootDir>/build'
    ],
    transform: {
        '^.+\\.(j|t)sx?$': 'ts-jest'
    },
    transformIgnorePatterns: ['<rootDir>/node_modules/(?!@flyteorg/flyteidl)'],
    coverageDirectory: '.coverage',
    collectCoverageFrom: ['**/*.{ts,tsx}'],
    coveragePathIgnorePatterns: [
        '__stories__',
        '<rootDir>/.storybook',
        '<rootDir>/node_modules',
        '<rootDir>/dist',
        '<rootDir>/build',
        '<rootDir>/src/tsd',
        '<rootDir>/.eslintrc.js',
        '\\.config.js$'
    ],
    coverageReporters: ['text', 'json'],
    clearMocks: true,
    setupFiles: ['<rootDir>/node_modules/regenerator-runtime/runtime'],
    setupFilesAfterEnv: ['<rootDir>/src/test/setupTests.ts'],
    preset: 'ts-jest'
};

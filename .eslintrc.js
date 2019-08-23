// We are using eslint with the typescript parser plugin only to validate things
// which are not supported in tslint, such as react hooks
module.exports = {
    parser: '@typescript-eslint/parser',
    parserOptions: {
        jsx: true,
        sourceType: 'module'
    },
    plugins: ['react-hooks'],
    rules: {
        'react-hooks/rules-of-hooks': 'error'
    }
};

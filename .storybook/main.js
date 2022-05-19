var path = require('path');
module.exports = {
  core: {
    builder: 'webpack5',
  },
  stories: ['../packages/**/*.stories.@(js|jsx|ts|tsx|mdx)'],
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    '@storybook/addon-interactions',
  ],
  framework: '@storybook/react',
  webpackFinal: async (config, { configType }) => {
    config.resolve.modules = ['packages/zapp/console/src', 'node_modules'];
    config.module.rules = [
      ...config.module.rules,
      {
        // needed to ensure proper same level of typing between app and storybook built.
        // without it we have troubles with exporting flyteidl types due to nested namespace nature of
        // flyteidl.d.ts file.
        test: /\.tsx?$/,
        exclude: /node_modules/,
        use: ['babel-loader', { loader: 'ts-loader', options: { transpileOnly: true } }],
      },
    ];

    config.resolve.alias = {
      ...config.resolve.alias,
      '@flyteconsole/locale': path.resolve(__dirname, '../packages/basics/locale/src'),
      '@flyteconsole/ui-atoms': path.resolve(__dirname, '../packages/composites/ui-atoms/src'),
      '@flyteconsole/components': path.resolve(__dirname, '../packages/plugins/components/src'),
      '@flyteconsole/flyte-api': path.resolve(__dirname, '../packages/plugins/flyte-api/src'),
    };

    return config;
  },
};

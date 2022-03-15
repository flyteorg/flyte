module.exports = {
  core: {
    builder: 'webpack5'
  },
  stories: ['../src/**/*.stories.mdx', '../src/**/*.stories.@(js|jsx|ts|tsx)'],
  addons: ['@storybook/addon-links', '@storybook/addon-essentials', '@storybook/addon-interactions'],
  framework: '@storybook/react',
  webpackFinal: async (config, { configType }) => {
    config.resolve.modules = ['src', 'node_modules'];
    config.module.rules = [
      ...config.module.rules,
      {
        // needed to ensure proper same level of typing between app and storybook built.
        // without it we have troubles with exporting flyteidl types due to nested namespace nature of
        // flyteidl.d.ts file.
        test: /\.tsx?$/,
        exclude: /node_modules/,
        use: ['babel-loader', { loader: 'ts-loader', options: { transpileOnly: true } }]
      }
    ];
    return config;
  }
};

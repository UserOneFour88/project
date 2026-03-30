import type { StorybookConfig } from '@storybook/react-webpack5';

const config: StorybookConfig = {
  stories: ['../src/**/*.mdx', '../src/**/*.stories.@(js|jsx|mjs|ts|tsx)'],
  addons: [
    '@storybook/addon-webpack5-compiler-swc',
    '@storybook/addon-a11y',
    '@storybook/addon-docs',
  ],
  framework: '@storybook/react-webpack5',
  webpackFinal: async (baseConfig) => {
    const rules = baseConfig.module?.rules ?? [];

    baseConfig.module = {
      ...baseConfig.module,
      rules: rules.map((rule) => {
        if (
          typeof rule === 'object' &&
          rule !== null &&
          'test' in rule &&
          rule.test instanceof RegExp &&
          rule.test.test('styles.css') &&
          'use' in rule &&
          Array.isArray(rule.use)
        ) {
          const alreadyHasPostCss = rule.use.some(
            (entry) =>
              (typeof entry === 'string' && entry.includes('postcss-loader')) ||
              (typeof entry === 'object' &&
                entry !== null &&
                'loader' in entry &&
                typeof entry.loader === 'string' &&
                entry.loader.includes('postcss-loader'))
          );

          if (!alreadyHasPostCss) {
            return {
              ...rule,
              use: [...rule.use, 'postcss-loader'],
            };
          }
        }

        return rule;
      }),
    };

    return baseConfig;
  },
};

export default config;
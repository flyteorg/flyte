import path from 'path';
import { fileURLToPath } from 'url';

// eslint-disable-next-line no-underscore-dangle
const __filename = fileURLToPath(import.meta.url);

// eslint-disable-next-line no-underscore-dangle
const __dirname = path.dirname(__filename);

export const projectTypeSoSettingsMap = {
  basics: {
    packagePartialPath: 'packages/basics/',
    targetDirectoryPartialPath: path.resolve(__dirname, '../../../packages/basics/'),
    // using the same template for basics, components and microapps for now
    templateDirectory: path.resolve(__dirname, '../templates', 'basics'),
  },
  composites: {
    packagePartialPath: 'packages/composites/',
    targetDirectoryPartialPath: path.resolve(__dirname, '../../../packages/composites/'),
    // using the same template for basics, components and microapps for now
    templateDirectory: path.resolve(__dirname, '../templates', 'basics'),
  },
  plugins: {
    packagePartialPath: 'packages/plugins/',
    targetDirectoryPartialPath: path.resolve(__dirname, '../../../packages/plugins/'),
    // using the same template for basics, components and microapps for now
    templateDirectory: path.resolve(__dirname, '../templates', 'basics'),
  },
};

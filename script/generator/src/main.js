import chalk from 'chalk';
import ncp from 'ncp';
import { promisify } from 'util';
import Listr from 'listr';
import { editPackageJSON } from './utils.js';

const copy = promisify(ncp);

async function copyTemplateFiles(options) {
  return copy(options.templateDirectory, options.targetDirectory, {
    clobber: false, // if set to false, ncp will not overwrite destination files that already exist
  });
}

async function createProject(config) {
  const tasks = new Listr([
    {
      title: 'Copy template files',
      task: async () => copyTemplateFiles(config),
    },
    {
      title: 'Update package.json',
      task: async () => editPackageJSON(config),
      enabled: () => true,
    },
  ]);

  await tasks.run();

  console.log(`${chalk.green.bold('DONE')}. Project ready`);

  return true;
}

export default createProject;

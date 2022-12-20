import path from 'path';
import chalk from 'chalk';
import inquirer from 'inquirer';
import { checkPathExists } from './utils.js';
import { projectTypeSoSettingsMap } from './constants.js';

const askQuestions = async () => {
  console.log(
    chalk.hex('#e7c99a')('Use the up and down arrow keys to navigate multi-choice questions'),
  );

  const questionsSetProjectType = [
    {
      name: 'type',
      type: 'list',
      message: 'Please choose a project type to use: ',
      choices: [
        {
          name: 'Basic',
          value: 'basics',
        },
        {
          name: 'Composite',
          value: 'composites',
        },
        {
          name: 'Plugin',
          value: 'plugins',
        },
      ],
      default: 'composites',
    },
  ];

  /* Initial set of questions */
  const getQuestionsSetFolderName = (templateDirectory, targetDirectory) => [
    {
      name: 'name',
      type: 'input',
      message: 'Project name(folder): ',
      validate: (projectName) => {
        let valid = true;

        /* Reg ex to ensure that project name starts with a letter, includes letters, numbers, underscores and hashes */
        if (/^[a-z]*-?[a-z]*$/gm.test(projectName)) {
          valid = valid && true;
        } else {
          return 'Project must \n 1) start with a letter \n 2) name may only include letters, numbers, underscores and hashes.';
        }

        /* Check that no folder exists at this location */
        if (!checkPathExists(templateDirectory)) {
          return 'Could not find a template for your selected choice';
        }

        /* Check that no folder exists at this location */
        if (checkPathExists(path.resolve(targetDirectory, projectName))) {
          return 'Project with this name already exists at this location';
        }

        return valid && true;
      },
    },
    {
      name: 'description',
      type: 'input',
      message: 'Project description: ',
    },
  ];

  try {
    /* Actually ask the questions */
    const answersA = await inquirer.prompt(questionsSetProjectType);

    const projectType = answersA.type;

    const { targetDirectoryPartialPath, templateDirectory, packagePartialPath } =
      projectTypeSoSettingsMap[projectType];

    const questionsB = getQuestionsSetFolderName(templateDirectory, targetDirectoryPartialPath);
    const answersB = await inquirer.prompt(questionsB);

    const projectName = answersB.name;

    /* Collate answers */
    const answers = {
      ...answersA,
      ...answersB,
      templateDirectory,
      targetDirectory: path.resolve(targetDirectoryPartialPath, projectName),
      testPath: path.join(packagePartialPath, projectName),
    };

    return answers;
  } catch (err) {
    if (err) {
      switch (err.status) {
        case 401:
          console.error('401');
          break;
        default:
          console.error(err);
      }
    }
  }

  return {};
};

export default askQuestions;

import fs from 'fs';

const checkPathExists = (pathToCheck) => {
  try {
    return fs.existsSync(pathToCheck);
  } catch (err) {
    console.error(err);
  }

  return false;
};

async function editPackageJSON(options) {
  const targetDir = options.targetDirectory;
  let jsonFile;

  await fs.readFile(`${targetDir}/package.json`, (err, data) => {
    /* If no package.json, this will be skipped  */
    if (!err) {
      jsonFile = JSON.parse(data);
      jsonFile.name = `@flyteconsole/${options.name}`;
      jsonFile.description = options.description;

      jsonFile.scripts.test = jsonFile.scripts.test.replace('folder-path', options.testPath);

      fs.writeFile(`${targetDir}/package.json`, JSON.stringify(jsonFile, null, '\t'), (err2) => {
        if (err2) {
          throw new Error('Unable to update package.json');
        }
      });
    }
  });
}

const mapToTemplates = {
  'Node-Express-Mongo-JS': 'jemn',
  'HTML,CSS,JS': 'basic',
};

export { checkPathExists, mapToTemplates, editPackageJSON };

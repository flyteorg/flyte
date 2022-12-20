import askQuestions from './questions.js';
import createProject from './main.js';

export async function cli() {
  const options = await askQuestions();
  createProject(options);
}

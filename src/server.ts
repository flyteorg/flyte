import * as express from 'express';
import * as fs from 'fs';
import * as path from 'path';

interface ServerRendererArguments {
  currentDirectory: string;
}

/**
 * Universal render function in development mode
 */
export default function serverRenderer({ currentDirectory }: ServerRendererArguments) {
  const html = fs.readFileSync(path.join(currentDirectory, 'dist/index.html')).toString();

  return (_req: express.Request, res: express.Response) => {
    if (html === '') {
      throw new ReferenceError('Could not find index.html to render');
    }

    res.status(200).send(html);
  };
}

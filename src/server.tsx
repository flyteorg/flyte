import * as cheerio from 'cheerio';
import * as express from 'express';
import * as fs from 'fs';
import MemoryFileSystem from 'memory-fs';
import * as path from 'path';
import { Helmet } from 'react-helmet';
import { processEnv } from '../env';

interface Stats {
  publicPath: string;
  assets: Array<{ name: string }>;
}

interface ServerRendererArguments {
  clientStats: Stats;
  fileSystem: MemoryFileSystem;
  currentDirectory: string;
}

/**
 * Universal render function in development mode
 */
export default function serverRenderer({ fileSystem, currentDirectory }: ServerRendererArguments) {
  const env = process.env.NODE_ENV || 'development';
  const isDev = env === 'development';
  const isProd = env === 'production';
  let html = '';
  if (isProd) {
    html = fs.readFileSync(path.join(currentDirectory, 'dist/index.html')).toString();
  }

  return (_req: express.Request, res: express.Response) => {
    if (isDev) {
      const indexPath = path.join(currentDirectory, 'dist', 'index.html');
      html = fileSystem.readFileSync(indexPath).toString();
    }

    if (html === '') {
      throw new ReferenceError('Could not find index.html to render');
    }

    // populate the app content...
    const $ = cheerio.load(html);

    // populate Helmet content
    const helmet = Helmet.renderStatic();
    $('head').append($.parseHTML(`${helmet.title.toString()} ${helmet.meta.toString()} ${helmet.link.toString()}`));

    // Populate process.env into window.env
    $('head').append($(`<script>window.env = ${JSON.stringify(processEnv)}</script>`));

    res.status(200).send($.html());
  };
}

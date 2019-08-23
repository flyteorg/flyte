import * as cheerio from 'cheerio';
import * as express from 'express';
import * as fs from 'fs';
// tslint:disable-next-line:import-name
import MemoryFileSystem from 'memory-fs';
import * as path from 'path';
import { Helmet } from 'react-helmet';

import { processEnv } from '../env';
// import { getConfig } from './configLoader';

// tslint:disable:prefer-array-literal
interface Stats {
    assetsByChunkName: {
        bootstrap: string[] | string;
    };
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
export default function serverRenderer({
    fileSystem,
    clientStats,
    currentDirectory
}: ServerRendererArguments) {
    const env = process.env.NODE_ENV || 'development';
    const isDev = env === 'development';
    const isProd = env === 'production';
    let html = '';
    if (isProd) {
        html = fs
            .readFileSync(path.join(currentDirectory, 'dist/index.html'))
            .toString();
    }

    return (
        req: express.Request,
        res: express.Response,
        next: express.NextFunction
    ) => {
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
        $('head').append(
            $.parseHTML(
                `${helmet.title.toString()} ${helmet.meta.toString()} ${helmet.link.toString()}`
            )
        );

        // Populate process.env into window.env
        $('head').append(
            $(`<script>window.env = ${JSON.stringify(processEnv)}</script>`)
        );

        // TODO: Populate this with the loaded config if/when we need to
        const initialData = {};

        $('body').prepend(
            $(
                `<script>window.__INITIAL_DATA__ = ${JSON.stringify(
                    initialData
                )}</script>`
            )
        );

        // add additional chunk scripts
        try {
            let bootstrapScriptName = clientStats.assetsByChunkName.bootstrap;
            if (Array.isArray(bootstrapScriptName)) {
                bootstrapScriptName = bootstrapScriptName.find(name =>
                    name.endsWith('.js')
                )!;
            }
        } catch (e) {
            // tslint:disable-next-line:no-console
            console.error(e);
        }

        res.status(200).send($.html());
    };
}

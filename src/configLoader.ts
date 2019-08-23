const Cache = require('cache');
import * as fs from 'fs';
// tslint:disable-next-line:import-name
import * as yaml from 'js-yaml';
import * as path from 'path';
import { promisify } from 'util';

import { Config, ConfigResult } from 'config';
import * as env from '../env';

const readFile = promisify(fs.readFile);
const readdir = promisify(fs.readdir);

const configDir = env.CONFIG_DIR;
const cacheTTL = Number.parseInt(env.CONFIG_CACHE_TTL_SECONDS, 10);
const cacheKey = 'CONFIG';

const cache = new Cache(cacheTTL * 1000);

const defaultConfig = {
    projects: [],
    domains: []
};

const isYamlFile = (fileName: string) =>
    fileName.substring(fileName.length - 5) === '.yaml';

async function getConfigFileList(configDir: string): Promise<string[]> {
    const fileList = await readdir(configDir);
    return fileList
        .filter(isYamlFile)
        .map(fileName => path.resolve(path.join(configDir, fileName)));
}

/** Parses a yaml file and returns it as an object to be merged into the overall config */
async function loadYamlFile(fileName: string): Promise<Config> {
    const contents = await readFile(fileName, 'utf8');
    return yaml.safeLoad(contents) as Config;
}

function validateConfig(config: Config) {}

/** Parses a list of Yaml files, each of which is expected to resolve to an object with keys.
 * These objects will be merged into a final config dictionary and returned.
 * Will fail if any file fails to load/parse.
 */
async function loadConfigFiles(): Promise<Config> {
    const fileList = await getConfigFileList(configDir);
    const filePromises = fileList.map(fileName => loadYamlFile(fileName));
    const configObjects = await Promise.all(filePromises);
    return configObjects.reduce(
        (out, config) => Object.assign(out, config),
        {}
    );
}

export async function getConfig(): Promise<ConfigResult> {
    const cached = cache.get(cacheKey);
    if (cached !== null) {
        return cached;
    }

    const result: ConfigResult = {};

    // Note: in either case, we'll cache the result of the load so we aren't hammering the
    // filesystem on every request.
    try {
        const loadedConfig = await loadConfigFiles();
        validateConfig(loadedConfig);
        result.data = Object.assign({}, defaultConfig, loadedConfig);
    } catch (e) {
        result.errorString = `${e}`;
    }
    cache.put(cacheKey, result);
    return result;
}

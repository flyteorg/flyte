import { VersionInfo } from './versionDisplay';

export const generateVersionInfo = (name: string, version: string): VersionInfo => {
  return {
    name: name,
    version: version,
    url: `#some.fake.link/${name}/v${version}`,
  };
};

import { useAdminVersion } from 'components/hooks/useVersion';
import { AdminFlag, AdminVersion, baseAdminConfig } from './defaultConfig';

export const useIsEnabledInAdmin = (flag: AdminFlag): boolean => {
  const { adminVersion } = useAdminVersion();
  if (!adminVersion || adminVersion === '') {
    return false;
  }

  // Split version to two array items - [Major][Minor.Patch]
  const versionSplit = adminVersion.replace(/\./, '&').split('&');
  const curVersion: AdminVersion = {
    major: parseInt(versionSplit[0], 10) ?? 0,
    minor: parseFloat(versionSplit[1]) ?? 0.1,
  };

  const requieredVersion = baseAdminConfig[flag] ?? null;
  // required version is less or equal current version - return true.
  if (
    requieredVersion &&
    (requieredVersion.major < curVersion.major ||
      (requieredVersion.major === curVersion.major && requieredVersion.minor <= curVersion.minor))
  ) {
    return true;
  }

  return false;
};

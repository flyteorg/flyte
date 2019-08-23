export interface Config {
    [k: string]: Object;
}

export interface ConfigResult {
    errorString?: string;
    data?: {};
}

export interface Env extends NodeJS.ProcessEnv {
    ADMIN_API_URL?: string;
    BASE_URL?: string;
    CORS_PROXY_PREFIX: string;
    DISABLE_ANALYTICS?: string;
}

export interface Env extends NodeJS.ProcessEnv {
    ADMIN_API_URL?: string;
    BASE_URL?: string;
    CORS_PROXY_PREFIX: string;
    DISABLE_ANALYTICS?: string;
    NODE_ENV?: 'development' | 'production' | 'test';
    STATUS_URL?: string;
}

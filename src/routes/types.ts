export interface Route {
  path?: string;
  makeUrl?: (...parts: string[]) => string;
}

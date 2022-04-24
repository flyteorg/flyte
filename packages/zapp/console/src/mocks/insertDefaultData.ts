import { AdminServer } from './createAdminServer';
import { projects } from './data/projects';

/** Inserts default global mock data. This can be extended by inserting additional
 * mock data fixtures.
 */
export function insertDefaultData(server: AdminServer): void {
  server.insertProjects([projects.flyteTest]);
}

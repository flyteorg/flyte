/** Indicates failure to fetch a resource because it does not exist (404) */
export class NotFoundError extends Error {
    constructor(
        public name: string,
        msg: string = 'The requested item could not be found'
    ) {
        super(msg);
    }
}
/** Indicates failure to fetch a resource because the user is not authorized (401) */
export class NotAuthorizedError extends Error {
    constructor(msg: string = 'User is not authorized to view this resource') {
        super(msg);
    }
}

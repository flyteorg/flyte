/** Indicates failure to fetch a resource */
export class NotFoundError extends Error {
    constructor(
        public name: string,
        msg: string = 'The requested item could not be found'
    ) {
        super(msg);
    }
}

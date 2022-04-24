/** Indicates that a protobuf message is missing one or more expected fields */
export class MessageMissingRequiredFieldsError extends Error {
  constructor(
    public fields: string[],
    msg = `Message is missing required fields: ${fields.join(',')}`,
  ) {
    super(msg);
  }
}

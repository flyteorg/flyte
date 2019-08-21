package validators

func ValidateEmptyStringField(field, fieldName string) error {
	if field == "" {
		return NewMissingArgumentError(fieldName)
	}
	return nil
}

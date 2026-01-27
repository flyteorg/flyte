package service

// TruncateShortDescription truncates the short description to 255 characters if it exceeds the maximum length.
func truncateShortDescription(description string) string {
	if len(description) > 255 {
		return description[:255]
	}
	return description
}

// truncateLongDescription truncates the long description to 2048 characters if it exceeds the maximum length.
func truncateLongDescription(description string) string {
	if len(description) > 2048 {
		return description[:2048]
	}
	return description
}

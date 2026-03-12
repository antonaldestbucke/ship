package shipinternal

func firstNonEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

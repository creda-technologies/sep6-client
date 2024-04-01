package sep6-client

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Field + " - " + e.Message
}

type FetchError struct {
	Field   string
	Message string
}

func (e *FetchError) Error() string {
	return "validation error: " + e.Field + " - " + e.Message
}

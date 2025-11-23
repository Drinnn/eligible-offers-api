package errors

type ServiceError struct {
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewServiceError(message string) error {
	return &ServiceError{Message: message}
}

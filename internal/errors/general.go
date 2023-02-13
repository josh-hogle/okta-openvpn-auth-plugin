package errors

// None indicates there is no error at all.
type None struct {
}

// InternalError returns the internal error object.
func (e *None) InternalError() error {
	return nil
}

// Error returns the string version of the error.
func (e *None) Error() string {
	return "the command completed successfully"
}

// Code returns the corresponding error code.
func (e *None) Code() int {
	return NoneCode
}

// Usage indicates there was a usage error.
type Usage struct {
	Err error
}

// InternalError returns the internal error object.
func (e *Usage) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *Usage) Error() string {
	return e.Err.Error()
}

// Code returns the corresponding error code.
func (e *Usage) Code() int {
	return UsageCode
}

// GeneralFailure indicates there was a general system error.
type GeneralFailure struct {
	Msg string
	Err error
}

// InternalError returns the internal error object.
func (e *GeneralFailure) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *GeneralFailure) Error() string {
	return e.Msg
}

// Code returns the corresponding error code.
func (e *GeneralFailure) Code() int {
	return GeneralFailureCode
}

// Package errors provides application error handling functionality.
//
// Any methods in packages which return an error object will return a custom error type implementing the
// go.innotegrity.dev/toolbox/errors.ExtendedError interface. The specific errors returned by the method are noted in
// its documentation. Each integer code corresponding the the specific error is unique across the application.
//
// To determine the specific type of error that was returned, you can use 1 of 3 methods:
//
//	 ◽ Use the errors.Is() method with new(ErrType) as the second argument where ErrType is the actual type of error
//	   that is expected.
//	 ◽ Use the errors.As() method to determine if you can convert the error into the given object. The advantage of
//			this method is that you can then also access each error's custom fields.
//	 ◽ Cast the error object to an ExtendedError object and then use the Code() method to compare the error code to the
//	   corresponding error code constant defined as the error type with 'Code' appended to it (eg:
//	   ErrDoRequestFailureCode, ErrResourceWaitFailureCode)
package errors

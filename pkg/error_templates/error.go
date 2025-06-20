package error_templates

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
OutputError struct that realize Errorer interface
errorMessage is field for human-understandable error message
errorDetail is field for detail error message for logging
grpcStatusCode is field for grpc response code
httpStatusCode is field for http response code.
*/
type OutputError struct {
	errorMessage   string
	errorDetail    error
	grpcStatusCode codes.Code
	httpStatusCode int
}

// Error returns OutputError.errorMessage field.
func (e *OutputError) Error() string {
	return e.errorMessage
}

// ErrorDetail returns OutputError.errorDetail field.
func (e *OutputError) ErrorDetail() error {
	return e.errorDetail
}

// ErrorDetailFromError returns OutputError.errorDetail field if err is OutputError and input error if not.
func ErrorDetailFromError(err error) error {
	if outErr, ok := err.(*OutputError); ok {
		return outErr.ErrorDetail()
	}

	return err
}

// New is the constructor for OutputError/
func New(errorMessage string, errorDetail error, grpcCode codes.Code, httpCode int) *OutputError {
	return &OutputError{
		errorMessage:   errorMessage,
		errorDetail:    errorDetail,
		grpcStatusCode: grpcCode,
		httpStatusCode: httpCode,
	}
}

// WrapErrorDetail is the function that wraps OutputError.errorDetail field by input message/
func WrapErrorDetail(err error, wrapInfo string) error {
	if outErr, ok := err.(*OutputError); ok {
		outErr.errorDetail = errors.Wrap(outErr.errorDetail, wrapInfo)
		return outErr
	}

	return New(wrapInfo, err, codes.Unknown, http.StatusInternalServerError)
}

// GetGRPC returns formed grpc status error by OutputError info.
func (e *OutputError) GetGRPC() error {
	return status.Error(e.grpcStatusCode, e.errorMessage)
}

// GetHTTP returns data  for http error response by OutputError info.
func (e *OutputError) GetHTTP() (int, string) {
	return e.httpStatusCode, e.errorMessage
}

// HandleResponseErrorGRPC is a function that wraps or forms error from grpc client.
func HandleResponseErrorGRPC(err error, wrapInfo string, grpcCode ...codes.Code) error {
	var outErr = new(OutputError)
	if stat, ok := status.FromError(err); ok {
		if len(grpcCode) != 0 {
			outErr = New(stat.Message(), errors.New(stat.Message()), grpcCode[0], ResponseCodeFromGRPCToHTTP(grpcCode[0]))
		} else {
			outErr = New(stat.Message(), errors.New(stat.Message()), stat.Code(), ResponseCodeFromGRPCToHTTP(stat.Code()))
		}

		return WrapErrorDetail(outErr, wrapInfo)
	}

	if len(grpcCode) != 0 {
		outErr = New(err.Error(), err, grpcCode[0], ResponseCodeFromGRPCToHTTP(grpcCode[0]))
	} else {
		outErr = New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
	}
	return WrapErrorDetail(outErr, wrapInfo)
}

func ResponseCodeFromGRPCToHTTP(codeGRPC codes.Code) (codeHTTP int) {
	switch codeGRPC {
	case codes.Canceled, codes.Unknown, codes.Internal, codes.DataLoss:
		codeHTTP = http.StatusInternalServerError
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		codeHTTP = http.StatusBadRequest
	case codes.DeadlineExceeded:
		codeHTTP = http.StatusGatewayTimeout
	case codes.NotFound:
		codeHTTP = http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		codeHTTP = http.StatusConflict
	case codes.PermissionDenied:
		codeHTTP = http.StatusForbidden
	case codes.ResourceExhausted:
		codeHTTP = http.StatusTooManyRequests
	case codes.Unimplemented:
		codeHTTP = http.StatusNotImplemented
	case codes.Unauthenticated:
		codeHTTP = http.StatusUnauthorized
	default:
		codeHTTP = http.StatusInternalServerError
	}
	return
}

func ResponseCodeFromHTTPToGRPC(codeHTTP int) (codeGRPC codes.Code) {
	switch codeHTTP {
	case http.StatusInternalServerError:
		codeGRPC = codes.Internal
	case http.StatusBadRequest:
		codeGRPC = codes.Internal
	case http.StatusGatewayTimeout:
		codeGRPC = codes.DeadlineExceeded
	case http.StatusNotFound:
		codeGRPC = codes.NotFound
	case http.StatusConflict:
		codeGRPC = codes.AlreadyExists
	case http.StatusForbidden:
		codeGRPC = codes.PermissionDenied
	case http.StatusTooManyRequests:
		codeGRPC = codes.ResourceExhausted
	case http.StatusNotImplemented:
		codeGRPC = codes.Unimplemented
	case http.StatusUnauthorized:
		codeGRPC = codes.Unauthenticated
	default:
		codeGRPC = codes.Unknown
	}
	return
}

func WrapErrorEndpoint(err error, reqID string) error {
	var outErr *OutputError
	if errors.As(err, &outErr) {
		if outErr.grpcStatusCode == codes.Unavailable || outErr.grpcStatusCode == codes.Internal {
			msg := fmt.Sprintf("internal error, request_id: %s", reqID)
			err = errors.New(msg)
		}
	} else {
		if stat, ok := status.FromError(err); ok {
			if stat.Code() == codes.Unavailable || stat.Code() == codes.Internal {
				msg := fmt.Sprintf("internal error, request_id: %s", reqID)
				err = errors.New(msg)
			}
		}
	}
	return err
}

func BadRequestError(err error) error {
	return New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
}

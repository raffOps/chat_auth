package errs

import (
	"errors"
	"fmt"
	"github.com/raffops/chat/pkg/logger"
	"go.uber.org/zap"
	"runtime/debug"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrBadRequest       = errors.New("bad request")
	ErrInternal         = errors.New("internal server error")
	ErrNotAuthorized    = errors.New("not authorized")
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrConflict         = errors.New("conflict")
)

type ChatError interface {
	error
	AppError() error
	SvcError() error
}

type authError struct {
	appError error
	svcError error
}

func (e authError) AppError() error {
	return e.appError
}

func (e authError) SvcError() error {
	return e.svcError
}

func (e authError) Error() string {
	if e.AppError() == nil {
		return e.SvcError().Error()
	}
	return fmt.Sprintf("%s: %s", e.svcError.Error(), e.appError.Error())
}

func NewError(svcError, appError error) ChatError {
	if errors.Is(svcError, ErrInternal) {
		logger.Debug(
			svcError.Error(),
			zap.Error(appError),
			zap.String("stack", string(debug.Stack())),
		)
		return &authError{svcError: svcError, appError: nil}
	}
	return &authError{appError: appError, svcError: svcError}
}

func GetHttpStatusCode(err ChatError) int {
	switch err := err.SvcError(); {
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrBadRequest):
		return 400
	case errors.Is(err, ErrInternal):
		return 500
	case errors.Is(err, ErrNotAuthorized):
		return 401
	case errors.Is(err, ErrNotAuthenticated):
		return 401
	case errors.Is(err, ErrConflict):
		return 409
	default:
		return 500
	}
}

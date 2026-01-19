package xerr

import "fmt"

type HttpError struct {
	httpCode int
	errMsg   string
}

func (h *HttpError) GetHttpCode() int {
	return h.httpCode
}

func (h *HttpError) GetErrMsg() string {
	return h.errMsg
}

func (h *HttpError) Error() string {
	return fmt.Sprintf("HttpCode:%d，ErrMsg:%s", h.httpCode, h.errMsg)
}

func NewHttpError(httpCode int, errMsg string) *HttpError {
	return &HttpError{
		httpCode: httpCode,
		errMsg:   errMsg,
	}
}

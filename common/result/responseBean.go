package result

type ResponseSuccessBean[T any] struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
type NullJson struct{}

func Success[T any](data T) *ResponseSuccessBean[T] {
	return &ResponseSuccessBean[T]{0, "OK", data}
}

type ResponseErrorBean struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
}

func Error(errCode uint32, errMsg string) *ResponseErrorBean {
	return &ResponseErrorBean{errCode, errMsg}
}

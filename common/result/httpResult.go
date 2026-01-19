package result

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tespkg/bytes-be/common/xerr"

	"net/http"
)

func HttpResult(w http.ResponseWriter, resp interface{}, err error) {
	if err == nil {
		WriteJson(w, http.StatusOK, Success(resp))
	} else {
		errCode := xerr.ServerCommonError
		errMsg := "The server is out of service, try again later"

		httpCode := http.StatusInternalServerError

		causeErr := errors.Cause(err)
		if e, ok := causeErr.(*xerr.CodeError); ok {
			errCode = e.GetErrCode()
			errMsg = e.GetErrMsg()

			httpCode = http.StatusOK
		}

		if e, ok := causeErr.(*xerr.HttpError); ok {
			httpCode = e.GetHttpCode()
			errMsg = e.GetErrMsg()
		}

		//WriteJson(w, http.StatusBadRequest, Error(errCode, errMsg))
		WriteJson(w, httpCode, Error(errCode, errMsg))
	}
}

func ParamErrorResult(w http.ResponseWriter, err error) {
	errMsg := fmt.Sprintf("%s ,%s", xerr.MapErrMsg(xerr.RequestParamError), err.Error())
	WriteJson(w, http.StatusBadRequest, Error(xerr.RequestParamError, errMsg))
}

// WriteJson writes v as json string into w with code.
func WriteJson(w http.ResponseWriter, code int, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if n, err := w.Write(bs); err != nil {
		// http.ErrHandlerTimeout has been handled by http.TimeoutHandler,
		// so it's ignored here.
		if !errors.Is(err, http.ErrHandlerTimeout) {
			logrus.Errorf("write response failed, error: %s", err)
		}
	} else if n < len(bs) {
		logrus.Errorf("actual bytes: %d, written bytes: %d", len(bs), n)
	}
}

// Ok writes HTTP 200 OK into w.
func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// OkJson writes v into w with 200 OK.
func OkJson(w http.ResponseWriter, v interface{}) {
	WriteJson(w, http.StatusOK, v)
}

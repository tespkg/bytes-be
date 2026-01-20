package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tespkg/bytes-be/common/result"
	"github.com/tespkg/bytes-be/svc/staff/logic"
	"github.com/tespkg/bytes-be/svc/staff/model/dto"
)

// SendVerifyCode
// @Summary send verify code
// @Tags Common
// @Accept json
// @Produce json
// @Param email body string false "email"
// @Param phone body string false "phone"
// @Success 200 {object} result.ResponseSuccessBean[NullJson]
// @Router /api/v1/common/code/verify [post]
func (s *Server) SendVerifyCode(c *gin.Context) {
	var req *dto.SendVerifyCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Error("c.ShouldBindJSON fail:", err)
		result.ParamErrorResult(c.Writer, errors.New("c.ShouldBindJSON fail"))
		return
	}

	if req.Phone == "" && req.Email == "" {
		result.ParamErrorResult(c.Writer, errors.New("phone or email is required"))
		return
	}

	err := logic.SendVerifyCode(c.Request.Context(), req)
	if err != nil {
		logrus.Errorf("send verify code fail: %s", err)
	}
	result.HttpResult(c.Writer, nil, err)
}

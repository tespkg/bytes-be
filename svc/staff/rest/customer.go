package rest

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/tespkg/bytes-be/common/global"
	"github.com/tespkg/bytes-be/common/result"
	"github.com/tespkg/bytes-be/svc/staff/logic"
	"github.com/tespkg/bytes-be/svc/staff/model/dao"
	"github.com/tespkg/bytes-be/svc/staff/model/dto"
)

// Register
// @Summary customer register
// @Tags Customer
// @Accept json
// @Produce json
// @Param req body dto.CustomerRegisterReq true "customer register request"
// @Success 200 {object} result.ResponseSuccessBean[dto.CustomerRegisterResp]
// @Router /api/v1/customer/register [post]
func (s *Server) Register(c *gin.Context) {
	var req *dto.CustomerRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Error("c.ShouldBindJSON fail:", err)
		result.ParamErrorResult(c.Writer, errors.New("c.ShouldBindJSON fail"))
		return
	}

	//check phone or email is registered
	if req.Phone != "" {
		user, err := dao.GetUserByPhoneAndRole(s.db, req.Phone, dao.RoleCustomer)
		if err != nil {
			logrus.Error("dao.GetUserByPhoneAndRole fail:", err)
			result.HttpResult(c.Writer, nil, err)
			return
		}
		if user != nil && user.Id > 0 {
			result.ParamErrorResult(c.Writer, errors.New("This phone is registered a customer."))
			return
		}
	}

	if req.Email != "" {
		user, err := dao.GetUserByEmailAndRole(s.db, req.Email, dao.RoleCustomer)
		if err != nil {
			logrus.Error("dao.GetUserByEmailAndRole fail:", err)
			result.HttpResult(c.Writer, nil, err)
			return
		}
		if user != nil && user.Id > 0 {
			result.ParamErrorResult(c.Writer, errors.New("This email is registered a customer."))
			return
		}
	}

	//verify code
	key := fmt.Sprintf("bytes_be:verify_code:phone:%s", req.Phone)
	val, err := global.GlobalClientSets.RedisClient.Get(c.Request.Context(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			result.ParamErrorResult(c.Writer, errors.New("verify code not found"))
			return
		}

		result.ParamErrorResult(c.Writer, err)
		return
	}
	if req.Code != val {
		result.ParamErrorResult(c.Writer, errors.New("email verify code fail"))
		return
	}

	//customer register
	resp, err := logic.CustomerRegister(s.db, req)
	if err != nil {
		logrus.Errorf("customer register fail: %s", err)
	}
	result.HttpResult(c.Writer, resp, err)
}

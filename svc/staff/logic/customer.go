package logic

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tespkg/bytes-be/common/token"
	"github.com/tespkg/bytes-be/svc/staff/model/dao"
	"github.com/tespkg/bytes-be/svc/staff/model/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CustomerRegister(session *gorm.DB, req *dto.CustomerRegisterReq) (*dto.CustomerRegisterResp, error) {
	//get role
	role, err := dao.GetRoleByName(session, dao.RoleCustomer)
	if err != nil {
		return nil, errors.Wrap(err, ">>CustomerRegister, get role fail")
	}
	if role == nil || role.Id == 0 {
		return nil, errors.New(">>CustomerRegister, role not found")
	}

	tx := session.Begin()
	if err = tx.Error; err != nil {
		return nil, errors.Wrap(err, ">>CustomerRegister, transaction begin fail")
	}
	defer tx.Rollback()

	//create user
	var userId int64
	user := dao.User{
		Uuid:      uuid.NewString(),
		Email:     &(req.Email),
		Phone:     &(req.Phone),
		Password:  nil,
		IsEnabled: true,
	}
	hash, generatePasswordErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if generatePasswordErr != nil {
		return nil, errors.Wrap(generatePasswordErr, ">>CustomerRegister, bcrypt.GenerateFromPassword fail")
	}
	user.Password = hash
	if err = user.Save(tx); err != nil {
		return nil, errors.Wrap(err, ">>CustomerRegister, user.Save fail")
	}
	userId = user.Id

	//create user_role
	userRole := dao.UserRole{
		UserId: userId,
		RoleId: role.Id,
	}
	if err = userRole.Save(tx); err != nil {
		return nil, errors.Wrap(err, ">>CustomerRegister, userRole.Save fail")
	}

	//create customer
	customer := dao.Customer{
		UserId: userId,
	}
	if err = customer.Save(tx); err != nil {
		return nil, errors.Wrap(err, ">>CustomerRegister, customer.Save fail")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.Wrap(err, ">>CustomerRegister, transaction commit fail")
	}

	//general login token
	tokenStr, err := token.GenToken(&token.GenTokenDto{
		Session:       session,
		UserId:        fmt.Sprintf("%d", userId),
		Platform:      "",
		Imei:          "",
		ClientVersion: "",
		Model:         "",
		SystemVersion: "",
	})
	if err != nil {
		return nil, errors.Wrap(err, ">>CustomerRegister, token.GenToken fail")
	}

	return &dto.CustomerRegisterResp{
		LoginToken: tokenStr,
	}, nil
}

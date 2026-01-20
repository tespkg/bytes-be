package dto

type SendVerifyCodeReq struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}

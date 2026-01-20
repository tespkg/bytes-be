package dto

type CustomerRegisterReq struct {
	Email    string `json:"email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type CustomerRegisterResp struct {
	LoginToken string `json:"loginToken"`
}

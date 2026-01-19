package xerr

const OK uint32 = 200

// ServerCommonError global error code
const ServerCommonError uint32 = 100001
const RequestParamError uint32 = 100002
const TokenExpireError uint32 = 100003
const TokenGenerateError uint32 = 100004
const DbError uint32 = 100005

const DeliveryTimeOutOperatingHours uint32 = 40

const (
	UserNotExist           = 100006
	UserPasswordInvalid    = 100007
	TraveICAuthTokenError  = 100008
	OrderNotExist          = 100009
	TicketNotExist         = 100010
	AccountSettingNotExist = 100011
	EmailRegistered        = 100012
	ThirdPartyLoginFail    = 100013
)

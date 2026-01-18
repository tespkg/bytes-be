package token

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/tespkg/bytes-be/common/global"
	"gorm.io/gorm"
	"time"
)

const (
	Issuer    = "bytes-be"
	jwtSecret = "z4kP5aDMcR#o[dgV"
)

type UserClaims struct {
	Platform      string `json:"platform,omitempty"` //ios, android
	IMEI          string `json:"imei,omitempty"`
	ClientVersion string `json:"clientVersion,omitempty"` // client version客户端版本
	Model         string `json:"model,omitempty"`         // phone model手机型号
	SystemVersion string `json:"systemVersion,omitempty"` // phone system version手机操作系统版本
	jwt.RegisteredClaims
}

func GenClaims(dto *GenTokenDto) UserClaims {
	defaultTokenExpireDuration := global.GlobalClientSets.TokenExpireDurationHour

	return UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(defaultTokenExpireDuration))),
			Issuer:    Issuer,
			ID:        dto.UserId,
			Subject:   dto.UserId,
		},
		Platform:      dto.Platform,
		IMEI:          dto.Imei,
		ClientVersion: dto.ClientVersion,
		Model:         dto.Model,
		SystemVersion: dto.SystemVersion,
	}
}

type GenTokenDto struct {
	Session       *gorm.DB
	UserId        string
	Platform      string
	Imei          string
	ClientVersion string
	Model         string
	SystemVersion string
}

func GenToken(dto *GenTokenDto) (string, error) {
	claims := GenClaims(dto)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	var secret string
	var err error
	if global.GlobalClientSets.EnableSingleLogin {
		secret, err = RefreshUserSecret(dto)
		if err != nil {
			return "", errors.Wrap(err, ">>GenToken ")
		}
	} else {
		secret, err = GetUserSecret(dto)
		if err != nil {
			return "", errors.Wrap(err, ">>GenToken ")
		}
	}

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.Wrap(err, ">>GenToken ")
	}
	return signedToken, nil
}

var TokenFormatError = errors.New("token format error")
var TokenInvalidError = errors.New("invalid token")
var TokenExpiredError = errors.New("expired token")
var TokenUnknownError = errors.New("unknown token error")
var TokenClaimsError = errors.New("token claims error")
var TokenSignatureInvalidError = errors.New("token invalid signature error")

func VerifyToken(ctx context.Context, session *gorm.DB, tokenString string) (context.Context, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if claims, ok := token.Claims.(*UserClaims); !ok {
			return nil, TokenClaimsError
		} else {
			secret, err := GetUserSecret(&GenTokenDto{
				Session:       session,
				UserId:        claims.ID,
				Platform:      "",
				Imei:          "",
				ClientVersion: "",
				Model:         "",
				SystemVersion: "",
			})
			if err != nil {
				return nil, err
			}
			return []byte(secret), nil
		}

		//if third.GlobalClients.EnableSingleLogin {
		//	if claims, ok := token.Claims.(*WalletClaims); !ok {
		//		return nil, TokenClaimsError
		//	} else {
		//		secret, err := GetUserSecret(session, claims.ID)
		//		if err != nil {
		//			return nil, err
		//		}
		//		return []byte(secret), nil
		//	}
		//}
		//
		//return []byte(jwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return ctx, TokenFormatError
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return ctx, TokenExpiredError
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return ctx, TokenExpiredError
		} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return ctx, TokenSignatureInvalidError
		} else {
			return ctx, TokenUnknownError
		}
	}

	if !token.Valid {
		return ctx, TokenInvalidError
	}

	if claims, ok := token.Claims.(*UserClaims); ok {
		// fmt.Printf("claims %+v\n", claims)
		ctx = context.WithValue(ctx, ClaimsCtx, claims)
		return ctx, nil
	} else {
		return ctx, TokenClaimsError
	}
}

package middle

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tespkg/bytes-be/common/token"
	"github.com/tespkg/bytes-be/svc/staff/model/dao"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

func WithToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		ctx := r.Context()

		// access token from authorization header
		accessToken, err := extractTokenFromHeader(r)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		} else {
			ctx, err = token.VerifyToken(ctx, db, accessToken)
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusUnauthorized)
				c.Abort()
				return
			}
			c.Set("token", accessToken)
			claims := ctx.Value(token.ClaimsCtx).(*token.UserClaims)
			c.Set("claims", claims)
			c.Next()
		}
	}
}

func extractTokenFromHeader(r *http.Request) (string, error) {
	authHeaders, ok := r.Header["Authorization"]
	if !ok {
		return "", errors.New("Authorization header is empty")
	}
	if len(authHeaders) != 1 {
		return "", errors.New("More than one Authorization headers sent")
	}

	parts := strings.SplitN(authHeaders[0], " ", 2)
	if len(parts) != 2 {
		return "", errors.New("Bad Authorization header")
	}
	if parts[0] != "Bearer" {
		return "", errors.New("Only Bearer tokens accepted")
	}

	return parts[1], nil
}

func WithUserInfo(db *gorm.DB) gin.HandlerFunc {
	EmptyClaimsErr := errors.New("empty claims")
	EmptyUserErr := errors.New("empty user")

	return func(c *gin.Context) {
		claims, ok := c.Get("claims")
		if !ok {
			http.Error(c.Writer, EmptyClaimsErr.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		userClaims, _ := claims.(*token.UserClaims)
		userId := userClaims.Subject
		c.Set("user_id", userId)
		uId, _ := strconv.ParseInt(userId, 10, 64)

		user, err := dao.GetUserById(db, uId)
		if err != nil {
			http.Error(c.Writer, EmptyUserErr.Error(), http.StatusInternalServerError)
			c.Abort()
			return
		}
		if user == nil || user.Id == 0 {
			http.Error(c.Writer, EmptyUserErr.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

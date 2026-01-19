package middle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

func WithTimezone() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		timezone, err := extractTimezoneFromHeader(r)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			c.Abort()
			return
		}
		fmt.Printf("request %s with timezone %s\n", r.URL.Path, timezone)
		c.Set("timezone", timezone)
		c.Next()
	}
}

func extractTimezoneFromHeader(r *http.Request) (string, error) {
	timezoneHeaders, ok := r.Header["Timezone"]
	if !ok {
		fmt.Println("missing timezone in header")
		return "", nil
	}
	if len(timezoneHeaders) != 1 {
		return "", errors.New("More than one timezone headers sent")
	}

	return timezoneHeaders[0], nil
}

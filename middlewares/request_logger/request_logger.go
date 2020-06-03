package requestlogger

import (
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vavas/go_1month_challenge/internals/stats"
	"github.com/vavas/go_services/utils"
)

// RequestLoggerHeaders contains all white-listed headers that will be included
// in incoming request log messages
var RequestLoggerHeaders = []string{
	"accept",
	"accept-charset",
	"accept-encoding",
	"accept-language",
	"cache-control",
	"connection",
	"content-encoding",
	"content-language",
	"content-length",
	"content-location",
	"content-range",
	"content-type",
	"dnt",
	"date",
	"forwarded",
	"from",
	"host",
	"keep-alive",
	"last-modified",
	"location",
	"origin",
	"range",
	"referer",
	"te",
	"tk",
	"upgrade-insecure-requests",
	"user-agent",
	"via",
	"warning",
	"x-forwarded-for",
	"x-forwarded-host",
	"x-forwarded-proto",
}

// RequestLogger is gin middleware that logs information about requests.
// Two log messages are generated. The first log message is output when
// this middleware first encounters a request. It contains whitelisted
// header values, the remote address, HTTP method, url, and request id.
// The second log message is output when the request has been handled
// and the response has been sent. This message includes the time duration
// required by the server to process the request, as well as the remote
// address, HTTP method, url, and request id. This middleware also reports
// error rate statistics.
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	//logger = logger.With(zap.Field{
	//	Key: "context",
	//	String: "Request",
	//})

	return func(c *gin.Context) {
		requestID, err := utils.RandomHex(16)
		if err != nil {
			utils.NotifyError(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Header.Set("REQUEST-ID", requestID)
		c.Set("requestID", requestID)

		reqLogFields := map[string]interface{}{
			"remote_address": c.ClientIP(),
			"method":         c.Request.Method,
			"url":            c.Request.URL,
			"content_length": c.Request.ContentLength,
			"request_id":     requestID,
		}
		for _, header := range RequestLoggerHeaders {
			if headerValue := c.Request.Header.Get(header); len(headerValue) > 0 {
				headerKey := strings.Replace(header, "-", "_", -1)
				reqLogFields[headerKey] = headerValue
			}
		}
		c.Set("logger", reqLogFields)

		startTime := time.Now()
		c.Next()
		endTime := time.Now()

		duration := endTime.Sub(startTime)
		statusCode := c.Writer.Status()

		logger.Info("Outgoing response",
			zap.String("remote_address", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.Any("url", c.Request.URL),
			zap.Int("status_code", statusCode),
			zap.Any("duration", duration),
			zap.String("request_id", requestID),
		)

		if !strings.HasPrefix(c.Request.URL.Path, "/.") {
			go stats.UpdateStats(duration.Nanoseconds(), statusCode >= 400 && statusCode < 500, statusCode >= 500)
		}
	}
}

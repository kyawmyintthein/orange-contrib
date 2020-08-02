package middlewarex

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type RequestIDCfg struct {
	IncomingRequestID string `json:"incoming_requet_id" mapstructure:"incoming_requet_id"`
}

type RequestIDMW interface {
	RequestID(http.Handler) http.Handler
	GinRequestID() gin.HandlerFunc
}

type requestIDMW struct {
	config *RequestIDCfg
}

const RequestIDKey string = "RequestIDKey"

var (
	prefix string
	reqID  uint64
)

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

func NewRequestIDMW(cfg *RequestIDCfg) RequestIDMW {
	return &requestIDMW{
		config: cfg,
	}
}

func (mw *requestIDMW) RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			myid := atomic.AddUint64(&reqID, 1)
			requestID = fmt.Sprintf("%s-%06d", prefix, myid)
		}
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func (mw *requestIDMW) GinRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for incoming header, use it if exists
		requestID := c.Request.Header.Get("X-Request-Id")
		// Create request id with UUID4
		if requestID == "" {
			myid := atomic.AddUint64(&reqID, 1)
			requestID = fmt.Sprintf("%s-%06d", prefix, myid)
		}
		// Expose it for use in the application
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)
		// Set X-Request-Id header
		c.Request.Header.Set("X-Request-Id", requestID)
		c.Next()
	}
}

// GetReqID returns a request ID from the given context if one is present.
// Returns the empty string if a request ID cannot be found.
func GetReqID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// NextRequestID generates the next request ID in the sequence.
func NextRequestID() uint64 {
	return atomic.AddUint64(&reqID, 1)
}

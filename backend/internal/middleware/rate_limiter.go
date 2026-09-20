package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens float64
	last   time.Time
}

// RateLimiter 用户维度令牌桶限流。
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
	burst   float64
}

// NewRateLimiter 构造限流器。
func NewRateLimiter(perMin int) *RateLimiter {
	return &RateLimiter{buckets: map[string]*bucket{}, rate: float64(perMin) / 60.0, burst: float64(perMin)}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.burst, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Limit 限流中间件。
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if v, ok := c.Get(UserIDKey); ok {
			key = "u" + itoa(v)
		}
		if !rl.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": constants.CodeRateLimited, "message": constants.MsgRateLimited, "data": nil,
			})
			return
		}
		c.Next()
	}
}

func itoa(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if u, ok := v.(uint); ok {
		return strconv.FormatUint(uint64(u), 10)
	}
	return "0"
}

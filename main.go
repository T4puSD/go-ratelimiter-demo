package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

type RateLimiter struct {
	limit       int
	interval    time.Duration
	redisClient *redis.Client
	luaScriptID string
}

func (limiter *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		rateLimited, err := limiter.isRateLimited(clientIP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
			c.Abort()
			return
		}

		if rateLimited {
			c.JSON(http.StatusTooManyRequests, gin.H{"message": "Rate limit exceeded"})
			c.Abort()
			return
		}

		// Execute the stored Lua script in Redis to increment the rate and handle expiration
		if err := limiter.redisClient.EvalSha(context.Background(), limiter.luaScriptID, []string{clientIP}, limiter.interval.Seconds()).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (limiter *RateLimiter) isRateLimited(clientIP string) (bool, error) {
	counter, err := limiter.redisClient.Get(context.Background(), clientIP).Int64()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, err
	}

	if counter > int64(limiter.limit) {
		return true, nil
	}

	return false, err
}

func main() {
	router := gin.Default()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Replace with your Redis server address
		Password: "",               // Replace with your Redis password
		DB:       0,                // Replace with the desired Redis database index
	})

	rateLimiterScript := `
		-- KEYS[1] is the key name
		-- ARGV[1] is the timeToLeaveInSeconds

		local counter = redis.call("incr",KEYS[1])
		if counter == 1 then
			redis.call("expire",KEYS[1],ARGV[1])
		end
		return counter
		`

	luaScriptID, err := redisClient.ScriptLoad(context.Background(), rateLimiterScript).Result()
	if err != nil {
		panic(err)
	}

	limiter := &RateLimiter{
		limit:       50,              // Maximum number of requests allowed
		interval:    5 * time.Minute, // Time interval for rate limiting
		redisClient: redisClient,     // Redis client instance
		luaScriptID: luaScriptID,     // Lua script ID
	}

	// Apply rate limiting middleware to the desired routes
	router.GET("/limited-route", limiter.RateLimit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Limited route"})
	})

	router.Run(":8080")
}

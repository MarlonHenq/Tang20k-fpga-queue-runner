package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Submission struct {
	Key      string `json:"key"`
	Code     string `json:"code"`
	Exercise string `json:"exercise"`
}

var ctx = context.Background()
var redisClient *redis.Client

func main() {
	godotenv.Load()
	r := gin.Default()

	redisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	r.POST("/submit", func(c *gin.Context) {
		var sub Submission
		if err := c.BindJSON(&sub); err != nil {
			c.JSON(400, gin.H{"error": "bad request"})
			return
		}
		if sub.Key != os.Getenv("API_KEY") {
			c.JSON(403, gin.H{"error": "invalid key"})
			return
		}

		data, _ := json.Marshal(sub)
		err := redisClient.LPush(ctx, "verilog_jobs", data).Err()
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to enqueue"})
			return
		}
		c.JSON(202, gin.H{"status": "queued"})
	})

	r.Run(":8080")
}

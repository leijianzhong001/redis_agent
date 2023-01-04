package utils

import (
	"context"
	"fmt"
	"testing"
)

var ctx = context.Background()

func TestRedisClient(t *testing.T) {
	redisClient := GetRedisClient()
	result, err := redisClient.Info(ctx).Result()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(result)
}

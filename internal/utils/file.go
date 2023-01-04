package utils

import (
	"github.com/leijianzhong001/redis_agent/internal/log"
	"os"
)

func DoesFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.PanicError(err)
		}
	}
	return true
}

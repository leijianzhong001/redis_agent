package utils

import (
	"fmt"
	"testing"
)

func TestMallocOverhead(t *testing.T) {
	bin := mallocOverhead(20)
	if bin != 32 {
		t.Error("")
	}
}

func TestZslRandomLevelFunc(t *testing.T) {
	for i := 0; i < 10000; i++ {
		fmt.Println(ZslRandomLevel())
	}
}

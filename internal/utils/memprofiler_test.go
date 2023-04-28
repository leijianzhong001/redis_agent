package utils

import (
	"fmt"
	"testing"
)

func TestMallocOverhead(t *testing.T) {
	bin := MallocOverhead(20)
	if bin != 32 {
		t.Error("")
	}
}

func TestZslRandomLevelFunc(t *testing.T) {
	levelAndElementCount := GenLevelAndLevelSize(10000000)
	for i, u := range levelAndElementCount {
		fmt.Printf("level: %d, count: %d\n", i, u)
	}
}

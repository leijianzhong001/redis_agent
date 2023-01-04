package rdb

import (
	"bufio"
	"fmt"
	"github.com/leijianzhong001/redis_agent/internal/reader"
	"os"
	"testing"
)

func TestBioRead(t *testing.T) {
	file, err := os.OpenFile("E:\\学习\\学习笔记\\studynote\\redis\\redis缓存淘汰策略.md", os.O_RDONLY, 0666)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buf := make([]byte, 100)
	num, err := reader.Read(buf)
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Printf("第一次读取%d个字节，内容为:%s\n", num, string(buf))

	num, err = reader.Read(buf)
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Printf("第二次读取%d个字节，内容为:%s\n", num, string(buf))
}

func TestReadRDB(t *testing.T) {
	reader := reader.NewRDBReader("F:\\202212\\dump.rdb")
	ch := reader.StartRead()
	for entry := range ch {
		fmt.Println(entry.Argv)
	}
}

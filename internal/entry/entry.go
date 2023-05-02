package entry

import "fmt"

type Entry struct {
	Id     uint64
	IsBase bool //  whether the command is decoded from dump.rdb file
	DbId   int
	Argv   []string

	Key      string // 当前key
	Overhead uint64 // 当前key的内存开销,单位是字节
}

func NewEntry() *Entry {
	e := new(Entry)
	return e
}

func (e *Entry) ToString() string {
	return fmt.Sprintf("%v", e.Argv)
}

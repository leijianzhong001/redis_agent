package types

import (
	"github.com/leijianzhong001/redis_agent/internal/log"
	"github.com/leijianzhong001/redis_agent/internal/rdb/structure"
	"io"
)

type ModuleObject struct {
}

func (o *ModuleObject) LoadFromBuffer(rd io.Reader, key string, typeByte byte) {
	if typeByte == rdbTypeModule {
		log.Panicf("module type with version 1 is not supported, key=[%s]", key)
	}
	moduleId := structure.ReadLength(rd)
	moduleName := moduleTypeNameByID(moduleId)
	opcode := structure.ReadByte(rd)
	for opcode != rdbModuleOpcodeEOF {
		switch opcode {
		case rdbModuleOpcodeSINT:
		case rdbModuleOpcodeUINT:
			structure.ReadLength(rd)
		case rdbModuleOpcodeFLOAT:
			structure.ReadFloat(rd)
		case rdbModuleOpcodeDOUBLE:
			structure.ReadDouble(rd)
		case rdbModuleOpcodeSTRING:
			structure.ReadString(rd)
		default:
			log.Panicf("unknown module opcode=[%d], module name=[%s]", opcode, moduleName)
		}
		opcode = structure.ReadByte(rd)
	}
}

func (o *ModuleObject) Rewrite() []RedisCmd {
	log.Panicf("module Rewrite not implemented")
	return nil
}

func (o *ModuleObject) MemOverhead() uint64 {
	return 0
}

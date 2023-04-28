package types

import (
	"github.com/leijianzhong001/redis_agent/internal/log"
	"github.com/leijianzhong001/redis_agent/internal/rdb/structure"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"io"
)

// quicklist node container formats
const (
	quicklistNodeContainerPlain  = 1 // QUICKLIST_NODE_CONTAINER_PLAIN
	quicklistNodeContainerPacked = 2 // QUICKLIST_NODE_CONTAINER_PACKED
)

type ListObject struct {
	key      string
	elements []string
}

func (o *ListObject) LoadFromBuffer(rd io.Reader, key string, typeByte byte) {
	o.key = key
	switch typeByte {
	case rdbTypeList:
		o.readList(rd)
	case rdbTypeListZiplist:
		o.elements = structure.ReadZipList(rd)
	case rdbTypeListQuicklist:
		o.readQuickList(rd)
	case rdbTypeListQuicklist2:
		o.readQuickList2(rd)
	default:
		log.Panicf("unknown list type %d", typeByte)
	}
}

func (o *ListObject) Rewrite() []RedisCmd {
	cmds := make([]RedisCmd, len(o.elements))
	for inx, ele := range o.elements {
		cmd := RedisCmd{"rpush", o.key, ele}
		cmds[inx] = cmd
	}
	return cmds
}

func (o *ListObject) readList(rd io.Reader) {
	size := int(structure.ReadLength(rd))
	for i := 0; i < size; i++ {
		ele := structure.ReadString(rd)
		o.elements = append(o.elements, ele)
	}
}

func (o *ListObject) readQuickList(rd io.Reader) {
	// 这个是节点个数，不是元素个数
	size := int(structure.ReadLength(rd))
	for i := 0; i < size; i++ {
		ziplistElements := structure.ReadZipList(rd)
		o.elements = append(o.elements, ziplistElements...)
	}
}

func (o *ListObject) readQuickList2(rd io.Reader) {
	size := int(structure.ReadLength(rd))
	for i := 0; i < size; i++ {
		container := structure.ReadLength(rd)
		if container == quicklistNodeContainerPlain {
			ele := structure.ReadString(rd)
			o.elements = append(o.elements, ele)
		} else if container == quicklistNodeContainerPacked {
			listpackElements := structure.ReadListpack(rd)
			o.elements = append(o.elements, listpackElements...)
		} else {
			log.Panicf("unknown quicklist container %d", container)
		}
	}
}

const maxZipListSize = uint64(8192)

// MemOverhead 计算当前key加载到redis中以后的内存开销
// 一个`quicklist`存储结构最终会产生以下几个消耗内存的结构(相关代码可查阅`t_hash.c`中的`hashTypeLookupWriteOrCreate`函数)：
//		- 1个`dictEntry`结构，24字节，负责保存当前的哈希对象；
//		- 1个`SDS`结构，用作`key`字符串，占`4~18`个字节；
//		- 1个`redisObject`结构，`16`字节，其指针指向当前`key`下属的`quicklist`结构；
//		- 1个`quicklist`结构，40字节，负责保存哈希对象的键值对；
//		- `quicklist.len`个`quicklistNode`结构，每个`quicklistNode`占用32字节,总长度为`quicklist.len * 32`
//		- `quicklistNode`中的`ziplist`的长度按照`8kb`算，所以`ziplist`的总长度为`quicklist.len * 8192`
// 按照默认的配置，redis的所有节点不压缩，且每个ziplist的最大长度为`8KB`。而`quicklist`的插入方式是首先判断当前头节点(或尾节点)插入新元素以后是否小于8KB，如果小于，则将当前元素插入到当前头节点(或尾节点)的`ziplist`中，否则创建一个新的`quicklistNode`
// 所以我们在计算`quicklistNode`中的`ziplist`长度时，可以简单的按照`8KB`来算
// `ziplist`不是具体的结构体，所以没有下级结构
// 单个key的内存消耗 = `dictEntry`结构大小 + key_SDS大小 + redisObject大小 + `quicklist`结构 + quicklist.len * 32 + quicklist.len * 8192
func (o *ListObject) MemOverhead() uint64 {
	// `dictEntry`结构大小 + key_SDS大小 + redisObject大小 + `quicklist`结构
	topLevelObjOverhead := utils.DictEntryOverhead() + utils.SdsOverhead(o.key) + utils.RedisObjOverhead() + utils.QuicklistOverhead()

	var dataOverhead uint64
	var previousEntryLength uint64
	// 当前ziplist长度
	currentZipListSize := utils.ZiplistOverhead()
	// 当前quicklist中 quickListNode的数量
	quickListNodeCount := uint64(0)
	// quicklist.len * 32 + quicklist.len * 8192
	for _, element := range o.elements {
		// ziplist结构 = ziplist header + ziplist end + 所有entry
		// entry结构 = previous_entry_length + encoding + len(content)
		// 当前元素所在的entry的长度
		currentEntrySize := utils.ZlentryOverhead(previousEntryLength, element)
		if currentZipListSize+currentEntrySize > maxZipListSize {
			// 说明当前ziplist已经满了，此时将quickListNode数量 + 1
			quickListNodeCount++
			// 累加ziplist长度，到这个地方应用jmalloc规则分配一次内存
			dataOverhead += utils.MallocOverhead(currentZipListSize)
			// 满了以后重置当前zipList长度，开始下一个zipList累加
			currentZipListSize = utils.ZiplistOverhead()
		} else {
			// 当前ziplist没满，继续累加
			currentZipListSize += currentEntrySize
		}

		// 上一个zlentry的长度
		previousEntryLength = uint64(len(element))

	}

	return topLevelObjOverhead + utils.QuicklistNodeOverhead()*quickListNodeCount + dataOverhead
}

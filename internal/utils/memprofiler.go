// Copyright 2017 XUEQIU.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"math/rand"
	"sort"
	"strconv"
)

var (
	skiplistMaxLevel    = 32
	skiplistP           = 0.25
	redisSharedInterges = int64(10000)
	longSize            = uint64(8)
	pointerSize         = uint64(8)
	jemallocSizeClasses = []uint64{
		8,
		16, 32, 48, 64, 80, 96, 112, 128, // 16
		80, 96, 112, 128, // 16
		160, 192, 224, 256, // 32
		320, 384, 448, 512, // 64
		640, 768, 896, 1024, // 128
		1280, 1536, 1792, 2048, // 256
		2560, 3072, 3584, 4096, // 512
		5120, 6144, 7168, 8192, // 1024 => 1KB
		10240, 12288, 14336, 16384, // 2048
		20480, 24576, 28672, 32768, // 4096
		40960, 49152, 57344, 65536, // 8192
		81920, 98304, 114688, 131072, // 16384
		163840, 196608, 229376, 262144, // 32768
		327680, 393216, 458752, 524288, // 65536
		655360, 786432, 917504, 1048576, // 131072
		1310720, 1572864, 1835008, 2097152, // 262144
		2621440, 3145728, 3670016, 4194304, // 524288
		5242880, 6291456, 7340032, 8388608, // 1048576 ==> 1MB
		10485760, 12582912, 14680064, 16777216, // 2097152
		20971520, 25165824, 29360128, 33554432, // 4194304
		41943040, 50331648, 58720256, 67108864, // 8388608
		83886080, 100663296, 117440512, 134217728, // 16777216
		167772160, 201326592, 234881024, 268435456, // 33554432
		335544320, 402653184, 469762048, 536870912, // 67108864
		671088640, 805306368, 939524096, 1073741824, // 134217728
		1342177280, 1610612736, 1879048192, 2147483648, // 268435456
		2684354560, 3221225472, 3758096384, 4294967296, // 536870912
		5368709120, 6442450944, 7516192768, 8589934592, // 1073741824 ==> 1GB
		10737418240, 12884901888, 15032385536, 17179869184, // 2147483648
		21474836480, 25769803776, 30064771072, 34359738368, // 4294967296
		42949672960, 51539607552, 60129542144, 68719476736, // 8589934592
		85899345920, 103079215104, 120259084288, 137438953472, // 17179869184
		171798691840, 206158430208, 240518168576, 274877906944, // 34359738368
		343597383680, 412316860416, 481036337152, 549755813888, // 68719476736
		687194767360, 824633720832, 962072674304, 1099511627776, // 137438953472
		1374389534720, 1649267441664, 1924145348608, 2199023255552, // 274877906944
		2748779069440, 3298534883328, 3848290697216, 4398046511104, // 549755813888
		5497558138880, 6597069766656, 7696581394432, 8796093022208, // 1099511627776 => 1TB
		10995116277760, 13194139533312, 15393162788864, 17592186044416, // 2199023255552
	}
)

// MemProfiler get memory use for all kinds of data stuct
type MemProfiler struct{}

// MallocOverhead used memory
func MallocOverhead(size uint64) uint64 {
	idx := sort.Search(len(jemallocSizeClasses),
		func(i int) bool { return jemallocSizeClasses[i] >= size })
	if idx < len(jemallocSizeClasses) {
		return jemallocSizeClasses[idx]
	}
	return size
}

// TopLevelObjOverhead get memory use of a top level object
// Each top level object is an entry in a dictionary, and so we have to include
// the overhead of a dictionary entry
func (m *MemProfiler) TopLevelObjOverhead(key []byte, expiry int64) uint64 {
	return DictEntryOverhead() + SdsOverhead(string(key)) + RedisObjOverhead() + KeyExpiryOverhead(expiry)
}

// todo 修正 过期时间long long已经计算在dictEntry的开销中了，不需要再次计算。 倒是该附带产生的rehash开销没有计算
// KeyExpiryOverhead get memory useage of a key expiry
// Key expiry is stored in a hashtable, so we have to pay for the cost of a hashtable entry
// The timestamp itself is stored as an int64, which is a 8 bytes
func KeyExpiryOverhead(expiry int64) uint64 {
	//If there is no expiry, there isn't any overhead
	if expiry <= 0 {
		return 0
	}
	return DictEntryOverhead() + 8
}

func (m *MemProfiler) SizeofStreamRadixTree(numElements uint64) uint64 {
	numNodes := uint64(float64(numElements) * 2.5)
	return 16*numElements + numNodes*4 + numNodes*30*8
}

func (m *MemProfiler) StreamOverhead() uint64 {
	return 2*pointerSize + 8 + 16 + // stream struct
		pointerSize + 8*2 // rax struct
}

func (m *MemProfiler) StreamConsumer(name []byte) uint64 {
	return pointerSize*2 + 8 + SdsOverhead(string(name))
}

func (m *MemProfiler) StreamCG() uint64 {
	return pointerSize*2 + 16
}

func (m *MemProfiler) StreamNACK(length uint64) uint64 {
	return length * (pointerSize + 8 + 8)
}

// LinkedlistOverhead get memory use of a linked list
// See https://github.com/antirez/redis/blob/unstable/src/adlist.h
// A list has 5 pointers + an unsigned long
func (m *MemProfiler) LinkedlistOverhead() uint64 {
	return longSize + 5*pointerSize
}

// LinkedListEntryOverhead get memory use of a linked list entry
// See https://github.com/antirez/redis/blob/unstable/src/adlist.h
// A node has 3 pointers
func (m *MemProfiler) LinkedListEntryOverhead() uint64 {
	return 3 * pointerSize
}

// SkiplistOverhead get memory use of a skiplist
func (m *MemProfiler) SkiplistOverhead(size uint64) uint64 {
	return 2*pointerSize + DictOverhead() + (2*pointerSize + 16)
}

// SkiplistEntryOverhead get memory use of a skiplist entry
func (m *MemProfiler) SkiplistEntryOverhead() uint64 {
	return DictEntryOverhead() + 2*pointerSize + 8 + (pointerSize+8)*zsetRandLevel()
}

func (m *MemProfiler) ZiplistHeaderOverhead() uint64 {
	return 4 + 4 + 2 + 1
}

func (m *MemProfiler) ZiplistEntryOverhead(value []byte) uint64 {
	header := 0
	size := 0

	if n, err := strconv.ParseInt(string(value), 10, 64); err == nil {
		header = 1
		switch {
		case n < 12:
			size = 0
		case n < 256:
			size = 1
		case n < 65536:
			size = 2
		case n < 16777216:
			size = 3
		case n < 4294967296:
			size = 4
		default:
			size = 8
		}
	} else {
		size = len(value)
		if size <= 63 {
			header = 1
		} else if size <= 16383 {
			header = 2
		} else {
			header = 5

			if size >= 254 {
				header += 5
			}
		}
	}

	return uint64(header + size)
}

// todo 修正 过期时间long long已经计算在dictEntry的开销中了，不需要再次计算。 倒是该附带产生的rehash开销没有计算
// KeyExpiryOverhead get memory useage of a key expiry
// Key expiry is stored in a hashtable, so we have to pay for the cost of a hashtable entry
// The timestamp itself is stored as an int64, which is a 8 bytes
func (m *MemProfiler) KeyExpiryOverhead(expiry int64) uint64 {
	//If there is no expiry, there isn't any overhead
	if expiry <= 0 {
		return 0
	}
	return DictEntryOverhead() + 8
}

// ElemLen get length of a element
func (m *MemProfiler) ElemLen(element []byte) uint64 {
	MaxInt64 := int64(1<<63 - 1)
	MinInt64 := int64(-1 << 63)
	if num, err := strconv.ParseInt(string(element), 10, 64); err == nil {
		if num < MinInt64 || num > MaxInt64 {
			return 16
		}
		return 8
	}
	return uint64(len(element))
}

func zsetRandLevel() uint64 {
	level := 1
	rint := rand.Intn(0xFFFF)
	for rint < int(0xFFFF*1/4) { //skiplistP
		level++
		rint = rand.Intn(0xFFFF)
	}
	if level < skiplistMaxLevel {
		return uint64(level)
	}
	return uint64(skiplistMaxLevel)
}

const (
	// 这个长度已经加上了结尾的\0
	sizeOfSdshdr8  = 4
	sizeOfSdshdr16 = 6
	sizeOfSdshdr32 = 10
	sizeOfSdshdr64 = 18
)

// SdsOverhead 一个sds结构的开销
// sds.c/_sdsnewlen 函数
func SdsOverhead(val string) uint64 {
	size := len(val)
	if size < 256 {
		// sdshdr8
		return MallocOverhead(uint64(sizeOfSdshdr8 + size))
	} else if size < 65536 {
		// sdshdr16
		return MallocOverhead(uint64(sizeOfSdshdr16 + size))
	} else if size < 4294967296 {
		// sdshdr32
		return MallocOverhead(uint64(sizeOfSdshdr32 + size))
	} else {
		// sdshdr64
		return MallocOverhead(uint64(sizeOfSdshdr64 + size))
	}
}

const (
	intEncodingMaxLen    = 20
	embStrEncodingMaxLen = 44
)

// StringValueOverhead get memory use of a string
// https://github.com/antirez/redis/blob/unstable/src/sds.h
func StringValueOverhead(val string) uint64 {
	// 字符串长度
	stringLen := uint64(len(val))
	// 是否可以转为数字
	_, err := strconv.ParseInt(val, 10, 64)
	// 长度小于20并且可以转为long long类型。 在c语言中，long long类型占用8个字节，取值范围是-9223372036854775808~9223372036854775807，因此最多能保存长度为19的字符串转换之后的数值，再加上负号位数，一共20位
	if err == nil && stringLen <= intEncodingMaxLen {
		// 这里规定 maxmemory 必定设置， 所以不可以共享对象,不进行判断
		// 小于20位，并且是数字，则位OBJ_ENCODING_INT编码， 直接存放到ptr指针处，所以是8。 但是这里的8会算在redisObject结构体中，所以实际上sdsSize是0
		return RedisObjOverhead() + 0
	}

	// OBJ_ENCODING_RAW 和 OBJ_ENCODING_EMBSTR的区别在于其内存分配操作进行了两次，一次给redisObject, 一次给SDS, 但二者占用的内存大小是一致的
	sdsSize := SdsOverhead(val)
	return RedisObjOverhead() + sdsSize
}

// DictEntryOverhead get memory use of hashtable entry
// See  https://github.com/antirez/redis/blob/unstable/src/dict.h
// Each dictEntry has 3 pointers
// typedef struct dictEntry {
//     void *key;
//     union {
//         void *val;
//         uint64_t u64;
//         int64_t s64;
//         double d;
//     } v;
//     struct dictEntry *next;
// } dictEntry;
func DictEntryOverhead() uint64 {
	return MallocOverhead(pointerSize + pointerSize + pointerSize)
}

const LRU_BITS = 3 // 24 bit

// RedisObjOverhead redisObject结构体开销
// typedef struct redisobject {
//     unsigned type:4;
//     unsigned encoding:4;
//     unsigned lru:lru_bits; /* lru time (relative to server.lruclock) */
//     int refcount;
//     void *ptr;
// } robj;
func RedisObjOverhead() uint64 {
	return MallocOverhead(1 + LRU_BITS + 4 + pointerSize)
}

// DictOverhead get memory use of a hashtable
// See  https://github.com/antirez/redis/blob/unstable/src/dict.h
// See the structures dict and dictht
// 2 * (3 unsigned longs + 1 pointer) + int + long + 2 pointers
//
// Additionally, see **table in dictht
// The length of the table is the next power of 2
// When the hashtable is rehashing, another instance of **table is created
// Due to the possibility of rehashing during loading, we calculate the worse
// case in which both tables are allocated, and so multiply
// the size of **table by 1.5
// typedef struct dict {
//    dictType *type;      // 字典类型                                                                          8
//    void *privdata;      // 私有数据                                                                          8
//    dictht ht[2];        // 哈希表数组                                                                         64
//    long rehashidx;      // rehash索引，代表下一次执行扩容单步操作要迁移的ht[0]hash表数组索引，当不进行rehash时，值为-1 8
//    int iterators;       // 当前该字典迭代器个数,迭代器用于遍历字典键值对                                           4
//} dict;
func DictOverhead() uint64 {
	return MallocOverhead(pointerSize + pointerSize + DictHtOverhead()*2 + pointerSize + 4)
}

// DictHtOverhead 返回dictht结构体占用的内存大小
//typedef struct dictht {
//    dictEntry **table;        // 哈希表节点数组                           8
//    unsigned long size;       // 哈希表大小                              8
//    unsigned long sizemask;   // 哈希表大小掩码,用于计算索引值,等于size-1    8
//    unsigned long used;       // 该哈希表已有节点的数量                    8
//} dictht;
func DictHtOverhead() uint64 {
	return MallocOverhead(pointerSize * 4)
}

// FieldBucketOverhead 计算rehash过程中的额外开销
func FieldBucketOverhead(size uint64) uint64 {
	return MallocOverhead(NextPower(size)*pointerSize) + MallocOverhead(NextPower(size)/2*pointerSize)
}

// QuicklistOverhead 计算Quicklist结构体开销
///* 快速列表是一个描述快速列表的 40 字节结构（在 64 位系统上）。
// * 'count' 是条目总数。
// * 'len' 是快速列表节点的数量。
// * 'compress' 如果禁用压缩，则为 0，否则它是在快速列表末尾保持未压缩的快速列表节点数。
// * 'fill' 是用户请求的（或默认）填充因子。
// * 'bookmakrs 是 realloc 这个结构使用的可选功能，这样它们在不使用时就不会消耗内存。
// */
//typedef struct quicklist {
//    quicklistNode *head; // 8 byte
//    quicklistNode *tail; // 8 byte
//    unsigned long count;        /* total count of all entries in all ziplists */ // 8 byte
//    unsigned long len;          /* number of quicklistNodes */ // 8 byte
//    int fill : QL_FILL_BITS;              /* fill factor for individual nodes */ // 16 bit ==> 2 byte
//    unsigned int compress : QL_COMP_BITS; /* depth of end nodes not to compress;0=off */ // 16 bit ==> 2 byte
//    unsigned int bookmark_count: QL_BM_BITS; // 4 bit
//    quicklistBookmark bookmarks[]; // 0
//} quicklist;
func QuicklistOverhead() uint64 {
	// 多出来的3是内存对齐的部分，最后是40个字节
	return MallocOverhead(pointerSize + pointerSize + longSize + longSize + 2 + 2 + 1 + 3)
}

// QuicklistNodeOverhead 计算QuicklistNode开销
// /* quicklistNode 是一个 32 字节的结构，用于描述快速列表的 ziplist。我们使用位字段(bit fields)将 quicklistNode 保持在 32 字节。
// * count：16 位，最大 65536（最大 ZL字节数为 65K，因此最大计数实际上< 32K）。
// * encoding：2 位，RAW=1，LZF=2。
// * container：2 bits, NONE=1, ZIPLIST=2
// * recompress：1 位，布尔值，如果节点临时解压缩以供使用，则为 true。
// * attempted_compress：1 位，布尔值，用于在测试期间进行验证。
// * extra：10位，免费供将来使用; 填充 32 位的其余部分
// */
//typedef struct quicklistNode {
//    struct quicklistNode *prev; // 8 byte
//    struct quicklistNode *next; // 8 byte
//    unsigned char *zl; // 8 byte
//    unsigned int sz;             /* ziplist size in bytes */ // 4 byte
//    unsigned int count : 16;     /* count of items in ziplist */ // 16 bit ==> 2 byte
//    unsigned int encoding : 2;   /* RAW==1 or LZF==2 */ // 2 bit
//    unsigned int container : 2;  /* NONE==1 or ZIPLIST==2 */ // 2 bit
//    unsigned int recompress : 1; /* was this node previous compressed? */ // 1 bit
//    unsigned int attempted_compress : 1; /* node can't compress; too small */ // 1 bit
//    unsigned int extra : 10; /* more bits to steal for future usage */ // 10 bit
//} quicklistNode;
func QuicklistNodeOverhead() uint64 {
	// 32个字节
	return MallocOverhead(pointerSize + pointerSize + pointerSize + 4 + 2 + 2)
}

// ZiplistOverhead 结构开销
// 压缩列表的结构 ：<zlbytes> <zltail> <zllen> <entry> <entry> ... <entry> <zlend>
//		- `zlbytes`  4字节，记录整个压缩列表占用的内存字节数。
//		- `zltail`   4字节，记录压缩列表表尾节点的位置。
//		- `zllen` 2字节，记录压缩列表节点个数。
//		- `zlentry`  列表节点，长度不定，由内容决定。
//		- `zlend` 1字节，0xFF 标记压缩的结束。
// 见ziplist.c#ziplistNew函数
func ZiplistOverhead() uint64 {
	zipListHeaderSize := 4 + 4 + 2
	zipListEndSize := 1
	// ziplist的内存分配会在每次插入元素时重新分配，最差的情况可能会导致整个ziplist连带数据重新分配一次内存。 因此不在这里应用jmalloc规则，而是当ziplist满了以后应用一次
	return uint64(zipListHeaderSize + zipListEndSize)
}

// ZlentryOverhead 结构开销
// zlentry结构如下路所示：
// * | --------------------- |--------- | ------- |
// * | previous_entry_length | encoding | content |
// * | --------------------- | -------- | ------- |
//- 1、`previous_entry_length`：**前一节点**的长度，占1个或5个字节。
//  	- 如果前一节点的**长度小于254字节**，则采用1个字节来保存这个长度值
//  	- 如果前一节点的**长度大于254字节**，则采用5个字节来保存这个长度值，第一个字节为`0xfe`，后四个字节才是真实长度数据
//- 2、`encoding`：编码属性，**记录`content`的数据类型**（字符串还是整数）以及长度，占用1个、2个或5个字节
//- 3、`contents`：负责保存节点的数据，可以是字符串或整数
// entry结构 = previous_entry_length + encoding + len(content) = 1 + 2 + len(content) = 3 + len(content)
// 参考 quicklist.c/_quicklistNodeAllowInsert
func ZlentryOverhead(previousEntryLength uint64, context string) uint64 {
	// 实际字符串内容的长度
	sz := len(context)

	// 存储前一个节点长度需要占用的字节数
	previousEntryLengthOverhead := 1
	if previousEntryLength >= 254 {
		previousEntryLengthOverhead = 5
	}

	/* size of forward offset
	 * ziplist的字符串编码方式 这里只考虑字符串，因为redis官方判断的时候好像也只考虑了字符串，因此如果实际值是数字的话，占用会被高估
	 *  | **编码**                                          | **编码长度**  | **字符串大小**            |
	 *  | ------------------------------------------------ | ------------ | ----------------------- |
	 *  | `|00pppppp|`                                     | **1 bytes**  | **<= 63 bytes**         |
	 *  | `|01pppppp|qqqqqqqq|`                            | **2 bytes**  | **<= 16383 bytes**      |
	 *  | `|10000000|qqqqqqqq|rrrrrrrr|ssssssss|tttttttt|` | **5 bytes**  | **<= 4294967295 bytes** |
	 * */
	encodingOverhead := 1
	if sz < 64 {
		encodingOverhead = 1
	} else if sz < 16384 {
		encodingOverhead = 2
	} else {
		encodingOverhead = 5
	}
	// 当我们新插入一个元素到ziplist中时， redis会重新分配【当前ziplist大小 + 插入的元素大小 + 后驱节点的prevlen属性长度要调整的大小】的内存，所以应用jemalloc规则时，应该在计算ziplist总开销时应用，而不是在这里
	return uint64(previousEntryLengthOverhead + encodingOverhead + sz)
}

// ZsetOverhead Zset结构开销
// typedef struct zset {
//    dict *dict; // 8
//    zskiplist *zsl; // 8
//} zset;
func ZsetOverhead() uint64 {
	return MallocOverhead(pointerSize + pointerSize)
}

// ZskiplistOverhead Zskiplist结构开销
// typedef struct zskiplist {
//    struct zskiplistNode *header, *tail; // 表头节点和表尾结点 16
//    unsigned long length; // 表中节点的数量 8
//    int level; // 表中节点的最大层数 4
//} zskiplist;
func ZskiplistOverhead() uint64 {
	// 多出来的4个字节是考虑到内存对齐到32
	return MallocOverhead(pointerSize + pointerSize + longSize + 4 + 4)
}

// ZskiplistNodeOverhead ZskiplistNode结构开销
// typedef struct zskiplistNode {
//    sds ele;  // 成员对象 8
//    double score; // 成员对象分值 8
//    struct zskiplistNode *backward; // 后退节点的指针，一个节点只有第一层（最底下那一层）有后退结点指针，所以zskiplist中的第一层是一个双向链表 8
//    // lever数组就是用于存储多级索引的，每一个元素就是一级索引
//    struct zskiplistLevel {
//        struct zskiplistNode *forward; // 本层前进节点指针 8
//        unsigned long span; // 本层的后继节点（forward前进指针指向的就是后继节点）跨越了多少个第一层节点，用于计算节点索引值。这个值其实就是距离头节点的偏移量，从0开始计算，每个节点递增1   8
//    } level[];
//} zskiplistNode;
func ZskiplistNodeOverhead(element string) uint64 {
	return MallocOverhead(pointerSize+longSize+pointerSize) + MallocOverhead(SdsOverhead(element))
}

// zskiplistLevelOverhead zskiplistLevel结构开销
// struct zskiplistLevel {
//        struct zskiplistNode *forward; // 本层前进节点指针 8
//        unsigned long span; // 本层的后继节点（forward前进指针指向的就是后继节点）跨越了多少个第一层节点，用于计算节点索引值。这个值其实就是距离头节点的偏移量，从0开始计算，每个节点递增1   8
//    }
func zskiplistLevelOverhead(element string) uint64 {
	return MallocOverhead(pointerSize + longSize)
}

func NextPower(size uint64) uint64 {
	power := uint64(1)
	for power <= size {
		power = power << 1
	}
	return power
}

const ZSKIPLIST_P float64 = 0.25
const ZSKIPLIST_MAXLEVEL = 32

func ZslRandomLevel() uint64 {
	// 注意这里初始值时1，所以必然会生成1层索引
	level := uint64(1)
	// random()函数返回0~rand_max的随机数，(random()&0xFFFF)小于等于0xFFFF。而 ZSKIPLIST_P为0.25， 则函数中while语句继续执行（增加层数）的概率为0.25,也就是从概率上将，k层节点的数量是k+1层节点的4倍。
	// 所以，redis的skiplist从概率上讲，相当于一颗4叉树
	r := rand.Intn(65535)
	rf := float64(r & 0xFFFF)
	for rf < (ZSKIPLIST_P * 0xFFFF) {
		level += 1
	}

	if level < ZSKIPLIST_MAXLEVEL {
		return level
	}
	return ZSKIPLIST_MAXLEVEL
}

// GenLevelAndLevelSize 生成索引层数以及每层元素数量分布情况
func GenLevelAndLevelSize(elementCount uint64) []uint64 {
	levelAndElementCount := make([]uint64, 0, 32)
	currentLevelElementCount := elementCount
	for true {
		if currentLevelElementCount >= 1 {
			levelAndElementCount = append(levelAndElementCount, currentLevelElementCount)
		} else {
			break
		}
		currentLevelElementCount = currentLevelElementCount / 4
	}
	return levelAndElementCount
}

package structure

import (
	"bufio"
	"encoding/binary"
	"github.com/leijianzhong001/redis_agent/internal/log"
	"io"
	"strconv"
	"strings"
)

const (
	zipStr06B = 0x00 // 0000 ZIP_STR_06B
	zipStr14B = 0x01 // 0001
	zipStr32B = 0x02 // 0010

	zipInt04B = 0x0f // high 4 bits of Int 04 encoding

	zipInt08B = 0xfe // 11111110
	zipInt16B = 0xc0 // 11000000
	zipInt24B = 0xf0 // 11110000
	zipInt32B = 0xd0 // 11010000
	zipInt64B = 0xe0 // 11100000
)

func ReadZipList(rd io.Reader) []string {
	rd = bufio.NewReader(strings.NewReader(ReadString(rd)))

	// The general layout of the ziplist is as follows:
	// <zlbytes> <zltail> <zllen> <entry> <entry> ... <entry> <zlend>
	_ = ReadUint32(rd) // zlbytes 整个压缩列表占用的字节数
	_ = ReadUint32(rd) // zltail 尾节点的起始地址距离压缩列表起始地址的偏移量

	size := int(ReadUint16(rd)) // zllen 压缩列表中的元素数量
	var elements []string
	if size == 65535 { // 2^16-1, we need to traverse the entire list to know how many items it holds.
		// 如果节点实际数量超出了最大值65534，则记录为65535，节点的真实数量只有完全遍历整个压缩列表才能知道
		for firstByte := ReadByte(rd); firstByte != 0xFE; firstByte = ReadByte(rd) {
			ele := readZipListEntry(rd, firstByte)
			elements = append(elements, ele)
		}
	} else {
		for i := 0; i < size; i++ {
			firstByte := ReadByte(rd)
			ele := readZipListEntry(rd, firstByte)
			elements = append(elements, ele)
		}
		if lastByte := ReadByte(rd); lastByte != 0xFF {
			log.Panicf("invalid zipList lastByte encoding: %d", lastByte)
		}
	}
	return elements
}

/*
 * So practically an entry is encoded in the following way:
 *
 * <prevlen from 0 to 253> <encoding> <entry>
 *
 * Or alternatively if the previous entry length is greater than 253 bytes
 * the following encoding is used:
 *
 * 0xFE <4 bytes unsigned little endian prevlen> <encoding> <entry>
 *
 * 1、`previous_entry_length`：**前一节点**的长度，占1个或5个字节。
 *  	- 如果前一节点的**长度小于254字节**，则采用1个字节来保存这个长度值
 *      - 如果前一节点的**长度大于254字节**，则采用5个字节来保存这个长度值， 第一个字节为`0xFE` ，后四个字节才是真实长度数据
 * 2、`encoding`：编码属性，**记录`content`的数据类型**（字符串还是整数）以及长度，占用1个、2个或5个字节
 * 3、`contents`：负责保存节点的数据，可以是字符串或整数
 */
func readZipListEntry(rd io.Reader, firstByte byte) string {
	// read prevlen
	if firstByte == 0xFE {
		_ = ReadUint32(rd) // read 4 bytes prevlen
	}

	// read encoding
	firstByte = ReadByte(rd)
	first2bits := (firstByte & 0xc0) >> 6 // first 2 bits of encoding
	switch first2bits {
	case zipStr06B:
		length := int(firstByte & 0x3f) // 0x3f = 00111111
		return string(ReadBytes(rd, length))
	case zipStr14B:
		secondByte := ReadByte(rd)
		length := (int(firstByte&0x3f) << 8) | int(secondByte)
		return string(ReadBytes(rd, length))
	case zipStr32B:
		lenBytes := ReadBytes(rd, 4)
		length := binary.BigEndian.Uint32(lenBytes)
		return string(ReadBytes(rd, int(length)))
	}
	switch firstByte {
	case zipInt08B:
		v := ReadInt8(rd)
		return strconv.FormatInt(int64(v), 10)
	case zipInt16B:
		v := ReadInt16(rd)
		return strconv.FormatInt(int64(v), 10)
	case zipInt24B:
		v := ReadInt24(rd)
		return strconv.FormatInt(int64(v), 10)
	case zipInt32B:
		v := ReadInt32(rd)
		return strconv.FormatInt(int64(v), 10)
	case zipInt64B:
		v := ReadInt64(rd)
		return strconv.FormatInt(v, 10)
	}
	if (firstByte >> 4) == zipInt04B {
		v := int64(firstByte & 0x0f) // 0x0f = 00001111
		v = v - 1                    // 1-13 -> 0-12
		if v < 0 || v > 12 {
			log.Panicf("invalid zipInt04B encoding: %d", v)
		}
		return strconv.FormatInt(v, 10)
	}
	log.Panicf("invalid encoding: %d", firstByte)
	return ""
}

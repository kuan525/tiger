package utils

import "hash/crc32"

// 也许是crc32的某个硬件指令集加速[有限子域矩阵问题]
func HashStr(key string) uint32 {
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

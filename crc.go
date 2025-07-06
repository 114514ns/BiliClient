package bili

import (
	"fmt"
	"strconv"
	"time"
)

const CRCPOLYNOMIAL = 0xEDB88320

var crctable [256]uint32

func createTable() {
	for i := 0; i < 256; i++ {
		crcreg := uint32(i)
		for j := 0; j < 8; j++ {
			if crcreg&1 != 0 {
				crcreg = CRCPOLYNOMIAL ^ (crcreg >> 1)
			} else {
				crcreg >>= 1
			}
		}
		crctable[i] = crcreg
	}
}

func crc32Custom(s string) uint32 {
	crcstart := uint32(0xFFFFFFFF)
	for i := 0; i < len(s); i++ {
		index := (crcstart ^ uint32(s[i])) & 0xFF
		crcstart = (crcstart >> 8) ^ crctable[index]
	}
	return crcstart
}

func crc32LastIndex(s string) int {
	crcstart := uint32(0xFFFFFFFF)
	for i := 0; i < len(s); i++ {
		index := (crcstart ^ uint32(s[i])) & 0xFF
		crcstart = (crcstart >> 8) ^ crctable[index]
	}
	return int((crcstart ^ 0xFFFFFFFF) & 0xFF)
}

func getCRCIndex(t int) int {
	for i := 0; i < 256; i++ {
		if int(crctable[i]>>24) == t {
			return i
		}
	}
	return -1
}

func deepCheck(i int, index []int) (bool, string) {
	str := ""
	hashcode := crc32Custom(strconv.Itoa(i))
	tc := int(hashcode&0xFF) ^ index[2]
	if tc < 48 || tc > 57 {
		return false, ""
	}
	str += string(rune(tc))

	hashcode = crctable[index[2]] ^ (hashcode >> 8)
	tc = int(hashcode&0xFF) ^ index[1]
	if tc < 48 || tc > 57 {
		return false, ""
	}
	str += string(rune(tc))

	hashcode = crctable[index[1]] ^ (hashcode >> 8)
	tc = int(hashcode&0xFF) ^ index[0]
	if tc < 48 || tc > 57 {
		return false, ""
	}
	str += string(rune(tc))

	return true, str
}

func mainFunc(input string) string {
	index := make([]int, 4)
	ht, _ := strconv.ParseUint(input, 16, 32)
	ht ^= 0xFFFFFFFF

	for i := 3; i >= 0; i-- {
		index[3-i] = getCRCIndex(int(ht >> (i * 8) & 0xFF))
		snum := crctable[index[3-i]]
		ht ^= uint64(snum >> ((3 - i) * 8))
	}

	for i := 0; i < 100000000; i++ {
		lastindex := crc32LastIndex(strconv.Itoa(i))
		if lastindex == index[3] {
			if ok, suffix := deepCheck(i, index); ok {
				return fmt.Sprintf("%d%s", i, suffix)
			}
		}
	}
	return "-1"
}

func launch() {

	createTable()
	start := time.Now()
	result := mainFunc("1b475f23")
	fmt.Println(result)
	elapsed := time.Since(start)
	fmt.Printf("耗时: %.2f 秒\n", elapsed.Seconds())
}

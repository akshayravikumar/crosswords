package cross

import (
	"fmt"
	"github.com/golang-collections/go-datastructures/bitarray"
	"strings"
)

const DistSize = BoardSize/4 + 2

//       0   1   2   3
// key: 000 100 010 001
func GetDistAtIndex(dist [DistSize]uint8, i uint8) uint8 {
	return ((0x3 << (2 * (i % 4))) & dist[i/4]) >> (2 * (i % 4))
}

func SetDistAtIndex(dist *[DistSize]uint8, i uint8, val uint8) {
	(*dist)[i/4] |= (val << (2 * (i % 4)))
}

func DistToString(dist [DistSize]uint8) string {
	var b strings.Builder
	for i := uint8(0); i < BoardSize; i++ {
		fmt.Fprintf(&b, "%2d ", GetDistAtIndex(dist, i))
	}
	return b.String()
}

func InitDist(row1, row2, row3 uint32) [DistSize]uint8 {
	var ret [DistSize]uint8
	arr1, arr2, arr3 := RowToArray(row1), RowToArray(row2), RowToArray(row3)
	for i := uint8(0); i < BoardSize; i++ {
		cur := 0
		if arr3[i] {
			cur = 3
		} else if arr2[i] {
			cur = 2
		} else if arr1[i] {
			cur = 1
		}
		SetDistAtIndex(&ret, uint8(i), uint8(cur))
	}
	return ret
}

func IsValidDist(dist [DistSize]uint8) bool {
	for i := uint8(0); i < BoardSize; i++ {
		curDist := GetDistAtIndex(dist, i)
		if curDist != 0 && curDist != 3 {
			return false
		}
	}
	return true
}

func GetPossibleNextRowsForDist(dist [DistSize]uint8) bitarray.BitArray {
	// don't want any zeros wherever any of the dists are > 0
	// get the key into oneonebitarraymap
	var key uint32 = 0
	for i := uint8(0); i < BoardSize; i++ {
		curDist := GetDistAtIndex(dist, i)
		if curDist == 1 || curDist == 2 {
			key = key | (1 << (BoardSize - 1 - i))
		}
	}
	return AvoidOneOneBitarrayMap[key]
}

func ApplyDist(dist [DistSize]uint8, row uint32) [DistSize]uint8 {
	// this is a lot easier than reach lmao
	var ret [DistSize]uint8
	rowArr := RowToArray(row)
	for i := uint8(0); i < BoardSize; i++ {
		if rowArr[i] {
			SetDistAtIndex(&ret, i, 3)
		} else {
			curDist := GetDistAtIndex(dist, i)
			if curDist == 0 {
				// 000 -> still 000
				SetDistAtIndex(&ret, i, 0)
			} else {
				// 001 -> 010 -> 100 -> 000
				SetDistAtIndex(&ret, i, curDist-1)
			}
		}
	}
	return ret
}

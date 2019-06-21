package cross

import (
	"github.com/golang-collections/go-datastructures/bitarray"
)

var AllRows []uint32
var PossMiddleRows []uint32
var MiddleRowMap map[uint32]bool
var RevMap map[uint32]uint32
var RowToReachMap map[uint32][HalfBoardSize]uint8
var PossFromMiddleMap = make(map[uint32]bitarray.BitArray)
var AvoidOneOneBitarrayMap = make([]bitarray.BitArray, MaxVal)
var AvoidOneZeroBitArrayMap = make(map[uint32]bitarray.BitArray)

func init() {
	// AllRows stores all rows without any 0-runs of
	// size 1 or 2
	for row := uint32(0); row < MaxVal; row++ {
		boolArray := RowToArray(row)
		curWhiteCount := 0
		works := true
		for i := 0; i < BoardSize; i++ {
			if !boolArray[i] {
				curWhiteCount += 1
				if i < (BoardSize-1) && boolArray[i+1] {
					if curWhiteCount < 3 {
						works = false
					}
				}
			} else {
				curWhiteCount = 0
			}
		}
		if !boolArray[BoardSize-1] && curWhiteCount < 3 {
			works = false
		}
		if works && row != MaxRow {
			AllRows = append(AllRows, row)
		}
	}

	MiddleRowMap = make(map[uint32]bool)
	MiddleRowMap[MaxRow] = true

	// Find all possible middle rows (palindromes)
	for _, row := range AllRows {
		arr := RowToArray(row)
		works := true
		for i := 0; i < BoardSize; i++ {
			if arr[i] != arr[BoardSize-1-i] {
				works = false
			}
		}
		// middle row cannot be all ones
		if works && row != MaxRow {
			PossMiddleRows = append(PossMiddleRows, row)
			MiddleRowMap[row] = true
		}
	}

	AllRows = PossMiddleRows

	// Store a map of each row to its mirror image
	RevMap = make(map[uint32]uint32)
	RevMap[MaxRow] = MaxRow
	for _, row := range AllRows {
		boolArray := RowToArray(row)
		var revArray [BoardSize]bool
		for i := uint32(0); i < BoardSize; i++ {
			revArray[i] = boolArray[BoardSize-1-i]
		}
		RevMap[row] = ArrayToRow(revArray)
	}

	// Get reach for each row
	RowToReachMap = make(map[uint32][HalfBoardSize]uint8)
	for _, row := range AllRows {
		// make reachMap
		islandCount := uint8(1)
		var reach [HalfBoardSize]uint8
		for i := uint8(0); i < BoardSize; i++ {
			boolArray := RowToArray(row)
			// if white, set island
			if !boolArray[i] {
				SetReachValueAtIndex(&reach, i, islandCount)
			} else {
				// black and prev was white
				if i > 0 && !boolArray[i-1] {
					islandCount += 1
				}
			}
		}
		RowToReachMap[row] = reach
	}

	// Get all rows that can be adjacent to a given middle row
	PossFromMiddleMap = make(map[uint32]bitarray.BitArray)
	for _, first := range PossMiddleRows {
		PossFromMiddleMap[first] = bitarray.NewSparseBitArray()
		for _, second := range AllRows {
			if RevMap[second] & ^first & second == 0 {
				PossFromMiddleMap[first].SetBit(second)
			}
		}
	}

	// for binary x, stores all vals that dont share a 1 in the
	// same position as x
	AvoidOneOneBitarrayMap = make([]bitarray.BitArray, MaxVal)
	for i := uint32(0); i < MaxVal; i++ {
		AvoidOneOneBitarrayMap[i] = bitarray.NewSparseBitArray()
		for _, int2 := range AllRows {
			if i&int2 == 0 {
				AvoidOneOneBitarrayMap[i].SetBit(int2)
			}
		}
	}

	// for binary x, stores all vals that dont have a 0 where
	// x has a 1
	AvoidOneZeroBitArrayMap = make(map[uint32]bitarray.BitArray)
	for _, first := range AllRows {
		AvoidOneZeroBitArrayMap[first] = bitarray.NewSparseBitArray()
		for _, second := range AllRows {
			if first & ^second > 0 {
				// there's a 10, which is not possible
			} else {
				AvoidOneZeroBitArrayMap[first].SetBit(second)
			}
		}
	}
}

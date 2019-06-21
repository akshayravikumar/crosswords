package main

import (
	. "./cross"
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

const Profile bool = true
const NumThreads int = 6
const LogFrequency uint64 = 100000000

func isConnected(board [HalfBoardSize]uint32, grid, visited *[BoardSize][BoardSize]bool) bool {
	for i := 0; i < BoardSize; i++ {
		if i < HalfBoardSize {
			grid[i] = RowToArray(board[i])
		} else {
			grid[i] = RowToArray(RevMap[board[BoardSize-1-i]])
		}
	}
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			visited[i][j] = false
		}
	}
	var dfs func(i, j int)
	dfs = func(i, j int) {
		if i < 0 || j < 0 || i >= BoardSize || j >= BoardSize {
			return
		}
		// it is black, do not consider
		if grid[i][j] {
			return
		}
		if visited[i][j] {
			return
		}
		visited[i][j] = true
		dfs(i, j+1)
		dfs(i, j-1)
		dfs(i+1, j)
		dfs(i-1, j)
	}
	done := false
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if !grid[i][j] && !visited[i][j] {
				if done {
					return false
				}
				dfs(i, j)
				done = true
			}
		}
	}
	return true
}

func main() {
	// bit 1 represents black, 0 represents white

	// want a map from AllRows to a bitarray of things that work for that pairing
	// 100 110 111 001 can't be a 1 in 100 or x10
	// 1 is black, 0 is white
	// 1, then other bit must
	if Profile {
		f, _ := os.Create("profile_stuff.prof")
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	getNextValues := func(a, b, c uint32) []uint32 {
		if c == MaxRow {
			return []uint32{MaxRow}
		}
		poss := AvoidOneOneBitarrayMap[(a & ^c)|(b & ^c)]
		if b & ^c == 0 && a & ^b & ^c == 0 {
			poss.SetBit(MaxRow)
		}
		return poss.ToNums()
	}

	getNextValuesForTopRow := func(a, b, c uint32) []uint32 {
		if c == MaxRow {
			return []uint32{MaxRow}
		}
		poss := AvoidOneOneBitarrayMap[(a & ^c)|(b & ^c)].And(AvoidOneZeroBitArrayMap[c])
		if b & ^c == 0 && a & ^b & ^c == 0 {
			poss.SetBit(MaxRow)
		}
		return poss.ToNums()
	}

	startTime := time.Now()
	var totalCount uint64 = 0
	search := func(wg *sync.WaitGroup, thread_id int) {
		defer wg.Done()
		var board [HalfBoardSize]uint32
		var grid [BoardSize][BoardSize]bool
		var visited [BoardSize][BoardSize]bool
		var recurse func(curIndex int)
		midCount := 0
		getIndex := func(board [HalfBoardSize]uint32, idx int) uint32 {
			if idx < HalfBoardSize {
				return board[idx]
			} else {
				return RevMap[board[BoardSize-1-idx]]
			}
		}

		recurse = func(curIndex int) {
			// curIndex starts at middle rows then goes down to 0
			switch {
			case curIndex == -1:
				good := isConnected(board, &grid, &visited)
				if good {
					atomic.AddUint64(&totalCount, 1)
					// fmt.Printf("board %v\n", BoardToString(board))
				}
			case curIndex == HalfBoardSize-1:
				for idx, middleRow := range PossMiddleRows {
					if idx%NumThreads == thread_id {
						fmt.Printf(
							"thread %02d start middle %0*b of %v time %v\n",
							thread_id,
							BoardSize,
							middleRow,
							len(PossMiddleRows),
							time.Since(startTime),
						)
						board[curIndex] = middleRow
						recurse(curIndex - 1)
					}
				}
			case curIndex == HalfBoardSize-2:
				possNextRows := PossFromMiddleMap[board[HalfBoardSize-1]].ToNums()
				for _, nextRow := range possNextRows {
					midCount += 1
					fmt.Printf(
						"    thread %02d start middle %0*b %v of %v time %v\n",
						thread_id,
						BoardSize,
						nextRow,
						midCount,
						len(possNextRows),
						time.Since(startTime),
					)
					board[curIndex] = nextRow
					recurse(curIndex - 1)
				}
			case curIndex < 2:
				possNextRows := getNextValuesForTopRow(
					getIndex(board, curIndex+3),
					getIndex(board, curIndex+2),
					getIndex(board, curIndex+1),
				)
				for _, nextRow := range possNextRows {
					board[curIndex] = nextRow
					recurse(curIndex - 1)
				}
			default:
				possNextRows := getNextValues(
					getIndex(board, curIndex+3),
					getIndex(board, curIndex+2),
					getIndex(board, curIndex+1),
				)
				for _, nextRow := range possNextRows {
					board[curIndex] = nextRow
					recurse(curIndex - 1)
				}
			}
		}
		recurse(HalfBoardSize - 1)
	}
	var wg sync.WaitGroup
	wg.Add(NumThreads)
	for thread_id := 0; thread_id < NumThreads; thread_id++ {
		go search(&wg, thread_id)
	}
	wg.Wait()
	fmt.Printf("DONE! total %v\n", totalCount)
}

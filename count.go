package main

import (
	. "./cross"
	"fmt"
	"math/big"
	"os"
	"runtime/pprof"
	"sync"
)

const Profile bool = true
const NumThreads int = 15
const LogFrequency uint64 = 10000000

// These help with debugging / special cases
// but they increase the number of states / time / memory
const IncludeBoardString = false
const IncludeBoardArr = false
const IncludeHasEdge = true

func main() {

	if Profile {
		f, _ := os.Create("profile.prof")
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// initialize a bunch of useful info for
	// this BoardSize
	fmt.Printf("Initializing stuff for board size %v...\n", BoardSize)

	type State struct {
		lastRow     uint32
		topReach    [ReachSize]uint8
		bottomReach [ReachSize]uint8
		dist        [DistSize]uint8
		boardString string
		hasEdge     bool
		isMirror    bool
	}

	var mutex sync.Mutex
	dp := make(map[State]*big.Int)

	// Initialize the DP with the middle row and the row adjacent to it
	for _, middleRow := range PossMiddleRows {
		possFromMiddleRowList := PossFromMiddleMap[middleRow].ToNums()
		middleReach := RowToReachMap[middleRow]
		for _, possFromMiddleRow := range possFromMiddleRowList {
			if possFromMiddleRow > RevMap[possFromMiddleRow] {
				continue
			}

			// the first 3 rows are RevMap[possFromMiddleRow], middleRow, possFromMiddleRow
			initialDist := InitDist(RevMap[possFromMiddleRow], middleRow, possFromMiddleRow)
			ok, initialTopReach, initialBottomReach :=
				ApplyReach(
					middleReach,
					middleReach,
					middleRow,
					possFromMiddleRow,
				)

			if !ok {
				continue
			}

			boardString := ""
			if IncludeBoardString {
				boardString =
					fmt.Sprintf(
						"%v%v%v",
						RowToString(possFromMiddleRow),
						RowToString(middleRow),
						RowToString(RevMap[possFromMiddleRow]),
					)
			}

			state := State{
				possFromMiddleRow,
				initialTopReach,
				initialBottomReach,
				initialDist,
				boardString,
				!IncludeHasEdge || RowHasEdge(middleRow) || RowHasEdge(possFromMiddleRow),
				MiddleRowMap[possFromMiddleRow],
			}
			_, exists := dp[state]
			if !exists {
				dp[state] = big.NewInt(0)
			}
			dp[state].Add(dp[state], big.NewInt(1))
		}
	}

	isConnected := func(state State) bool {
		return GetMaxComponentInReach(state.topReach) < 2 &&
			GetMaxComponentInReach(state.bottomReach) < 2
	}

	getPossibleNextRows := func(state State, index int) []uint32 {
		// After adding one all-black row, we can only add more
		// all-black rows
		if state.lastRow == MaxRow {
			return []uint32{MaxRow}
		} else {
			possRows := GetPossibleNextRowsForDist(state.dist)
			if index > HalfBoardSize-3 {
				possRows = possRows.And(AvoidOneZeroBitArrayMap[state.lastRow])
			}
			ret := possRows.ToNums()

			// If we're at a valid stopping state, we can start
			// adding all-black rows
			if isConnected(state) && IsValidDist(state.dist) {
				ret = append(ret, MaxRow)
			}
			return ret
		}
	}

	fmt.Println("Starting DP...")
	// Do the actual DP, parallelize some stuff too
	for curIndex := 2; curIndex < HalfBoardSize; curIndex++ {
		newDp := make(map[State]*big.Int)
		var keyList []State
		for key, _ := range dp {
			keyList = append(keyList, key)
		}

		run := func(wg *sync.WaitGroup, thread_id int) {
			defer wg.Done()
			for ind, state := range keyList {
				if ind%NumThreads != thread_id {
					continue
				}

				// iterate over all possible next rows
				for _, nextRow := range getPossibleNextRows(state, curIndex) {
					// Once we're done with the mirror-symmetric part, we only need to
					// store one copy of each mirror-symmetric board. So only take
					// one of every row, RevMap[row] pair
					if state.isMirror && nextRow > RevMap[nextRow] {
						continue
					}

					ok, nextTopReach, nextBottomReach :=
						ApplyReach(
							state.topReach,
							state.bottomReach,
							state.lastRow,
							nextRow,
						)
					if !ok {
						continue
					}

					nextDist := ApplyDist(state.dist, nextRow)
					boardString := ""
					if IncludeBoardString {
						boardString =
							fmt.Sprintf(
								"%v   %v\n%v\n%v   %v",
								ReachToString(nextTopReach),
								RowToString(nextRow),
								state.boardString,
								ReachToString(nextBottomReach),
								RowToString(RevMap[nextRow]),
							)
					}

					nextState := State{
						nextRow,
						nextTopReach,
						nextBottomReach,
						nextDist,
						boardString,
						!IncludeHasEdge || state.hasEdge || RowHasEdge(nextRow),
						state.isMirror && MiddleRowMap[nextRow],
					}
					mutex.Lock()
					_, exists := newDp[nextState]
					if !exists {
						newDp[nextState] = big.NewInt(0)
					}
					newDp[nextState].Add(newDp[nextState], dp[state])
					mutex.Unlock()
				}
			}
		}
		var wg sync.WaitGroup
		wg.Add(NumThreads)
		for thread_id := 0; thread_id < NumThreads; thread_id++ {
			go run(&wg, thread_id)
		}
		wg.Wait()
		dp = newDp
		fmt.Printf("Done with round %v, DP has %v states.\n", curIndex, len(dp))
	}

	ans := big.NewInt(0)
	for state, count := range dp {
		if !isConnected(state) {
			continue
		}
		if !state.hasEdge || state.lastRow == MaxRow {
			continue
		}
		if state.isMirror {
			ans.Add(ans, count)
		} else {
			ans.Add(ans, count)
			ans.Add(ans, count)
		}
		if IncludeBoardString {
			fmt.Printf("board %v\n\n", state.boardString)
		}
	}
	fmt.Printf("DONE! total %v\n", ans)
}

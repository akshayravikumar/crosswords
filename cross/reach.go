package cross

import (
	"fmt"
	"sort"
	"strings"
)

const ReachSize = HalfBoardSize

func GetReachValueAtIndex(reach [ReachSize]uint8, i uint8) uint8 {
	ret := (0xF << (4 * (i % 2)) & reach[i/2]) >> (4 * (i % 2))
	return ret
}

func ReachToArray(reach [ReachSize]uint8) [BoardSize]uint8 {
	var ret [BoardSize]uint8
	for i := uint8(0); i < BoardSize; i++ {
		ret[i] = GetReachValueAtIndex(reach, i)
	}
	return ret
}

func SetReachValueAtIndex(reach *[ReachSize]uint8, i uint8, val uint8) {
	(*reach)[i/2] |= ((val) << (4 * (i % 2)))
}

func ReachToString(reach [ReachSize]uint8) string {
	var b strings.Builder
	for i := uint8(0); i < BoardSize; i++ {
		fmt.Fprintf(&b, "%2d ", GetReachValueAtIndex(reach, i))
	}
	return b.String()
}

func GetMaxComponentInReach(reach [ReachSize]uint8) uint8 {
	ret := uint8(0)
	for i := uint8(0); i < BoardSize; i++ {
		val := GetReachValueAtIndex(reach, i)
		if val > ret {
			ret = val
		}
	}
	return ret
}

func ApplyReach(oldTopReach, oldBottomReach [ReachSize]uint8, oldTopRow, newTopRow uint32) (bool, [ReachSize]uint8, [ReachSize]uint8) {
	// special edge case for all-black row, assume it always works
	if newTopRow == MaxRow {
		return true, RowToReachMap[MaxRow], RowToReachMap[MaxRow]
	}

	oldTopRowArray := RowToArray(oldTopRow)
	oldBottomRowArray := RowToArray(RevMap[oldTopRow])

	newTopRowArray := RowToArray(newTopRow)
	newBottomRowArray := RowToArray(RevMap[newTopRow])

	oldTopReachArray := ReachToArray(oldTopReach)
	oldBottomReachArray := ReachToArray(oldBottomReach)
	adjMapTop, newIndexMapTop, connectedTop :=
		applyReachNoRenumbering(
			oldTopReachArray,
			oldTopRowArray,
			newTopRowArray,
		)
	adjMapBottom, newIndexMapBottom, connectedBottom :=
		applyReachNoRenumbering(
			oldBottomReachArray,
			oldBottomRowArray,
			newBottomRowArray,
		)

	// get a list of all keys in the old reach, and
	allKeys := make(map[uint8]bool)
	for _, key := range oldTopReachArray {
		allKeys[key] = true
	}
	for _, key := range oldBottomReachArray {
		allKeys[key] = true
	}
	var sortedKeys []uint8
	for key, _ := range allKeys {
		if key != 0 {
			sortedKeys = append(sortedKeys, key)
		}
	}
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	// If a reach isn't connected to the new row from either side,
	// return false because this means the grid is unconnected
	for _, key := range sortedKeys {
		if !connectedTop[key] && !connectedBottom[key] {
			var dummy [ReachSize]uint8
			return false, dummy, dummy
		}
	}

	// now, number the reaches in the new row
	// by enforcing an order in which we fill things, hopefully we prevent
	// having states that are essentially equivalent to each other
	oldReachToComponentMap := getNewReachValueComponents(adjMapTop, adjMapBottom, sortedKeys)

	currentComponent := uint8(0)
	newComponentToComponentMap := make(map[uint8]uint8)

	getFinalReach := func(row [BoardSize]bool, indexMap [BoardSize]uint8) [ReachSize]uint8 {
		var ret [ReachSize]uint8
		for i := uint8(0); i < BoardSize; i++ {
			// it's a white square, we care about its reach
			if !row[i] {
				// check if reachable from old row
				oldReachValue := indexMap[i]
				if oldReachValue > 0 {
					oldComponent := oldReachToComponentMap[oldReachValue]
					newReachValue, exists := newComponentToComponentMap[oldComponent]
					if exists {
						SetReachValueAtIndex(&ret, i, newReachValue)
					} else {
						currentComponent += 1
						newComponentToComponentMap[oldComponent] = currentComponent
						SetReachValueAtIndex(&ret, i, currentComponent)
					}
				} else {
					// this cell is not reachable from the previous row
					if i == 0 {
						// this must be component 1
						currentComponent += 1
						SetReachValueAtIndex(&ret, i, currentComponent)
					} else {
						prevReachValue := GetReachValueAtIndex(ret, i-1)
						if prevReachValue == 0 {
							// previous is black, increment component
							currentComponent += 1
							SetReachValueAtIndex(&ret, i, currentComponent)
						} else {
							// previous is white, set same component
							SetReachValueAtIndex(&ret, i, prevReachValue)
						}
					}
				}
			}
		}
		return ret
	}

	return true, getFinalReach(newTopRowArray, newIndexMapTop), getFinalReach(newBottomRowArray, newIndexMapBottom)
}

// private helpers

func applyReachNoRenumbering(
	oldReachArray [BoardSize]uint8,
	oldRowArray [BoardSize]bool,
	newRowArray [BoardSize]bool,
) (map[uint8]map[uint8]bool, [BoardSize]uint8, map[uint8]bool) {
	var grid [2][BoardSize]bool
	connected := make(map[uint8]bool)
	grid[0] = oldRowArray
	grid[1] = newRowArray

	// get a map of reach to indices it contains from old row
	oldReachValueToIndicesMap := make(map[uint8][]uint8)
	for i := uint8(0); i < BoardSize; i++ {
		// if reach is zero, it's a black square
		if oldReachArray[i] > 0 {
			cur := oldReachValueToIndicesMap[oldReachArray[i]]
			oldReachValueToIndicesMap[oldReachArray[i]] = append(cur, i)
		}
	}

	// lets do it
	var newRowIndexToOldReachValue [BoardSize]uint8
	oldReachValueAdjacencyMap := make(map[uint8]map[uint8]bool)

	for oldReachValue, indicesInOldReachValue := range oldReachValueToIndicesMap {
		var visited [2][BoardSize]bool
		oldReachValueAdjacencyMap[oldReachValue] = make(map[uint8]bool)

		// check if this component is adjacent to a white square
		// in the next row
		connectedToNewRow := false
		for _, ind := range indicesInOldReachValue {
			if !grid[1][ind] {
				connectedToNewRow = true
				break
			}
		}
		if connectedToNewRow {
			connected[oldReachValue] = true
		}
		// run a bfs from this component and mark everything reachable
		// from this component
		// while we're doing this, create an adjacency map of components
		// in the old row that can reach each other via the new row
		queue := indicesInOldReachValue
		count := 0
		for len(queue) > 0 {
			count += 1
			cur := queue[0]
			queue = queue[1:]
			row, index := cur/BoardSize, cur%BoardSize
			visited[row][index] = true

			existingOldReachValue := newRowIndexToOldReachValue[index]
			if row == 1 {
				if existingOldReachValue == 0 {
					// this is our first time visiting this
					newRowIndexToOldReachValue[index] = oldReachValue
				} else {
					// we've seen this before, so mark these reach values as
					// connected
					oldReachValueAdjacencyMap[oldReachValue][existingOldReachValue] = true
					oldReachValueAdjacencyMap[existingOldReachValue][oldReachValue] = true
				}
			}

			// add unvisited neighbors values to the queue
			// because i don't want to store tuples, we map
			// (i, j) --> BoardSize * i + j
			if !visited[1-row][index] && !grid[1-row][index] {
				queue = append(queue, BoardSize*(1-row)+index)
			}
			if index > 0 {
				if !grid[row][index-1] {
					if !visited[row][index-1] {
						queue = append(queue, cur-1)
					}
				}
			}
			if index < BoardSize-1 {
				if !visited[row][index+1] && !grid[row][index+1] {
					queue = append(queue, cur+1)
				}
			}
		}
	}
	return oldReachValueAdjacencyMap, newRowIndexToOldReachValue, connected
}

func getNewReachValueComponents(
	adjMapTop map[uint8]map[uint8]bool,
	adjMapBottom map[uint8]map[uint8]bool,
	sortedKeys []uint8,
) map[uint8]uint8 {
	// we have a list of which components in the old reach
	// are connected, now run a DFS and find our new
	// connected components
	reachToComponentMap := make(map[uint8]uint8)
	visited := make(map[uint8]bool)

	currentComponent := uint8(1)
	for _, key := range sortedKeys {
		if visited[key] {
			continue
		}
		queue := []uint8{key}
		visited[key] = true

		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			reachToComponentMap[cur] = currentComponent
			for neighbor, _ := range adjMapTop[cur] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
			for neighbor, _ := range adjMapBottom[cur] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
		currentComponent += 1
	}

	return reachToComponentMap
}

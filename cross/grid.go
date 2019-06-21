package cross

import (
	"fmt"
	"strings"
)

func RowToArray(row uint32) [BoardSize]bool {
	var ret [BoardSize]bool
	for i := uint32(0); i < BoardSize; i++ {
		ret[BoardSize-1-i] = (row & (1 << i)) > 0
	}
	return ret
}

func ArrayToRow(arr [BoardSize]bool) uint32 {
	var row uint32 = 0
	for i := uint32(0); i < BoardSize; i++ {
		if arr[i] {
			row = row | (1 << (BoardSize - i - 1))
		}
	}
	return row
}

func RowHasEdge(row uint32) bool {
	return ((row & (1 << 0)) == 0) || ((row & (1 << (BoardSize - 1))) == 0)
}

func RowToString(row uint32) string {
	return fmt.Sprintf("%0*b", BoardSize, row)
}

func SliceBoardToString(board []uint32) string {
	var b strings.Builder
	for _, row := range board {
		fmt.Fprintf(&b, "%v\n", RowToString(row))
	}
	return b.String()
}

func BoardToString(board [HalfBoardSize]uint32) string {
	var b strings.Builder
	for i := 0; i < HalfBoardSize; i++ {
		fmt.Fprintf(&b, "%0*b\n", BoardSize, board[i])
	}
	for i := HalfBoardSize - 2; i >= 0; i-- {
		fmt.Fprintf(&b, "%0*b\n", BoardSize, RevMap[board[i]])
	}
	return b.String()
}

const IsConnectedSize = BoardSize

var visited [IsConnectedSize][BoardSize]bool
var grid [IsConnectedSize][BoardSize]bool

func isConnected(board [IsConnectedSize]uint32) bool {
	for i := 0; i < IsConnectedSize; i++ {
		grid[i] = RowToArray(board[i])
	}
	for i := 0; i < IsConnectedSize; i++ {
		for j := 0; j < BoardSize; j++ {
			visited[i][j] = false
		}
	}
	var dfs func(i, j int)
	dfs = func(i, j int) {
		if i < 0 || j < 0 || i >= IsConnectedSize || j >= BoardSize {
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
	// dfs only from first row
	done := false
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if !grid[0][j] && !visited[0][j] {
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

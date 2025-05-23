package main

import (
	"fmt"
	"proj3-redesigned/engine"
	"proj3-redesigned/search"
)

func main() {
	board := engine.NewStandardBoard()
	printedBoard := board.StringBoard()

	fmt.Println(printedBoard)

	position := engine.NewStandardPosition()
	moves := position.Moves()
	fmt.Println(moves)

	move, eval := search.SearchBestMove(position, 5)
	fmt.Printf("best %s (eval %d)\n", move, eval)

}

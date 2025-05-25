package main

import (
	"fmt"
	"os"
	"proj3-redesigned/engine"
	"proj3-redesigned/search"
	"strconv"
	"time"
)

func main() {
	args := os.Args[1:]

	mode := args[0]                        // mode: 's': sequential, 'p': parallel, 'w': work-stealing
	numThreads, _ := strconv.Atoi(args[1]) // number of threads for parallel
	start_position := args[2]              // 'b' for beginning, 'f' for fischer
	depth, _ := strconv.Atoi(args[3])      // half-move plies

	var position engine.Position
	var board engine.Board
	const capacity = 128 // for work-stealing deque

	if start_position == "b" {
		position = engine.NewStandardPosition()
		board = engine.NewStandardBoard()
	} else {
		position = engine.NewFischerPosition()
		board = engine.NewFischerImmortalGameBoard()
	}

	stringBoard := board.StringBoard()
	fmt.Println(stringBoard)

	if mode == "p" {
		start_threads := time.Now()
		move_t, eval_t := search.SearchBestMoveParallelFixedThreadCount(position, depth, numThreads)
		elapsed_threads := time.Since(start_threads).Seconds()
		fmt.Printf("Best: %s (eval %d)\n", move_t, eval_t)
		fmt.Printf("P-TIME: %f\n", elapsed_threads)
	} else if mode == "w" {
		start_ws := time.Now()
		move_ws, eval_ws := search.SearchBestMoveParallelWorkStealing(position, depth, numThreads, capacity)
		elapsed_ws := time.Since(start_ws).Seconds()
		fmt.Printf("Best: %s (eval %d)\n", move_ws, eval_ws)
		fmt.Printf("WS-TIME: %f\n", elapsed_ws)
	} else { // sequential
		start_sequential := time.Now()
		move, eval := search.SearchBestMove(position, depth)
		elapsed_sequential := time.Since(start_sequential).Seconds()
		fmt.Printf("best %s (eval %d)\n", move, eval)
		fmt.Printf("S-TIME: %f\n", elapsed_sequential)
	}

}

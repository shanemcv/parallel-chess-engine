package main

import (
	"fmt"
	"proj3-redesigned/engine"
	"proj3-redesigned/search"
	"time"
)

func main() {
	//board := engine.NewStandardBoard()
	board := engine.NewFischerImmortalGameBoard()
	printedBoard := board.StringBoard()

	fmt.Println(printedBoard)

	//position := engine.NewStandardPosition()
	position := engine.NewFischerPosition()
	moves := position.Moves()
	fmt.Println(moves)

	/*
		start_sequential := time.Now()
		move, eval := search.SearchBestMove(position, 5)
		elapsed_sequential := time.Since(start_sequential).Seconds()
		fmt.Printf("best %s (eval %d)\n", move, eval)
		fmt.Printf("Sequential Time: %f\n", elapsed_sequential)
	*/

	start_parallel := time.Now()
	move_p, eval_p := search.SearchBestMoveParallel(position, 5)
	elapsed_parallel := time.Since(start_parallel).Seconds()
	fmt.Printf("Best: %s (Score: %d)\n", move_p, eval_p)
	fmt.Printf("Fully Parallel Time: %f\n", elapsed_parallel)

	start_threads := time.Now()
	numThreads := 4
	move_t, eval_t := search.SearchBestMoveParallelFixedThreadCount(position, 5, numThreads)
	elapsed_threads := time.Since(start_threads).Seconds()
	fmt.Printf("Best: %s (eval %d)\n", move_t, eval_t)
	fmt.Printf("Parallel with %d Threads Time: %f\n", numThreads, elapsed_threads)

	start_ws := time.Now()
	numThreads = 4
	const capacity = 128
	move_ws, eval_ws := search.SearchBestMoveParallelWorkStealing(position, 5, numThreads, capacity)
	elapsed_ws := time.Since(start_ws).Seconds()
	fmt.Printf("Best: %s (eval %d)\n", move_ws, eval_ws)
	fmt.Printf("Parallel with %d Threads and Work Stealing Time: %f\n", numThreads, elapsed_ws)

}

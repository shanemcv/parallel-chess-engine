package search

import (
	"fmt"
	"math"
	"math/rand"
	"proj3-redesigned/deque"
	"proj3-redesigned/engine"
	"sync"
)

/*
* MAIN SEARCHING FUNCTION --- SEQUENTIAL
* USING ALPHA-BETA PRUNING / NEGAMAX
* https://www.chessprogramming.org/Alpha-Beta
* depth is half-move 'plies', i.e. a depth of 6 is 6 half-moves and
* 3 full moves (white and black each move)
 */
func negamax(p engine.Position, depth, alpha, beta int) int {
	if depth == 0 {
		return p.Score
	}

	best := math.MinInt32
	for _, m := range p.Moves() {
		next := p.Move(m)
		score := -negamax(next, depth-1, -beta, -alpha) // recursive call
		if score > best {
			best = score
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			break
		}
	}
	if best == math.MinInt32 {
		return -engine.MateValue
	}
	return best
}

func SearchBestMove(start engine.Position, depth int) (bestMove engine.Move, bestScore int) {
	bestScore = math.MinInt32
	alpha := math.MinInt32
	beta := math.MaxInt32

	for _, m := range start.Moves() {
		next := start.Move(m)
		score := -negamax(next, depth-1, -beta, -alpha)
		if score > bestScore {
			bestScore = score
			bestMove = m
		}
		if bestScore > alpha {
			alpha = bestScore
		}
	}
	return bestMove, bestScore
}

/*
* MAIN SEARCHING FUNCTION --- FULLY PARALLEL
* (ONE THREAD PER POSSIBLE MOVE)
 */
func SearchBestMoveParallel(start engine.Position, depth int) (bestMove engine.Move, bestScore int) {
	bestScore = math.MinInt32
	alpha := math.MinInt32
	// beta := math.MaxInt32

	moves := start.Moves()
	var (
		mu sync.Mutex
		// cond = sync.NewCond(&mu)
		// completed int
		wg sync.WaitGroup
	)

	id := 0

	for _, m := range moves {
		wg.Add(1)
		id++
		currID := id
		move := m
		next := start.Move(m)
		go func(id int, move engine.Move) {
			defer wg.Done()
			score := SearchOneMoveTree(id, move, next, depth)
			mu.Lock()
			if score > bestScore {
				bestScore = score
				bestMove = move
			}
			if bestScore > alpha {
				alpha = bestScore
			}
			mu.Unlock()
		}(currID, move)
	}
	wg.Wait()
	return bestMove, bestScore
}

/*
* MAIN SEARCHING FUNCTION --- FIXED THREAD COUNT
* USES NUMBER OF THREADS SPECIFIED
* Threads will pick up additional possible moves.
 */
func SearchBestMoveParallelFixedThreadCount(start engine.Position, depth int, numThreads int) (bestMove engine.Move, bestScore int) {
	bestScore = math.MinInt32
	alpha := math.MinInt32
	// beta := math.MaxInt32

	moves := start.Moves()
	var (
		mu sync.Mutex
		// cond = sync.NewCond(&mu)
		// completed int
		wg sync.WaitGroup
	)

	MoveIdx := 0

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		currID := i + 1
		go func(id int) {
			defer wg.Done()
			for {
				mu.Lock()
				if MoveIdx >= len(moves) {
					mu.Unlock()
					return
				}
				move := moves[MoveIdx]
				MoveIdx++
				mu.Unlock()
				next := start.Move(move)
				score := SearchOneMoveTree(id, move, next, depth)
				mu.Lock()
				if score > bestScore {
					bestScore = score
					bestMove = move
				}
				if bestScore > alpha {
					alpha = bestScore
				}
				mu.Unlock()
			}
		}(currID)
	}

	wg.Wait()
	return bestMove, bestScore
}

/*
* MAIN SEARCHING FUNCTION --- WORK STEALING
* USES QUEUES FOR NUMBER OF THREADS SPECIFIED
* Threads will pick up additional possible moves from other queues.
 */
func SearchBestMoveParallelWorkStealing(start engine.Position, depth int, numThreads int, capacity int) (bestMove engine.Move, bestScore int) {
	bestScore = math.MinInt32
	alpha := math.MinInt32
	// beta := math.MaxInt32

	moves := start.Moves()
	var (
		mu sync.Mutex
		// cond = sync.NewCond(&mu)
		// completed int
		wg sync.WaitGroup
	)

	deques := make([]*deque.Deque, numThreads)
	for i := 0; i < numThreads; i++ {
		deques[i] = deque.NewDeque(capacity)
	}

	for i, m := range moves {
		move := m
		index := i % numThreads
		deques[index].PushBottom(func() {
			next := start.Move(move)
			score := SearchOneMoveTree(index+1, move, next, depth)
			mu.Lock()
			if score > bestScore {
				bestScore = score
				bestMove = move
			}
			if bestScore > alpha {
				alpha = bestScore
			}
			mu.Unlock()
		})
	}

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				task, success := deques[id].PopBottom() // own work
				if success {
					task()
					continue
				}
				stealFrom := rand.Intn(numThreads - 1) // steal work
				task, success = deques[stealFrom].PopTop()
				if success {
					task()
					continue
				}
				return
			}
		}(i)
	}

	wg.Wait()
	return bestMove, bestScore

}

func SearchOneMoveTree(id int, m engine.Move, movePos engine.Position, depth int) (score int) {
	alpha := math.MinInt32
	beta := math.MaxInt32
	score = -negamax(movePos, depth-1, -beta, -alpha)
	fmt.Printf("Thread ID %d for move %s, Score: %d\n", id, m, score)
	return score
}

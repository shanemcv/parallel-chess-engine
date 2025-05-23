package search

import (
	"math"
	"proj3-redesigned/engine"
)

/*
* MAIN SEARCHING CODE --- SEQUENTIAL
* USING ALPHA-BETA PRUNING / NEGAMAX
* https://www.chessprogramming.org/Alpha-Beta
* depth is half-move 'plies'
 */
func negamax(p engine.Position, depth, alpha, beta int) int {
	if depth == 0 {
		return p.Score
	}

	best := math.MinInt32
	for _, m := range p.Moves() { // this loop is most easily parallelized
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

func SearchBestMove(root engine.Position, depth int) (bestMove engine.Move, bestScore int) {
	bestScore = math.MinInt32
	alpha := math.MinInt32
	beta := math.MaxInt32

	for _, m := range root.Moves() {
		next := root.Move(m)
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

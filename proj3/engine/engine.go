package engine

// Backend chess engine code based on overview provided at
// https://zserge.com/posts/carnatus/

type Piece byte

func (p Piece) Value() int {
	pieceValues := make(map[Piece]int)
	pieceValues['P'] = 100
	pieceValues['N'] = 280
	pieceValues['B'] = 320
	pieceValues['R'] = 479
	pieceValues['Q'] = 929
	pieceValues['K'] = 60000 // should be value higher than any possible combination including all queen promotions
	return pieceValues[p]
}

func (p Piece) MySide() bool { return p.Value() > 0 }

func (p Piece) Flip() Piece {
	// changes piece color
	pieceFlips := make(map[Piece]Piece)
	pieceFlips['P'] = 'p'
	pieceFlips['p'] = 'P'
	pieceFlips['N'] = 'n'
	pieceFlips['n'] = 'N'
	pieceFlips['B'] = 'b'
	pieceFlips['b'] = 'B'
	pieceFlips['R'] = 'r'
	pieceFlips['r'] = 'R'
	pieceFlips['Q'] = 'q'
	pieceFlips['q'] = 'Q'
	pieceFlips['K'] = 'k'
	pieceFlips['k'] = 'K'
	pieceFlips[' '] = ' '
	pieceFlips['.'] = '.'
	return pieceFlips[p]
}

// Board uses 12x10 board for padding to evaluate possible moves
// where only the standard 8x8 chessboard is 'valid'
// the board is implemented as an array, so some move calculation
// is required to intuitively think of piece movements on this array
type Board [120]Piece

func (inBoard Board) Flip() (outBoard Board) {
	for i := 0; i < len(inBoard); i++ {
		outBoard[i] = inBoard[len(inBoard)-1-i].Flip()
	}
	return outBoard
}

func (board Board) StringBoard() (printedBoard string) {
	printedBoard = "\n"
	for row := 2; row < 10; row++ {
		for column := 1; column < 9; column++ {
			printedBoard = printedBoard + string(board[row*10+column])
		}
		printedBoard = printedBoard + "\n"
	}
	return printedBoard
}

func NewStandardBoard() Board {
	var b Board
	// padding
	for i := 0; i <= 20; i++ {
		b[i] = ' '
	}

	// Black pieces
	copy(b[21:29], []Piece{'r', 'n', 'b', 'q', 'k', 'b', 'n', 'r'})
	b[29] = ' '
	b[30] = ' '
	copy(b[31:39], []Piece{'p', 'p', 'p', 'p', 'p', 'p', 'p', 'p'})
	b[39] = ' '
	b[40] = ' '

	// middle
	for row := 4; row <= 7; row++ {
		start := row * 10
		for col := 1; col <= 8; col++ {
			b[start+col] = '.'
		}
		b[start+9] = ' '
		b[start+10] = ' '
	}

	// White pieces
	copy(b[81:89], []Piece{'P', 'P', 'P', 'P', 'P', 'P', 'P', 'P'})
	b[89] = ' '
	b[90] = ' '
	copy(b[91:99], []Piece{'R', 'N', 'B', 'Q', 'K', 'B', 'N', 'R'})
	b[99] = ' '
	b[100] = ' '

	// padding
	for i := 100; i < 120; i++ {
		b[i] = ' '
	}

	return b

}

// Square: index on the board
type Square int

// Specify corners to enable move optimization
const A1, H1, A8, H8 Square = 91, 98, 21, 28

func (s Square) Flip() Square {
	return 119 - s
}

func (s Square) String() string {
	return string([]byte{" abcdefgh "[s%10], "  87654321  "[s/10]})
}

type Move struct {
	from Square
	to   Square
}

// move direction constants
const U, R, D, L = -10, 1, 10, -1

// move as string (algebraic notation is used)
func (m Move) String() string {
	return m.from.String() + m.to.String()
}

type Position struct {
	board          Board
	Score          int     // board score, higher
	white_castling [2]bool // 2 bools representing queenside / kingside castle
	black_castling [2]bool
	en_passant     Square // current en-passant square, if applicable
	king_capture   Square // square where king can be captured during castling
}

func NewStandardPosition() Position {
	p := Position{
		board:          NewStandardBoard(),
		Score:          0,
		white_castling: [2]bool{true, true},
		black_castling: [2]bool{true, true},
		en_passant:     0,
		king_capture:   0,
	}
	return p
}

func (p Position) Flip() Position {
	np := Position{
		Score:          -p.Score,
		white_castling: [2]bool{p.black_castling[0], p.black_castling[1]},
		black_castling: [2]bool{p.white_castling[0], p.white_castling[1]},
		en_passant:     p.en_passant.Flip(),
		king_capture:   p.king_capture.Flip(),
	}
	np.board = p.board.Flip()
	return np
}

// return list of possible moves in the current position
func (p Position) Moves() (moves []Move) {
	var directions = map[Piece][]Square{
		'P': {U, 2 * U, U + L, U + R},
		'N': {2*U + R, 2*U + L, U + 2*L, U + 2*R, D + 2*L, D + 2*R, 2*D + L, 2*D + R},
		'B': {U + R, U + L, D + L, D + R},
		'R': {U, R, D, L},
		'Q': {U, R, D, L, U + R, U + L, D + L, D + R},
		'K': {U, R, D, L, U + R, U + L, D + L, D + R},
	}

	// check all squares. use only MySide pieces
	for index, piece := range p.board {
		if !piece.MySide() {
			continue
		}
		i := Square(index)

		for _, d := range directions[piece] {
			for j := i + d; ; j = j + d {
				sq := p.board[j]
				if sq == ' ' || (sq != '.' && sq.MySide()) {
					break
				}
				if piece == 'P' {
					if (d == U || d == 2*U) && sq != '.' {
						break
					}
					if d == 2*U && (i < A1+U || p.board[i+U] != '.') {
						break
					}
					if (d == U+L || d == U+R) && sq == '.' && (j != p.en_passant && j != p.king_capture && j != p.king_capture-1 && j != p.king_capture+1) {
						break
					}
				}
				moves = append(moves, Move{from: i, to: j})
				if piece == 'P' || piece == 'N' || piece == 'K' || (sq != ' ' && sq != '.' && !sq.MySide()) {
					break
				}
				// Castling
				if i == A1 && p.board[j+R] == 'K' && p.white_castling[0] {
					moves = append(moves, Move{from: j + R, to: j + L})
				}
				if i == H1 && p.board[j+L] == 'K' && p.white_castling[1] {
					moves = append(moves, Move{from: j + L, to: j + R})
				}
			}
		}
	}
	return moves
}

// this function completes a move, and flips the board.
func (p Position) Move(m Move) (np Position) {
	i := m.from
	j := m.to
	piece := p.board[m.from]
	np = p
	np.en_passant = 0
	np.king_capture = 0
	np.Score = p.Score + p.Value(m) // position value
	np.board[m.to] = p.board[m.from]
	np.board[m.from] = '.'
	if i == A1 {
		np.white_castling[0] = false
	}
	if i == H1 {
		np.white_castling[1] = false
	}
	if j == A8 {
		np.black_castling[0] = false
	}
	if j == H8 {
		np.black_castling[1] = false
	}
	if piece == 'K' {
		np.white_castling[0], np.white_castling[1] = false, false
		if (int(j-i)) == 2 || (int(i-j)) == -2 {
			if j < i {
				np.board[H1] = '.' // kingside
			} else {
				np.board[A1] = '.' // queenside
			}
			np.board[(i+j)/2] = 'R'
		}
	}
	if piece == 'P' {
		// promotion
		if A8 <= j && j <= H8 {
			np.board[j] = 'Q'
		}
		// first move
		if j-i == 2*U {
			np.en_passant = i + U
		}
		// en-passant check
		if j == p.en_passant {
			np.board[j+D] = '.'
		}
	}
	return np.Flip()
}

// Position - Value evaluates the score of the position with the given move.
// This leverages a Position-Square-Table common to most chess engines,
// i.e.
func (pos Position) Value(m Move) int {
	pst := map[Piece][120]int{
		'P': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 178, 183, 186, 173, 202, 182, 185, 190, 0, 0, 107, 129, 121, 144, 140, 131, 144, 107, 0, 0, 83, 116, 98, 115, 114, 0, 115, 87, 0, 0, 74, 103, 110, 109, 106, 101, 0, 77, 0, 0, 78, 109, 105, 89, 90, 98, 103, 81, 0, 0, 69, 108, 93, 63, 64, 86, 103, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'N': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 214, 227, 205, 205, 270, 225, 222, 210, 0, 0, 277, 274, 380, 244, 284, 342, 276, 266, 0, 0, 290, 347, 281, 354, 353, 307, 342, 278, 0, 0, 304, 304, 325, 317, 313, 321, 305, 297, 0, 0, 279, 285, 311, 301, 302, 315, 282, 0, 0, 0, 262, 290, 293, 302, 298, 295, 291, 266, 0, 0, 257, 265, 282, 0, 282, 0, 257, 260, 0, 0, 206, 257, 254, 256, 261, 245, 258, 211, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'B': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 261, 242, 238, 244, 297, 213, 283, 270, 0, 0, 309, 340, 355, 278, 281, 351, 322, 298, 0, 0, 311, 359, 288, 361, 372, 310, 348, 306, 0, 0, 345, 337, 340, 354, 346, 345, 335, 330, 0, 0, 333, 330, 337, 343, 337, 336, 0, 327, 0, 0, 334, 345, 344, 335, 328, 345, 340, 335, 0, 0, 339, 340, 331, 326, 327, 326, 340, 336, 0, 0, 313, 322, 305, 308, 306, 305, 310, 310, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'R': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 514, 508, 512, 483, 516, 512, 535, 529, 0, 0, 534, 508, 535, 546, 534, 541, 513, 539, 0, 0, 498, 514, 507, 512, 524, 506, 504, 494, 0, 0, 0, 484, 495, 492, 497, 475, 470, 473, 0, 0, 451, 444, 463, 458, 466, 450, 433, 449, 0, 0, 437, 451, 437, 454, 454, 444, 453, 433, 0, 0, 426, 441, 448, 453, 450, 436, 435, 426, 0, 0, 449, 455, 461, 484, 477, 461, 448, 447, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'Q': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 935, 930, 921, 825, 998, 953, 1017, 955, 0, 0, 943, 961, 989, 919, 949, 1005, 986, 953, 0, 0, 927, 972, 961, 989, 1001, 992, 972, 931, 0, 0, 930, 913, 951, 946, 954, 949, 916, 923, 0, 0, 915, 914, 927, 924, 928, 919, 909, 907, 0, 0, 899, 923, 916, 918, 913, 918, 913, 902, 0, 0, 893, 911, 0, 910, 914, 914, 908, 891, 0, 0, 890, 899, 898, 916, 898, 893, 895, 887, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'K': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 60004, 60054, 60047, 59901, 59901, 60060, 60083, 59938, 0, 0, 59968, 60010, 60055, 60056, 60056, 60055, 60010, 60003, 0, 0, 59938, 60012, 59943, 60044, 59933, 60028, 60037, 59969, 0, 0, 59945, 60050, 60011, 59996, 59981, 60013, 0, 59951, 0, 0, 59945, 59957, 59948, 59972, 59949, 59953, 59992, 59950, 0, 0, 59953, 59958, 59957, 59921, 59936, 59968, 59971, 59968, 0, 0, 59996, 60003, 59986, 59950, 59943, 59982, 60013, 60004, 0, 0, 60017, 60030, 59997, 59986, 60006, 59999, 60040, 60018, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	i := m.from
	j := m.to
	p := Piece(pos.board[i])
	q := Piece(pos.board[j])
	score := pst[p][j] - pst[p][i] // score of the direct move
	if q != '.' && q != ' ' && !q.MySide() {
		score += pst[q.Flip()][j.Flip()]
	}
	// castling
	if int(j-pos.king_capture) < 2 && int(j-pos.king_capture) > -2 {
		score += pst['K'][j.Flip()]
	}
	if p == 'K' && (int(i-j)) == 2 || (int(j-i)) == -2 {
		score = score + pst['R'][(i+j)/2]
		if j < i {
			score = score - pst['R'][A1]
		} else {
			score = score - pst['R'][H1]
		}
	}
	if p == 'P' {
		// promotion
		if A8 <= j && j <= H8 {
			score += pst['Q'][j] - pst['P'][j]
		}
		// en-passant
		if j == pos.en_passant {
			score += pst['P'][(j + D).Flip()]
		}
	}
	return score
}

var MateValue = 75000
var MaxTableSize = 10000000
var EvalRoughness = 13

// https://en.wikipedia.org/wiki/Game_of_the_Century_(chess)
func NewFischerImmortalGameBoard() Board {
	var b Board
	// padding
	for i := 0; i <= 20; i++ {
		b[i] = ' '
	}
	b[21] = 'r'
	b[22] = '.'
	b[23] = '.'
	b[24] = 'q'
	b[25] = '.'
	b[26] = 'r'
	b[27] = 'k'
	b[28] = '.'
	b[29] = ' '
	b[30] = ' '
	b[31] = 'p'
	b[32] = 'p'
	b[33] = '.'
	b[34] = '.'
	b[35] = 'p'
	b[36] = 'p'
	b[37] = 'b'
	b[38] = 'p'
	b[39] = ' '
	b[40] = ' '
	b[41] = '.'
	b[42] = 'n'
	b[43] = 'p'
	b[44] = '.'
	b[45] = '.'
	b[46] = 'n'
	b[47] = '.'
	b[48] = '.'
	b[49] = ' '
	b[50] = ' '
	b[51] = '.'
	b[52] = '.'
	b[53] = 'Q'
	b[54] = '.'
	b[55] = '.'
	b[56] = '.'
	b[57] = 'B'
	b[58] = '.'
	b[59] = ' '
	b[60] = ' '
	b[61] = '.'
	b[62] = '.'
	b[63] = '.'
	b[64] = 'P'
	b[65] = 'P'
	b[66] = '.'
	b[67] = 'b'
	b[68] = '.'
	b[69] = ' '
	b[70] = ' '
	b[71] = '.'
	b[72] = '.'
	b[73] = 'N'
	b[74] = '.'
	b[75] = '.'
	b[76] = 'N'
	b[77] = '.'
	b[78] = '.'
	b[79] = ' '
	b[80] = ' '
	b[81] = 'P'
	b[82] = 'P'
	b[83] = '.'
	b[84] = '.'
	b[85] = '.'
	b[86] = 'P'
	b[87] = 'P'
	b[88] = 'P'
	b[89] = ' '
	b[90] = ' '
	b[91] = '.'
	b[92] = '.'
	b[93] = '.'
	b[94] = 'R'
	b[95] = 'K'
	b[96] = 'B'
	b[97] = '.'
	b[98] = 'R'
	b[99] = ' '
	b[100] = ' '

	// padding
	for i := 100; i < 120; i++ {
		b[i] = ' '
	}

	return b.Flip()

}

func NewFischerPosition() Position {
	p := Position{
		board:          NewFischerImmortalGameBoard(),
		Score:          0,
		white_castling: [2]bool{false, true},
		black_castling: [2]bool{false, false},
		en_passant:     0,
		king_capture:   0,
	}
	return p
}

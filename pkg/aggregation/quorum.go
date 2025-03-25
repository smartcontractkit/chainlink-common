package aggregation

// ByzantineQuorum is a utility to calculated the number of responses from a
// DON before Byzantine tolerant quorum is achieved.
// NOTE: Typical usage is when N >= 3f + 1
//
// The formula is floor( (N+F)/2 ) + 1.
// F should be >= 1.
// If N = 1, a quorum size of 1 is returned.
// Double check usage for small N and F if N < 3f + 1.
func ByzantineQuorum(F int, N int) (Q int) {
	if N == 1 {
		return 1
	}
	return (N+F)/2 + 1
}

package db

func isSubset[E comparable](big []E, short []E) bool {
	if len(short) == 0 {
		return false
	}
	if len(big) < len(short) {
		return false
	}
	for i := 0; i < len(short); i++ {
		if big[i] != short[i] {
			return false
		}
	}
	return true
}

package opus

func MIN16(a,b int16) int16 {
	if (a < b) {
		return a
	} else {
		return b
	}
}
func MAX16(a,b int16) int16 {
	if (a > b) {
		return a
	} else {
		return b
	}
}
func MIN32(a,b int32) int32 {
	if (a < b) {
		return a
	} else {
		return b
	}
}
func MAX32(a,b int32) int32 {
	if (a > b) {
		return a
	} else {
		return b
	}
}
func IMIN(a,b int) int {
	if (a < b) {
		return a
	} else {
		return b
	}
}
func IMAX(a,b int) int {
	if (a > b) {
		return a
	} else {
		return b
	}
}

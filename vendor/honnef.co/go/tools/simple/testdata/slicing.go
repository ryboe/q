package pkg

func fn() {
	var s []int
	_ = s[:len(s)] // MATCH /omit second index/

	len := func(s []int) int { return -1 }
	_ = s[:len(s)]
}

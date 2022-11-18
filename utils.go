package maglevhash

import "math"

func Hash1(k int) int {
	s := uint64(2654435769)
	p := uint32(14)
	tmp := (s * uint64(k)) % (1 << 32)
	return int(tmp / (1 << (32 - p)))
}

func Hash2(k int) int {
	s := uint64(1654435769)
	p := uint32(14)
	tmp := (s * uint64(k)) % (1 << 32)
	return int(tmp / (1 << (32 - p)))
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	end := int(math.Sqrt(float64(n)))
	for i := 2; i <= end; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func findPrime(n int) int {
	// 始终有大于n的质数
	for {
		if isPrime(n) {
			return n
		}
		n++
	}
}

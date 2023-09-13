package lol

import "fmt"

func fibonnaci(n int) int {
	if n <= 1 {
		return n
	}

	return fibonnaci(n-1) + fibonnaci(n-2)
}

func main() {
	fmt.Println(fibonnaci(10))
}

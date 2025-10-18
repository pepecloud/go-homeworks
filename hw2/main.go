// Шахматная доска

package main

import "fmt"

func main() {
	w := 10
	h := 8

	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			if (row+col)%2 == 0 {
				fmt.Print(" ")
			} else {
				fmt.Print("#")
			}
		}
		fmt.Println()
	}
}

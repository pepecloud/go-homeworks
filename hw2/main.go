// Шахматная доска

package main

import "fmt"

func main() {
	var w, h int

	fmt.Print("Введите ширину доски: ")
	fmt.Scanln(&w)

	fmt.Print("Введите высоту доски: ")
	fmt.Scanln(&h)

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

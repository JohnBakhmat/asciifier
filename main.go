package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"sync"

	"github.com/crazy3lf/colorconv"
	"github.com/nfnt/resize"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
func run() error {
	char_greyscale := []rune(" .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$")

	if os.Args[1] == "" {
		return errors.New("Provide image path")
	}

	if os.Args[2] == "" {
		return errors.New("Provide output path")
	}
	image, err := readImage(os.Args[1])
	if err != nil {
		return err
	}

	new_size := uint(math.Pow(2, 8))

	image = resize.Resize(new_size, 0, image, resize.Lanczos3)

	bounds := image.Bounds()

	start_x, start_y := bounds.Min.X, bounds.Min.Y
	end_x, end_y := bounds.Max.X, bounds.Max.Y
	width, height := bounds.Dx(), bounds.Dy()

	char_grid := make([][]rune, height)
	luminance_grid := make([][]int, height)

	for i := range luminance_grid {
		luminance_grid[i] = make([]int, width)
		char_grid[i] = make([]rune, width)
		for j := range luminance_grid[i] {
			luminance_grid[i][j] = 0
			char_grid[i][j] = 0
		}
	}

	minL, maxL := 100, 0

	for y := start_y; y < end_y; y++ {
		for x := start_x; x < end_x; x++ {
			_, _, l := colorconv.ColorToHSL(image.At(x, y))
			l_int := int(l * 100)

			minL = min(minL, l_int)
			maxL = max(maxL, l_int)

			luminance_grid[y][x] = l_int
		}
	}

	n := len(char_greyscale) - 1

	var wg sync.WaitGroup

	for y, row := range luminance_grid {
		wg.Add(1)

		go func() {
			defer wg.Done()

			buffer := make([]rune, width)
			for x, cell := range row {
				i := (int(cell) - minL) * n / (maxL - minL)
				buffer[x] = char_greyscale[i]
			}
			char_grid[y] = buffer
		}()
	}
	wg.Wait()

	outFile, err := os.OpenFile(os.Args[2], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	for _, row := range char_grid {
		outFile.WriteString(fmt.Sprintln(string(row)))
	}

	return nil
}

func readImage(path string) (image.Image, error) {

	image_file, err := os.Open(os.Args[1])
	if err != nil {
		return nil, err
	}

	image, err := png.Decode(image_file)
	if err != nil {
		return nil, err
	}

	return image, nil
}

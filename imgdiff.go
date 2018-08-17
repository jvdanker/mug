package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
)

func main() {
	var (
		input1    = ""
		input2    = ""
		threshold = 0
	)

	flag.StringVar(&input1, "i1", input1, "image 1")
	flag.StringVar(&input2, "i2", input2, "image 2")
	flag.IntVar(&threshold, "t", threshold, "threshold")
	flag.Parse()

	if input1 == "" || input2 == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Comparing image %v to %v with threshold set to %v\n", input1, input2, threshold)

	i1, err := loadImage(input1)
	if err != nil {
		log.Fatal(err)
	}

	i2, err := loadImage(input2)
	if err != nil {
		log.Fatal(err)
	}

	if i1.ColorModel() != i2.ColorModel() {
		log.Fatal("different color models")
	}

	b := i1.Bounds()
	if !b.Eq(i2.Bounds()) {
		log.Fatal("different image sizes")
	}

	var sum int64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r1, g1, b1, _ := i1.At(x, y).RGBA()
			r2, g2, b2, _ := i2.At(x, y).RGBA()
			sum += diff(r1, r2)
			sum += diff(g1, g2)
			sum += diff(b1, b2)
		}
	}

	nPixels := (b.Max.X - b.Min.X) * (b.Max.Y - b.Min.Y)
	diff := float64(sum*100) / (float64(nPixels) * 0xffff * 3)
	fmt.Printf("Image difference: %f%%\n", diff)

	if diff > float64(threshold) {
		os.Exit(1)
	}
}

func loadImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func diff(a, b uint32) int64 {
	if a > b {
		return int64(a - b)
	}
	return int64(b - a)
}

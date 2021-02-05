package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func img2rgba(in image.Image) *image.RGBA {
	b := in.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), in, b.Min, draw.Src)
	return dst
}

func main() {
	rand.Seed(time.Now().UnixNano())

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	img1, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	dst := img2rgba(img1)

	f, err = os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	img2, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	img2 = img2rgba(img2)

	b := dst.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			color1 := dst.At(x, y)
			color2 := img2.At(x, y)
			r2, g2, b2, a2 := color2.RGBA()
			if r, g, b, _ := color1.RGBA(); r == 0xffff && g == 0xffff && b == 0xffff && (r2 != 0xffff || b2 != 0xffff || g2 != 0xffff) && a2 != 0 {
				dst.Set(x, y, color2)
			}
		}
	}

	id := strconv.FormatInt(rand.Int63(), 36)
	f, err = os.Create(id + ".png")
	if err != nil {
		panic(err)
	}

	fmt.Println(id + ".png")

	err = png.Encode(f, dst)
	if err != nil {
		panic(err)
	}
}

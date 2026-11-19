// Command appicon draws the app's icon, assets/Icon.png: two sparkles on a violet-to-pink rounded square, 1024 pixels wide.
//
//	go run ./tools/appicon
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/vector"
)

const size = 1024

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// A vertical gradient, masked by a rounded square.
	top, bottom := color.RGBA{0x7c, 0x5c, 0xff, 0xff}, color.RGBA{0xff, 0x6f, 0xa8, 0xff}
	grad := image.NewRGBA(img.Bounds())
	for y := 0; y < size; y++ {
		t := float64(y) / (size - 1)
		c := color.RGBA{lerp(top.R, bottom.R, t), lerp(top.G, bottom.G, t), lerp(top.B, bottom.B, t), 0xff}
		for x := 0; x < size; x++ {
			grad.SetRGBA(x, y, c)
		}
	}
	mask := vector.NewRasterizer(size, size)
	roundedRect(mask, 0, 0, size, size, 230)
	mask.Draw(img, img.Bounds(), grad, image.Point{})

	white := image.NewUniform(color.RGBA{0xff, 0xff, 0xff, 0xff})
	stars := vector.NewRasterizer(size, size)
	sparkle(stars, 450, 560, 300)
	sparkle(stars, 760, 290, 120)
	sparkle(stars, 770, 760, 70)
	stars.Draw(img, img.Bounds(), white, image.Point{})

	out := image.NewRGBA(img.Bounds())
	draw.Draw(out, out.Bounds(), img, image.Point{}, draw.Src)
	f, err := os.Create("assets/Icon.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, out); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func lerp(a, b uint8, t float64) uint8 { return uint8(float64(a) + (float64(b)-float64(a))*t + 0.5) }

// sparkle adds a four-pointed star whose points are r from its center, with sides curving in.
func sparkle(z *vector.Rasterizer, cx, cy, r float32) {
	k := r * 0.12 // how far the sides bend in toward the center
	z.MoveTo(cx, cy-r)
	z.QuadTo(cx+k, cy-k, cx+r, cy)
	z.QuadTo(cx+k, cy+k, cx, cy+r)
	z.QuadTo(cx-k, cy+k, cx-r, cy)
	z.QuadTo(cx-k, cy-k, cx, cy-r)
	z.ClosePath()
}

func roundedRect(z *vector.Rasterizer, x0, y0, x1, y1, r float32) {
	z.MoveTo(x0+r, y0)
	z.LineTo(x1-r, y0)
	z.QuadTo(x1, y0, x1, y0+r)
	z.LineTo(x1, y1-r)
	z.QuadTo(x1, y1, x1-r, y1)
	z.LineTo(x0+r, y1)
	z.QuadTo(x0, y1, x0, y1-r)
	z.LineTo(x0, y0+r)
	z.QuadTo(x0, y0, x0+r, y0)
	z.ClosePath()
}

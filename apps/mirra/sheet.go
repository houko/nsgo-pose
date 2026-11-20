package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strconv"

	// The pose gallery's references are PNG; JPEG is decoded too, for a gallery that changes format.
	_ "image/jpeg"

	"golang.org/x/image/draw"
)

// sheetHeight is the height each angle of a sheet's references is scaled to: five portrait mannequins then make an image about 2,600 pixels wide, under what the image APIs accept.
const sheetHeight = 768

// sideBySide puts the pose's angles in one image, left to right on white with a gap between them, so the model is sent one reference for the whole sheet: some APIs take only a few images.
func sideBySide(views []part) (part, error) {
	const gap = 24
	var imgs []image.Image
	width := gap
	for _, v := range views {
		img, _, err := image.Decode(bytes.NewReader(v.data))
		if err != nil {
			return part{}, fmt.Errorf("reading %s: %w", v.name, err)
		}
		b := img.Bounds()
		width += b.Dx()*sheetHeight/max(b.Dy(), 1) + gap
		imgs = append(imgs, img)
	}
	out := image.NewRGBA(image.Rect(0, 0, width, sheetHeight+2*gap))
	draw.Draw(out, out.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	x := gap
	for _, img := range imgs {
		b := img.Bounds()
		w := b.Dx() * sheetHeight / max(b.Dy(), 1)
		draw.CatmullRom.Scale(out, image.Rect(x, gap, x+w, gap+sheetHeight), img, b, draw.Over, nil)
		x += w + gap
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return part{}, err
	}
	return part{name: "angles.png", mime: "image/png", data: buf.Bytes()}, nil
}

// landscape turns a portrait size such as "1024x1536" on its side, for a sheet of angles side by side; a landscape or square size, or one written another way such as "2K", is kept.
func landscape(size string) string {
	w, h, ok := dimensions(size)
	if !ok || w >= h {
		return size
	}
	return strconv.Itoa(h) + "x" + strconv.Itoa(w)
}

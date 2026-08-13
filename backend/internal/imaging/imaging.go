package imaging

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
)

func Decode(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	return img, err
}

func Composite(base, overlay image.Image) *image.RGBA {
	dst := image.NewRGBA(base.Bounds())
	draw.Draw(dst, dst.Bounds(), base, base.Bounds().Min, draw.Src)
	draw.Draw(dst, dst.Bounds(), overlay, overlay.Bounds().Min, draw.Over)
	return dst
}

func EncodePNG(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}

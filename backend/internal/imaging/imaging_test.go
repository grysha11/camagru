package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func encodeTestPNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	return buf.Bytes()
}

func TestDecodeValidPNG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	data := encodeTestPNG(t, src)

	img, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Errorf("decoded bounds = %v, want 4x4", img.Bounds())
	}
}

func TestDecodeInvalidBytes(t *testing.T) {
	if _, err := Decode(bytes.NewReader([]byte("not an image"))); err == nil {
		t.Error("expected error decoding garbage bytes, got nil")
	}
}

func TestCompositeAlphaBlend(t *testing.T) {
	base := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	baseColor := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	for x := 0; x < 3; x++ {
		base.Set(x, 0, baseColor)
	}

	overlay := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	overlay.Set(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 0})
	overlay.Set(1, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 128})
	overlay.Set(2, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	dst := Composite(base, overlay)

	if got := color.NRGBAModel.Convert(dst.At(0, 0)).(color.NRGBA); got != baseColor {
		t.Errorf("transparent overlay pixel = %+v, want base color %+v", got, baseColor)
	}

	wantOpaque := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	if got := color.NRGBAModel.Convert(dst.At(2, 0)).(color.NRGBA); got != wantOpaque {
		t.Errorf("opaque overlay pixel = %+v, want overlay color %+v", got, wantOpaque)
	}

	blended := color.NRGBAModel.Convert(dst.At(1, 0)).(color.NRGBA)
	if blended == baseColor || blended == wantOpaque {
		t.Errorf("partial overlay pixel = %+v, expected a blend distinct from both base %+v and overlay %+v", blended, baseColor, wantOpaque)
	}
	if blended.G >= baseColor.G || blended.B >= baseColor.B {
		t.Errorf("partial overlay pixel = %+v, expected green/blue to drop from the base as red overlay blends in", blended)
	}
	if blended.A != 255 {
		t.Errorf("partial overlay pixel alpha = %d, want 255", blended.A)
	}
}

func TestEncodePNGRoundTrip(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 5, 7))
	src.Set(2, 3, color.RGBA{R: 10, G: 20, B: 30, A: 255})

	var buf bytes.Buffer
	if err := EncodePNG(&buf, src); err != nil {
		t.Fatalf("EncodePNG: %v", err)
	}

	decoded, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode round trip: %v", err)
	}
	if decoded.Bounds().Dx() != 5 || decoded.Bounds().Dy() != 7 {
		t.Errorf("round-tripped bounds = %v, want 5x7", decoded.Bounds())
	}

	got := color.RGBAModel.Convert(decoded.At(2, 3)).(color.RGBA)
	want := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	if got != want {
		t.Errorf("round-tripped pixel = %+v, want %+v", got, want)
	}
}

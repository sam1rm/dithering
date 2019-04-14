package dithering

import (
	"image"
	"image/color"
	"math"
)

// PixelError represents the error for each canal in the image
// when dithering an image
// Errors are floats because they are the result of a division
type PixelError struct {
	R, G, B, A float32
}

func (c PixelError) RGBA() (r, g, b, a uint32) {
	return uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)
}

func (c PixelError) Add(c2 PixelError) PixelError {
	r := c.R + c2.R
	g := c.G + c2.G
	b := c.B + c2.B
	return PixelError{r, g, b, 0}
}

func (c PixelError) Mul(v float32) PixelError {
	r := c.R * v
	g := c.G * v
	b := c.B * v
	return PixelError{r, g, b, 0}
}

func pixelErrorModel(c color.Color) color.Color {
	if _, ok := c.(PixelError); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	return PixelError{float32(r), float32(g), float32(b), float32(a)}
}

// ErrorImage is an in-memory image whose At method returns dithering.PixelError values
type ErrorImage struct {
	// Pix holds the image's pixels, in R, G, B, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []float32
	// Stride is the Pix stride between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
	// Min & Max values in the image
	Min, Max PixelError
}

func (p *ErrorImage) ColorModel() color.Model {
	return color.ModelFunc(pixelErrorModel)
}

func (p *ErrorImage) Bounds() image.Rectangle { return p.Rect }

func (p *ErrorImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return PixelError{}
	}
	i := p.PixOffset(x, y)

	r := (p.Pix[i+0]) + float32(math.Abs(float64(p.Min.R))) / (p.Max.R - p.Min.R) * 255
	g := (p.Pix[i+1]) + float32(math.Abs(float64(p.Min.G))) / (p.Max.G - p.Min.G) * 255
	b := (p.Pix[i+2]) + float32(math.Abs(float64(p.Min.B))) / (p.Max.B - p.Min.B) * 255

	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func (p *ErrorImage) PixelErrorAt(x, y int) PixelError {
	if !(image.Point{x, y}.In(p.Rect)) {
		return PixelError{}
	}
	i := p.PixOffset(x, y)
	r := p.Pix[i+0]
	g := p.Pix[i+1]
	b := p.Pix[i+2]
	a := p.Pix[i+3]

	return PixelError{r, g, b, a}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *ErrorImage) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*4
}

func (p *ErrorImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	c1 := color.ModelFunc(pixelErrorModel).Convert(c).(PixelError)
	if c1.R > p.Max.R {
		p.Max.R = c1.R
	}
	if c1.G > p.Max.G {
		p.Max.G = c1.G
	}
	if c1.B > p.Max.B {
		p.Max.B = c1.B
	}
	if c1.R < p.Min.R {
		p.Min.R = c1.R
	}
	if c1.G < p.Min.G {
		p.Min.G = c1.G
	}
	if c1.B < p.Min.B {
		p.Min.B = c1.B
	}
	p.Pix[i+0] = c1.R
	p.Pix[i+1] = c1.G
	p.Pix[i+2] = c1.B
	p.Pix[i+3] = c1.A
}

func (p *ErrorImage) SetPixelError(x, y int, c PixelError) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	if c.R > p.Max.R {
		p.Max.R = c.R
	}
	if c.G > p.Max.G {
		p.Max.G = c.G
	}
	if c.B > p.Max.B {
		p.Max.B = c.B
	}
	if c.R < p.Min.R {
		p.Min.R = c.R
	}
	if c.G < p.Min.G {
		p.Min.G = c.G
	}
	if c.B < p.Min.B {
		p.Min.B = c.B
	}
	i := p.PixOffset(x, y)
	p.Pix[i+0] = c.R
	p.Pix[i+1] = c.G
	p.Pix[i+2] = c.B
	p.Pix[i+3] = c.A
}

func NewErrorImage(r image.Rectangle) *ErrorImage {
	w, h := r.Dx(), r.Dy()
	buf := make([]float32, 4*w*h)
	return &ErrorImage{buf, 4 * w, r, PixelError{}, PixelError{}}
}

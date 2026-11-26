package main

import (
	"bytes"
	"image"
	"image/color"
	_ "image/gif" // register decoders
	"image/jpeg"
	_ "image/png"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// Size limits for what ends up in R2. The browser already shrinks pictures;
// this is the safety net for anything that arrives bigger.
const (
	avatarSide = 512
	photoSide  = 1600
	jpegQ      = 78
)

// optimizeImage shrinks a picture so its longest side is at most maxSide and
// re-encodes it as JPEG, unless the original is already small enough and smaller
// than the re-encoded version. Formats Go can't decode (HEIC) are kept as they are.
func optimizeImage(data []byte, ctype, ext string, maxSide int) ([]byte, string, string) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return data, ctype, ext
	}
	fits := cfg.Width <= maxSide && cfg.Height <= maxSide
	compact := ctype == "image/webp" || ctype == "image/jpeg"
	if fits && compact && len(data) <= 350<<10 {
		return data, ctype, ext // already what the browser sends: keep it
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, ctype, ext
	}
	w, h := cfg.Width, cfg.Height
	if !fits {
		if w >= h {
			h = max(1, h*maxSide/w)
			w = maxSide
		} else {
			w = max(1, w*maxSide/h)
			h = maxSide
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	// White background so transparent PNGs don't turn black in JPEG.
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: jpegQ}); err != nil {
		return data, ctype, ext
	}
	if fits && compact && buf.Len() >= len(data) {
		return data, ctype, ext
	}
	return buf.Bytes(), "image/jpeg", "jpg"
}

package qr

import "image/color"

var CU = Config{
	Size:            512,
	LogoScale:       0.2,
	Smoothing:       1,
	Background:      color.RGBA{R: 20, G: 20, B: 20, A: 255},
	Foreground:      color.RGBA{R: 230, G: 230, B: 230, A: 255},
	CornerRadius:    0.0015,
	RecoveryLevel:   3,
	QuietZone:       1,
	LogoBackground:  color.RGBA{R: 20, G: 20, B: 20, A: 255},
	LogoBorderWidth: 1,
	LogoFade:        1,
}

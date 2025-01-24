package generator

import "image/color"

type QRCodeConfig struct {
	Size            int
	LogoScale       float64
	Smoothing       float64 // Controls the overall smoothness of the QR code
	Background      color.Color
	Foreground      color.Color
	CornerRadius    float64 // Controls individual dot roundness
	RecoveryLevel   int
	QuietZone       int // Size of quiet zone around QR code
	LogoBackground  color.Color
	LogoBorderWidth float64 // Width of logo border
	LogoFade        float64 // Logo fade effect
}

var CU = &QRCodeConfig{
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

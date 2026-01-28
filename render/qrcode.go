package render

import (
	"image"
	"image/draw"

	"github.com/skip2/go-qrcode"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// zplECToQRCode maps ZPL error correction levels to go-qrcode levels.
var zplECToQRCode = map[zpl.QRCodeErrorCorrection]qrcode.RecoveryLevel{
	zpl.QRCodeECLow:      qrcode.Low,     // ~7% recovery
	zpl.QRCodeECMedium:   qrcode.Medium,  // ~15% recovery
	zpl.QRCodeECQuartile: qrcode.High,    // ~25% recovery (go-qrcode calls this "High")
	zpl.QRCodeECHigh:     qrcode.Highest, // ~30% recovery
}

// drawQRCode renders a QR code at the current position.
func (c *canvas) drawQRCode(bc *zpl.BarcodeQR) {
	if bc.Data == "" {
		return
	}

	// Map ZPL error correction to go-qrcode level
	ecLevel, ok := zplECToQRCode[bc.ErrorCorrection]
	if !ok {
		ecLevel = qrcode.Medium // Default
	}

	// Create the QR code
	qr, err := qrcode.New(bc.Data, ecLevel)
	if err != nil {
		// Silently skip if encoding fails (data may be too long or invalid)
		return
	}

	// Calculate size based on magnification factor
	// ZPL magnification 1-10 maps to module size in dots
	// Typical QR codes are 21x21 to 177x177 modules depending on version
	// Magnification factor is the size of each module in dots
	moduleSize := bc.MagnificationFactor
	if moduleSize < 1 {
		moduleSize = 3 // Default magnification
	}
	if moduleSize > 10 {
		moduleSize = 10
	}

	// Calculate target size
	// QR codes have a 4-module quiet zone by default
	// The size formula is: version*4 + 17 modules (plus quiet zone)
	qrModules := qr.VersionNumber*4 + 17 + 8 // +8 for quiet zone (4 on each side)
	targetSize := qrModules * moduleSize

	// Generate the QR code image at target size
	qrImg := qr.Image(targetSize)

	// Draw the QR code image onto our canvas at the current position
	x := c.curX
	y := c.curY

	// Handle orientation by rotating the QR code image if needed
	var finalImg = qrImg
	switch bc.Orientation {
	case zpl.OrientationRotated90: // 90° clockwise
		finalImg = rotateImage90CW(qrImg)
	case zpl.OrientationRotated180: // 180°
		finalImg = rotateImage180(qrImg)
	case zpl.OrientationRotated270: // 270° clockwise
		finalImg = rotateImage270CW(qrImg)
	}

	// Draw onto canvas
	draw.Draw(c.img, image.Rect(x, y, x+finalImg.Bounds().Dx(), y+finalImg.Bounds().Dy()),
		finalImg, image.Point{0, 0}, draw.Over)
}

// rotateImage90CW rotates an image 90 degrees clockwise.
func rotateImage90CW(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rotated := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rotated.Set(h-1-y, x, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return rotated
}

// rotateImage180 rotates an image 180 degrees.
func rotateImage180(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rotated := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rotated.Set(w-1-x, h-1-y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return rotated
}

// rotateImage270CW rotates an image 270 degrees clockwise (90 counter-clockwise).
func rotateImage270CW(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rotated := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rotated.Set(y, w-1-x, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return rotated
}

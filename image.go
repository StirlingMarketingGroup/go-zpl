package zpl

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"image"
	"image/color"
	_ "image/gif"  // Register GIF decoder
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"io"
	"math"
)

// Dithering represents the dithering algorithm to use when converting images.
type Dithering int

const (
	// DitheringNone uses simple threshold (no dithering).
	DitheringNone Dithering = iota
	// DitheringFloydSteinberg uses Floyd-Steinberg error diffusion dithering.
	DitheringFloydSteinberg
	// DitheringOrdered uses ordered (Bayer matrix) dithering.
	DitheringOrdered
)

// ImageConverter converts images to ZPL graphic fields.
type ImageConverter struct {
	// Threshold is the grayscale cutoff (0-255) for black/white conversion.
	// Pixels darker than this become black. Default is 128.
	Threshold uint8

	// Dithering specifies the dithering algorithm to use.
	Dithering Dithering

	// Invert swaps black and white in the output.
	Invert bool
}

// NewImageConverter creates a new image converter with default settings.
func NewImageConverter() *ImageConverter {
	return &ImageConverter{
		Threshold: 128,
		Dithering: DitheringNone,
		Invert:    false,
	}
}

// WithThreshold sets the black/white threshold (0-255).
func (c *ImageConverter) WithThreshold(t uint8) *ImageConverter {
	c.Threshold = t
	return c
}

// WithDithering sets the dithering algorithm.
func (c *ImageConverter) WithDithering(d Dithering) *ImageConverter {
	c.Dithering = d
	return c
}

// WithInvert sets whether to invert the output.
func (c *ImageConverter) WithInvert(invert bool) *ImageConverter {
	c.Invert = invert
	return c
}

// Convert converts an image to a ZPL GraphicField.
// The image is converted to 1-bit monochrome using the configured dithering.
func (c *ImageConverter) Convert(img image.Image) *GraphicField {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate bytes per row (8 pixels per byte, rounded up)
	bytesPerRow := (width + 7) / 8

	// Convert to grayscale first
	gray := c.toGrayscale(img)

	// Apply dithering to get 1-bit output
	bits := c.dither(gray, width, height)

	// Convert to hex string
	data := c.bitsToHex(bits, width, height, bytesPerRow)

	totalBytes := bytesPerRow * height

	return &GraphicField{
		Format:      GraphicFieldASCII,
		DataBytes:   totalBytes,
		TotalBytes:  totalBytes,
		BytesPerRow: bytesPerRow,
		Data:        data,
	}
}

// ConvertBinary converts an image to a ZPL GraphicField using binary format.
// Binary format is more compact than ASCII hex.
func (c *ImageConverter) ConvertBinary(img image.Image) *GraphicField {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate bytes per row (8 pixels per byte, rounded up)
	bytesPerRow := (width + 7) / 8

	// Convert to grayscale first
	gray := c.toGrayscale(img)

	// Apply dithering to get 1-bit output
	bits := c.dither(gray, width, height)

	// Convert to binary data
	data := c.bitsToBytes(bits, width, height, bytesPerRow)

	return &GraphicField{
		Format:      GraphicFieldBinary,
		DataBytes:   len(data),
		TotalBytes:  len(data),
		BytesPerRow: bytesPerRow,
		BinaryData:  data,
	}
}

// ConvertZ64 converts an image to a ZPL GraphicField using Z64 compression.
// Z64 is base64-encoded zlib-compressed data, which is typically much smaller
// than ASCII hex format for complex images.
func (c *ImageConverter) ConvertZ64(img image.Image) *GraphicField {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate bytes per row (8 pixels per byte, rounded up)
	bytesPerRow := (width + 7) / 8

	// Convert to grayscale first
	gray := c.toGrayscale(img)

	// Apply dithering to get 1-bit output
	bits := c.dither(gray, width, height)

	// Convert to binary data
	binaryData := c.bitsToBytes(bits, width, height, bytesPerRow)

	// Compress with zlib
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	_, _ = zw.Write(binaryData) // bytes.Buffer never errors
	_ = zw.Close()              // Close flushes but can't error on bytes.Buffer

	// Encode as base64
	b64Data := base64.StdEncoding.EncodeToString(compressed.Bytes())

	// Calculate CRC16 of the base64 string (ZPL uses CRC-CCITT)
	crc := crc16CCITT([]byte(b64Data))

	// Format: :Z64:<base64>:<crc>
	z64Data := ":Z64:" + b64Data + ":" + crc

	totalBytes := bytesPerRow * height

	return &GraphicField{
		Format:      GraphicFieldASCII, // Z64 uses A format with special data prefix
		DataBytes:   len(compressed.Bytes()),
		TotalBytes:  totalBytes,
		BytesPerRow: bytesPerRow,
		Data:        z64Data,
	}
}

// crc16CCITT calculates CRC-16-CCITT checksum used by ZPL Z64 format.
func crc16CCITT(data []byte) string {
	// CRC-16-CCITT polynomial: 0x1021, initial value: 0xFFFF
	crc := uint16(0xFFFF)

	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}

	// Return as uppercase hex
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[2:], crc)
	return string([]byte{
		"0123456789ABCDEF"[buf[2]>>4],
		"0123456789ABCDEF"[buf[2]&0x0F],
		"0123456789ABCDEF"[buf[3]>>4],
		"0123456789ABCDEF"[buf[3]&0x0F],
	})
}

// ImageToGraphicFieldZ64 converts an image to a GraphicField using Z64 compression.
func ImageToGraphicFieldZ64(img image.Image) *GraphicField {
	return NewImageConverter().ConvertZ64(img)
}

// ConvertReader reads an image from a reader and converts it to a GraphicField.
func (c *ImageConverter) ConvertReader(r io.Reader) (*GraphicField, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return c.Convert(img), nil
}

// toGrayscale converts an image to grayscale values (0-255).
func (c *ImageConverter) toGrayscale(img image.Image) [][]uint8 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	gray := make([][]uint8, height)
	for y := 0; y < height; y++ {
		gray[y] = make([]uint8, width)
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

			// Handle transparency - treat transparent pixels as white
			if a == 0 {
				gray[y][x] = 255
				continue
			}

			// Convert to grayscale using luminance formula
			// Note: RGBA() returns 16-bit values, so divide by 256
			lum := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256.0
			gray[y][x] = uint8(lum)
		}
	}
	return gray
}

// dither applies the configured dithering algorithm and returns 1-bit output.
// true = black, false = white
func (c *ImageConverter) dither(gray [][]uint8, width, height int) [][]bool {
	bits := make([][]bool, height)
	for y := range bits {
		bits[y] = make([]bool, width)
	}

	switch c.Dithering {
	case DitheringFloydSteinberg:
		c.floydSteinberg(gray, bits, width, height)
	case DitheringOrdered:
		c.orderedDither(gray, bits, width, height)
	default:
		c.thresholdDither(gray, bits, width, height)
	}

	// Apply invert if needed
	if c.Invert {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				bits[y][x] = !bits[y][x]
			}
		}
	}

	return bits
}

// thresholdDither applies simple threshold conversion.
func (c *ImageConverter) thresholdDither(gray [][]uint8, bits [][]bool, width, height int) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			bits[y][x] = gray[y][x] < c.Threshold
		}
	}
}

// floydSteinberg applies Floyd-Steinberg error diffusion dithering.
func (c *ImageConverter) floydSteinberg(gray [][]uint8, bits [][]bool, width, height int) {
	// Work with float errors for precision
	errors := make([][]float64, height)
	for y := range errors {
		errors[y] = make([]float64, width)
		for x := range errors[y] {
			errors[y][x] = float64(gray[y][x])
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			oldPixel := errors[y][x]
			var newPixel float64
			if oldPixel < float64(c.Threshold) {
				newPixel = 0
				bits[y][x] = true // black
			} else {
				newPixel = 255
				bits[y][x] = false // white
			}

			quantError := oldPixel - newPixel

			// Distribute error to neighbors (Floyd-Steinberg coefficients)
			//        X   7/16
			// 3/16  5/16  1/16
			if x+1 < width {
				errors[y][x+1] += quantError * 7 / 16
			}
			if y+1 < height {
				if x > 0 {
					errors[y+1][x-1] += quantError * 3 / 16
				}
				errors[y+1][x] += quantError * 5 / 16
				if x+1 < width {
					errors[y+1][x+1] += quantError * 1 / 16
				}
			}
		}
	}
}

// Bayer 4x4 ordered dithering matrix
var bayerMatrix = [4][4]uint8{
	{0, 8, 2, 10},
	{12, 4, 14, 6},
	{3, 11, 1, 9},
	{15, 7, 13, 5},
}

// orderedDither applies ordered (Bayer matrix) dithering.
func (c *ImageConverter) orderedDither(gray [][]uint8, bits [][]bool, width, height int) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Scale the Bayer value to threshold range
			threshold := float64(bayerMatrix[y%4][x%4]) / 16.0 * 255.0
			// Adjust threshold based on configured threshold
			adjustedThreshold := threshold + float64(c.Threshold) - 128

			bits[y][x] = float64(gray[y][x]) < adjustedThreshold
		}
	}
}

// bitsToHex converts 1-bit data to hex string.
func (c *ImageConverter) bitsToHex(bits [][]bool, width, height, bytesPerRow int) string {
	hexChars := "0123456789ABCDEF"
	result := make([]byte, 0, bytesPerRow*height*2)

	for y := 0; y < height; y++ {
		for byteIdx := 0; byteIdx < bytesPerRow; byteIdx++ {
			var b byte
			for bit := 0; bit < 8; bit++ {
				x := byteIdx*8 + bit
				if x < width && bits[y][x] {
					b |= 1 << (7 - bit)
				}
			}
			result = append(result, hexChars[b>>4], hexChars[b&0x0F])
		}
	}

	return string(result)
}

// bitsToBytes converts 1-bit data to binary bytes.
func (c *ImageConverter) bitsToBytes(bits [][]bool, width, height, bytesPerRow int) []byte {
	result := make([]byte, 0, bytesPerRow*height)

	for y := 0; y < height; y++ {
		for byteIdx := 0; byteIdx < bytesPerRow; byteIdx++ {
			var b byte
			for bit := 0; bit < 8; bit++ {
				x := byteIdx*8 + bit
				if x < width && bits[y][x] {
					b |= 1 << (7 - bit)
				}
			}
			result = append(result, b)
		}
	}

	return result
}

// ImageToGraphicField is a convenience function that converts an image to a GraphicField
// using default settings (threshold 128, no dithering).
func ImageToGraphicField(img image.Image) *GraphicField {
	return NewImageConverter().Convert(img)
}

// ImageToGraphicFieldDithered converts an image to a GraphicField using Floyd-Steinberg dithering.
func ImageToGraphicFieldDithered(img image.Image) *GraphicField {
	return NewImageConverter().WithDithering(DitheringFloydSteinberg).Convert(img)
}

// ResizeImage resizes an image to fit within the given dimensions while maintaining aspect ratio.
// If maxWidth or maxHeight is 0, that dimension is not constrained.
func ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scale factors
	scaleX := 1.0
	scaleY := 1.0

	if maxWidth > 0 && width > maxWidth {
		scaleX = float64(maxWidth) / float64(width)
	}
	if maxHeight > 0 && height > maxHeight {
		scaleY = float64(maxHeight) / float64(height)
	}

	// Use the smaller scale to maintain aspect ratio
	scale := math.Min(scaleX, scaleY)
	if scale >= 1.0 {
		return img // No resize needed
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Create new image with nearest-neighbor scaling (good for label printing)
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		srcY := int(float64(y) / scale)
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scale)
			dst.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return dst
}

// GrayscaleImage converts an image to grayscale.
func GrayscaleImage(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}

	return gray
}

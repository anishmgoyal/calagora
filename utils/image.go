package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
)

const (
	// MimeGif is the mime type for a gif image
	MimeGif = "image/gif"
	// MimeJpeg is the mime type for a jpeg image
	MimeJpeg = "image/jpeg"
	// MimePng is the mime type for a png image
	MimePng = "image/png"

	// FullSizeMidwayResize is the size of one step in the progressive downscale
	FullSizeMidwayResize = 1080
	// ThumbnailMidwayResize is the size of one step in the progressive downscale
	ThumbnailMidwayResize = 320
	// MaxFullSizeDim is the largest width or height of a fullsize image
	MaxFullSizeDim = 600
	// MaxThumbnailDim is the largest width or height of a thumbnail
	MaxThumbnailDim = 150

	imageThreadCount = 5
	imageChanSize    = 500
)

var supportedMimeTypes = map[string]bool{
	"image/gif":  true,
	"image/jpeg": true,
	"image/png":  true,
}

// StartImageService spawns threads to handle image processing requests
// and returns the channel they will listen to
func StartImageService() chan *ImageProcessRequest {
	ch := make(chan *ImageProcessRequest, imageChanSize)
	for i := 0; i < imageThreadCount; i++ {
		go processImages(ch)
	}
	return ch
}

// GetImageURLFromPrefix is a simple, standard way of retrieving the url of
// an image based on what is stored in the database
func GetImageURLFromPrefix(prefix string) string {
	return prefix + ".jpg"
}

// GetThumbnailURLFromPrefix is a simple, standard way of retrieving the url of
// an image based on what is stored in the database
func GetThumbnailURLFromPrefix(prefix string) string {
	return prefix + "_thumb.jpg"
}

// DeleteImage removes an image from permanent storage on S3, returns whether
// or not the operation was successful
func DeleteImage(prefix string) bool {
	lastIndex := 0
	for i := len(prefix) - 1; i > 0; i-- {
		if prefix[i] == '/' {
			lastIndex = i
			break
		}
	}
	requestedName := prefix[lastIndex+1:]
	return DeleteFileFromPublic(requestedName+".jpg") &&
		DeleteFileFromPublic(requestedName+"_thumb.jpg")
}

func processImages(ch chan *ImageProcessRequest) {
	for {
		ipr := <-ch
		processImage(ipr)
	}
}

func processImage(r *ImageProcessRequest) {
	defer os.Remove(r.File.Name())
	defer r.File.Close()

	var srcImage image.Image
	var err error

	r.File.Seek(0, 0)

	switch r.MimeType {
	case MimeGif:
		srcImage, err = gif.Decode(r.File)
	case MimeJpeg:
		srcImage, err = jpeg.Decode(r.File)
	case MimePng:
		srcImage, err = png.Decode(r.File)
	}

	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err.Error())
		if r.Error != nil {
			r.Error(r)
		}
		return
	}

	var orientation = getImageExifOrientation(r.File)

	fullSizeFile, err := ioutil.TempFile(TempDirectory, "processing")
	if err != nil {
		fmt.Println("Failed to create temp file.")
		if r.Error != nil {
			r.Error(r)
		}
		return
	}
	defer os.Remove(fullSizeFile.Name())
	defer fullSizeFile.Close()

	resized := resizeToFitBilinear(srcImage, FullSizeMidwayResize)
	resized = reorientImage(resized, orientation)

	var fullSize image.Image
	if srcImage.Bounds().Dx() > (FullSizeMidwayResize<<1) ||
		srcImage.Bounds().Dy() > (FullSizeMidwayResize<<1) {

		fullSize = resizeToFitBicubic(blurImage(resized), MaxFullSizeDim)
	} else {
		fullSize = resizeToFitBicubic(resized, MaxFullSizeDim)
	}

	err = jpeg.Encode(fullSizeFile, fullSize, &jpeg.Options{Quality: 72})
	if err == nil {
		UploadFileToPublic(r.RequestedName+".jpg", MimeJpeg, fullSizeFile)
	} else {
		fmt.Println("Err saving file.")
		if r.Error != nil {
			r.Error(r)
		}
	}

	thumbnailFile, err := ioutil.TempFile(TempDirectory, "processing")
	if err != nil {
		fmt.Println("Failed to create thumbnail.")
		return
	}
	defer os.Remove(thumbnailFile.Name())
	defer thumbnailFile.Close()

	resized = resizeToFitBilinear(fullSize, ThumbnailMidwayResize)
	thumbnail := resizeToFitBicubic(resized, MaxThumbnailDim)

	err = jpeg.Encode(thumbnailFile, thumbnail, &jpeg.Options{Quality: 72})
	if err == nil {
		UploadFileToPublic(r.RequestedName+"_thumb.jpg", MimeJpeg, thumbnailFile)
	} else {
		fmt.Println("Err saving file.")
		if r.Error != nil {
			r.Error(r)
		}
	}
	if r.Success != nil {
		r.Success(r)
	}
}

// Resizes to fit a bounding box using bicubic interpolation
func resizeToFitBilinear(original image.Image,
	edgeSize int) image.Image {

	oldBounds := original.Bounds()
	if oldBounds.Dx() < edgeSize && oldBounds.Dy() < edgeSize {
		// No resize necessary
		return original
	}

	newBounds := getNewBounds(oldBounds, edgeSize)

	resized := image.NewRGBA(newBounds)

	var ratioX = float64(oldBounds.Dx()) / float64(newBounds.Dx())
	var ratioY = float64(oldBounds.Dy()) / float64(newBounds.Dy())

	for x := 0; x < newBounds.Dx(); x++ {
		for y := 0; y < newBounds.Dy(); y++ {
			sourceX := ratioX * float64(x)
			minX := int(math.Floor(sourceX))
			coeffX := sourceX - float64(minX)

			sourceY := ratioY * float64(y)
			minY := int(math.Floor(sourceY))
			coeffY := sourceY - float64(minY)

			r0, g0, b0, a0 := quantizeColorRGBA(
				original.At(clampBounds(minX, minY, oldBounds)))
			r1, g1, b1, a1 := quantizeColorRGBA(
				original.At(clampBounds(minX+1, minY, oldBounds)))
			r2, g2, b2, a2 := quantizeColorRGBA(
				original.At(clampBounds(minX, minY+1, oldBounds)))
			r3, g3, b3, a3 := quantizeColorRGBA(
				original.At(clampBounds(minX+1, minY+1, oldBounds)))

			r := r0*((1-coeffX)*(1-coeffY)) +
				r1*(coeffX*(1-coeffY)) +
				r2*((1-coeffX)*coeffY) +
				r3*(coeffX*coeffY)
			g := g0*((1-coeffX)*(1-coeffY)) +
				g1*(coeffX*(1-coeffY)) +
				g2*((1-coeffX)*coeffY) +
				g3*(coeffX*coeffY)
			b := b0*((1-coeffX)*(1-coeffY)) +
				b1*(coeffX*(1-coeffY)) +
				b2*((1-coeffX)*coeffY) +
				b3*(coeffX*coeffY)
			a := a0*((1-coeffX)*(1-coeffY)) +
				a1*(coeffX*(1-coeffY)) +
				a2*((1-coeffX)*coeffY) +
				a3*(coeffX*coeffY)

			rf := uint8(clampRangeFloat(0, r, 255))
			gf := uint8(clampRangeFloat(0, g, 255))
			bf := uint8(clampRangeFloat(0, b, 255))
			af := uint8(clampRangeFloat(0, a, 255))

			resized.Set(x, y, color.RGBA{R: rf, G: gf, B: bf, A: af})
		}
	}

	return resized
}

// Resizes to fit a bounding box using bicubic interpolation
func resizeToFitBicubic(original image.Image,
	edgeSize int) image.Image {

	oldBounds := original.Bounds()
	if oldBounds.Dx() < edgeSize && oldBounds.Dy() < edgeSize {
		// No resize necessary
		return original
	}

	newBounds := getNewBounds(oldBounds, edgeSize)

	resized := image.NewRGBA(newBounds)

	var ratioX = float64(oldBounds.Dx()) / float64(newBounds.Dx())
	var ratioY = float64(oldBounds.Dy()) / float64(newBounds.Dy())

	for x := 0; x < newBounds.Dx(); x++ {
		for y := 0; y < newBounds.Dy(); y++ {
			sourceX := ratioX * float64(x)
			minX := int(math.Floor(sourceX))
			tx := sourceX - float64(minX)

			sourceY := ratioY * float64(y)
			minY := int(math.Floor(sourceY))
			ty := sourceY - float64(minY)

			xCoeffs := bicubicCoefficients(tx)
			yCoeffs := bicubicCoefficients(ty)

			var rgba = [4]float64{0.0, 0.0, 0.0, 0.0}
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					ii := minX - 1 + i
					jj := minY - 1 + j
					rPix, gPix, bPix, aPix := quantizeColorRGBA(
						original.At(clampBounds(ii, jj, oldBounds)))
					rgba[0] += float64(rPix) * xCoeffs[i] * yCoeffs[j]
					rgba[1] += float64(gPix) * xCoeffs[i] * yCoeffs[j]
					rgba[2] += float64(bPix) * xCoeffs[i] * yCoeffs[j]
					rgba[3] += float64(aPix) * xCoeffs[i] * yCoeffs[j]
				}
			}

			rf := uint8(clampRangeFloat(0.0, math.Floor(rgba[0]), 255.0))
			gf := uint8(clampRangeFloat(0.0, math.Floor(rgba[1]), 255.0))
			bf := uint8(clampRangeFloat(0.0, math.Floor(rgba[2]), 255.0))
			af := uint8(clampRangeFloat(0.0, math.Floor(rgba[3]), 255.0))

			resized.Set(x, y, color.RGBA{R: rf, G: gf, B: bf, A: af})
		}
	}

	return resized
}

var gaussian3 = [3][3]float64{
	[3]float64{0.0625, 0.1250, 0.0625},
	[3]float64{0.1250, 0.2500, 0.1250},
	[3]float64{0.0625, 0.1250, 0.0625},
}

// Gaussian filter
func blurImage(original image.Image) image.Image {
	blurred := image.NewRGBA(original.Bounds())

	for x := 0; x < blurred.Bounds().Dx(); x++ {
		for y := 0; y < blurred.Bounds().Dy(); y++ {
			var r = float64(0)
			var g = float64(0)
			var b = float64(0)
			var a = float64(0)
			for i := 0; i < len(gaussian3); i++ {
				for j := 0; j < len(gaussian3[i]); j++ {
					rs, gs, bs, as := quantizeColorRGBA(original.At(
						clampBounds(x+i-1, y+j-1, original.Bounds())))
					r += gaussian3[i][j] * rs
					g += gaussian3[i][j] * gs
					b += gaussian3[i][j] * bs
					a += gaussian3[i][j] * as
				}
			}

			rf := uint8(clampRangeFloat(0.0, math.Floor(r), 255.0))
			gf := uint8(clampRangeFloat(0.0, math.Floor(g), 255.0))
			bf := uint8(clampRangeFloat(0.0, math.Floor(b), 255.0))
			af := uint8(clampRangeFloat(0.0, math.Floor(a), 255.0))

			blurred.Set(x, y, color.RGBA{R: rf, G: gf, B: bf, A: af})
		}
	}
	return blurred
}

func getNewBounds(original image.Rectangle, edgeSize int) image.Rectangle {
	oDims := original

	var newX, newY int

	if oDims.Dx() > oDims.Dy() {
		newX = edgeSize
		newY = oDims.Dy() * newX / oDims.Dx()
	} else {
		newY = edgeSize
		newX = oDims.Dx() * newY / oDims.Dy()
	}

	return image.Rect(0, 0, newX, newY)
}

func reorientImage(original image.Image, currentOrientation int) image.Image {
	switch currentOrientation {
	case 1:
		return original
	case 2:
		return flipHorizontal(original)
	case 3:
		return rotate(original, 180)
	case 4:
		return flipVertical(original)
	case 5:
		return flipHorizontal(rotate(original, 90))
	case 6:
		return rotate(original, 90)
	case 7:
		return rotate(flipHorizontal(original), 270)
	case 8:
		return rotate(original, 270)
	default:
		return original
	}
}

func flipHorizontal(original image.Image) image.Image {
	flipped := image.NewRGBA(original.Bounds())
	maxCoords := original.Bounds().Max
	for x := 0; x < maxCoords.X; x++ {
		for y := 0; y < maxCoords.Y; y++ {
			flipped.Set(x, y, original.At(maxCoords.X-1-x, y))
		}
	}
	return flipped
}

func flipVertical(original image.Image) image.Image {
	flipped := image.NewRGBA(original.Bounds())
	maxCoords := original.Bounds().Max
	for x := 0; x < maxCoords.X; x++ {
		for y := 0; y < maxCoords.Y; y++ {
			flipped.Set(x, y, original.At(x, maxCoords.Y-1-y))
		}
	}
	return flipped
}

func rotate(original image.Image, angle int) image.Image {
	if angle%90 != 0 {
		return original
	}
	angle = angle % 360

	originalBounds := original.Bounds()
	newBounds := getRotationBounds(originalBounds, angle)
	rotated := image.NewRGBA(newBounds)

	for x := 0; x < newBounds.Dx(); x++ {
		for y := 0; y < newBounds.Dy(); y++ {
			switch angle {
			case 90:
				rotated.Set(x, y, original.At(y,
					originalBounds.Dy()-1-x))
			case 180:
				rotated.Set(x, y, original.At(originalBounds.Dx()-1-x,
					originalBounds.Dy()-1-y))
			case 270:
				rotated.Set(x, y, original.At(newBounds.Dy()-1-y,
					x))
			}
		}
	}
	return rotated
}

func getRotationBounds(original image.Rectangle, angle int) image.Rectangle {
	if angle%180 == 0 {
		return original
	}
	maxCoords := original.Bounds().Max
	return image.Rect(0, 0, maxCoords.Y, maxCoords.X)
}

func bicubicCoefficients(dist float64) [4]float64 {
	var coeffs [4]float64
	distSquare := dist * dist
	distCube := distSquare * dist

	coeffs[0] = (-distCube + 2.0*distSquare - dist)
	coeffs[1] = (distCube - 2.0*distSquare + 1.0)
	coeffs[2] = (-distCube + distSquare + dist)
	coeffs[3] = (distCube - distSquare)

	sum := coeffs[0] + coeffs[1] + coeffs[2] + coeffs[3]
	coeffs[0] /= sum
	coeffs[1] /= sum
	coeffs[2] /= sum
	coeffs[3] /= sum

	return coeffs
}

func clampBounds(x, y int, bounds image.Rectangle) (int, int) {
	if x < bounds.Min.X {
		x = bounds.Min.X
	} else if x > bounds.Max.X {
		x = bounds.Max.X
	}

	if y < bounds.Min.Y {
		y = bounds.Min.Y
	} else if y > bounds.Max.Y {
		y = bounds.Max.Y
	}

	return x, y
}

func clampRangeUint(lo, val, hi uint32) uint32 {
	if val > hi {
		return hi
	} else if val < lo {
		return lo
	} else {
		return val
	}
}

func clampRangeFloat(lo, val, hi float64) float64 {
	return math.Max(math.Min(val, hi), lo)
}

func quantizeColorRGBA(c color.Color) (float64, float64, float64, float64) {
	r, g, b, a := c.RGBA()
	return float64(r / (0xFFFF / 0xFF)),
		float64(g / (0xFFFF / 0xFF)),
		float64(b / (0xFFFF / 0xFF)),
		float64(a / (0xFFFF / 0xFF))
}

func normalize(x0, x1, x2, x3 float64) (float64, float64, float64, float64) {
	sum := x0 + x1 + x2 + x3
	return x0 / sum, x1 / sum, x2 / sum, x3 / sum
}

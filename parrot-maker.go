package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nfnt/resize"
)

var offsets = [2]int{-25, -15}
var size = 100

var positions = [][2]int{
	{55, 35},
	{40, 28},
	{25, 32},
	{15, 39},
	{12, 39},
	{19, 43},
	{35, 47},
	{48, 51},
	{54, 47},
	{64, 42},
}

var intermediatesDir = "intermediates"

func main() {
	inputImage, err := readImage("input.png")
	if err != nil {
		fmt.Println("Error reading new image:", err)
		return
	}

	inputImage = resizeImage(inputImage, size)

	// err = writeImage(filepath.Join(intermediatesDir, "resized.png"), inputImage)
	// if err != nil {
	// 	fmt.Println("Error creating new image file:", err)
	// 	return
	// }

	framesDir := "frames"
	frames, err := os.ReadDir(framesDir)
	if err != nil {
		fmt.Println("Error reading frames directory:", err)
		return
	}

	for index, frame := range frames {
		frameImage, err := readImage(filepath.Join(framesDir, frame.Name()))
		if err != nil {
			fmt.Println("Error opening frame:", err)
			continue
		}

		intermediateFrame := overlayImages(frameImage, inputImage, positions[index][0]+offsets[0], positions[index][1]+offsets[1])

		err = writeImage(filepath.Join(intermediatesDir, fmt.Sprintf("%d.png", index)), intermediateFrame)
		if err != nil {
			fmt.Println("Error creating intermediateFrame file:", err)
			continue
		}
	}

	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		generateGif()
		return
	}

	cmd := exec.Command("rm", "-rf", "output/animation.gif")
	cmd.Dir = "."
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error deleting last ouput:", err)
		return
	}

	cmd = exec.Command("rm", "-rf", filepath.Join(intermediatesDir, "palette.png"))
	cmd.Dir = "."
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error deleting last ouput:", err)
		return
	}

	cmd = exec.Command("ffmpeg", "-i", "intermediates/%d.png", "-vf", "palettegen=reserve_transparent=1", filepath.Join(intermediatesDir, "palette.png"))
	cmd.Dir = "."
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error creating palette gif using ffmpeg:", err)
		return
	}

	cmd = exec.Command("ffmpeg", "-framerate", "20", "-i", "intermediates/%d.png", "-i", filepath.Join(intermediatesDir, "palette.png"), "-lavfi", "paletteuse=alpha_threshold=128", "-gifflags", "-offsetting", "output/animation.gif")
	cmd.Dir = "."
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error creating gif animation using ffmpeg:", err)
		return
	}

}

func generateGif() {
	// fix palette: gif trasparency
	var gifpalette color.Palette = palette.Plan9
	gifpalette[0] = color.Transparent
	// var gifpalette color.Palette = palette.WebSafe
	// gifpalette = append(gifpalette, color.Transparent)

	animationFile, err := os.Create(filepath.Join("output", "animation.gif"))
	if err != nil {
		fmt.Println("Error creating animation file:", err)
		return
	}
	defer animationFile.Close()

	animation := gif.GIF{}
	for index := 0; index < len(positions); index++ {
		frameImage, err := readImage(filepath.Join(intermediatesDir, fmt.Sprintf("%d.png", index)))
		if err != nil {
			fmt.Println("Error opening intermediate frame:", err)
			continue
		}

		// Convert frameImage to *image.Paletted
		palettedImage := image.NewPaletted(frameImage.Bounds(), gifpalette)
		draw.Draw(palettedImage, palettedImage.Rect, frameImage, frameImage.Bounds().Min, draw.Src)

		animation.Image = append(animation.Image, palettedImage)
		animation.Disposal = append(animation.Disposal, gif.DisposalPrevious)
		animation.Delay = append(animation.Delay, 5)

	}

	err = gif.EncodeAll(animationFile, &animation)
	if err != nil {
		fmt.Println("Error encoding gif animation:", err)
		return
	}
}

func readImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func writeImage(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return err
	}
	return nil
}

func resizeImage(img image.Image, width int) image.Image {
	aspectRatio := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
	newWidth := width
	newHeight := int(float64(newWidth) / aspectRatio)
	return resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)
}

func overlayImages(background, overlay image.Image, x, y int) image.Image {
	offset := image.Pt(x, y)
	b := background.Bounds()
	m := image.NewRGBA(b)
	draw.Draw(m, b, background, image.Point{}, draw.Src)
	draw.Draw(m, overlay.Bounds().Add(offset), overlay, image.Point{}, draw.Over)
	return m
}

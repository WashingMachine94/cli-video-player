package main

import (
	"fmt"
	"math"
)

// Preprocesses a frame and converts to ASCII
func processFrame(frameptr *Frame, width int, height int, channels int) *string {
	asciiFrame := frameToAscii(frameptr, width, height, channels, DEFAULT_ASCII)
	// blurredFrame := gaussianBlur(width, height, frameptr, 1, 2)
	// blurredFrame1 := gaussianBlur(width, height, frameptr, 2, 4)
	// edgeFrame := subtractFrame(&blurredFrame, &blurredFrame1)
	// edgeOverlay := frameToAscii(edgeFrame, width, height, channels, EDGE_ASCII)
	// screen := addString(asciiFrame, edgeOverlay)
	return asciiFrame
}

func frameToAscii(frameptr *Frame, width int, height int, channels int, characters string) *string {
	frame := *frameptr

	// TODO: Preserve aspect ratio.

	var frameWidth = TERMINAL_WIDTH
	var frameHeight = TERMINAL_HEIGHT - 3
	var gamma float64 = 0.8

	var pixelWidth float32 = float32(width) / float32(frameWidth)
	var pixelHeight float32 = float32(height) / float32(frameHeight)

	var screen string = ""

	for row := 0; row < frameHeight; row++ {
		for col := 0; col < frameWidth; col++ {
			var x int = int(pixelWidth * float32(col))
			var y int = int(pixelHeight * float32(row))

			var brightnessSum int
			for pixelRow := 0; pixelRow < int(pixelHeight); pixelRow++ {
				for pixelCol := 0; pixelCol < int(pixelWidth); pixelCol++ {
					var localIndex int = (((y + pixelRow) * width) + x + pixelCol) * channels
					brightnessSum += int(frame[localIndex])
				}
			}

			var averageBrightness float64 = float64(brightnessSum / (int(pixelHeight) * int(pixelWidth)))
			var normalizedBrightness float64 = averageBrightness / 255.0
			var gammaCorrectedBrightness = math.Pow(normalizedBrightness, gamma)

			var charIndex = int((1 - gammaCorrectedBrightness) * float64(len(characters)-1))
			screen += string(characters[charIndex])
		}
	}
	return &screen
}

func subtractFrame(firstFrame *Frame, secondFrame *Frame) *Frame {
	if len(*firstFrame) != len(*secondFrame) {
		fmt.Errorf("subractFrame(): Frames are not of the same size")
	}

	var frameDiff Frame

	for i := 0; i < len(*firstFrame); i++ {
		value := (*firstFrame)[i] - (*secondFrame)[i]
		if value < 10 {
			frameDiff = append(frameDiff, value)
			continue
		}
		frameDiff = append(frameDiff, 255)
	}

	return &frameDiff
}

// overlays the second string on top of the first (space characters are ignored)
func addString(firstFrame *string, secondFrame *string) *string {
	if len(*firstFrame) != len(*secondFrame) {
		fmt.Errorf("addString(): Strings are not of the same size")
	}

	var frameSum string

	for i := 0; i < len(*firstFrame); i++ {
		if (*secondFrame)[i] == ' ' {
			frameSum += string((*firstFrame)[i])
			continue
		}
		frameSum += string((*secondFrame)[i])
	}

	return &frameSum
}

func printFrame(frame *string) {
	fmt.Print("\033[K\033[1G" + gotoCharacter(0, 0) + HELP_MENU + gotoCharacter(0, 1) + *frame)
}

func generateGaussianKernel(radius int, sigma float64) [][]float64 {
	size := 2*radius + 1
	kernel := make([][]float64, size)

	var sum float64

	// Calculate each value in the kernel
	for i := 0; i < size; i++ {
		kernel[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			x := float64(i - radius)
			y := float64(j - radius)
			kernel[i][j] = (1 / (2 * math.Pi * sigma * sigma)) * math.Exp(-(x*x+y*y)/(2*sigma*sigma))
			sum += kernel[i][j]
		}
	}

	// Normalize the kernel to sum of 1
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			kernel[i][j] /= sum
		}
	}

	return kernel
}

func gaussianBlur(width int, height int, frameptr *Frame, sigma float64, radius int) Frame {
	var kernel [][]float64 = generateGaussianKernel(radius, sigma)
	var blurredFrame Frame
	frame := *frameptr

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			var blurredPixel float64

			for kernelX := -radius; kernelX < radius; kernelX++ {
				for kernelY := -radius; kernelY < radius; kernelY++ {
					var combinedX int = x + kernelX
					var combinedY int = y + kernelY

					if combinedX < 0 || combinedX >= width {
						continue
					}
					if combinedY < 0 || combinedY >= height {
						continue
					}
					var index int = combinedX + combinedY*width
					blurredPixel += float64(frame[index]) * kernel[kernelX+radius][kernelY+radius]
				}
			}
			blurredFrame = append(blurredFrame, byte(int(blurredPixel)))
		}
	}

	return blurredFrame
}

// Gets the difference between 2 ASCII frames.
// Result also contains escape characters to move cursor to right locations.
// This results in having to print less characters to the screen.
func getFrameDiff(oldFramePtr *string, newFramePtr *string) string {
	oldFrame := *oldFramePtr
	newFrame := *newFramePtr

	var diff string = ""
	var prevCharEqual bool = true

	for char := 0; char < len(oldFrame); char++ {
		if oldFrame[char] != newFrame[char] {
			if prevCharEqual {
				currentLine := int(char / TERMINAL_WIDTH)
				currentChar := int(char % TERMINAL_WIDTH)
				diff += gotoCharacter(currentChar+1, currentLine+1)
			}
			diff += string(newFrame[char])
			prevCharEqual = false
			continue
		}
		prevCharEqual = true
	}

	return diff
}

func gotoCharacter(x int, y int) string {
	return fmt.Sprintf("\033[%d;%dH", y+1, x)
}

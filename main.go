package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	YELLOW_COLOR string = "\033[33m"
	BLUE_COLOR   string = "\033[34m"
	CYAN_COLOR   string = "\033[36m"
	RESET_COLOR  string = "\033[0m"
	ORANGE_COLOR string = "\033[38;5;208m"
	RED_COLOR    string = "\033[31m"
)
const (
	BUTTON_BACK_H    = ORANGE_COLOR + "[" + YELLOW_COLOR + "<" + ORANGE_COLOR + "]" + RESET_COLOR
	BUTTON_BACK      = "[<]"
	BUTTON_FORWARD_H = ORANGE_COLOR + "[" + YELLOW_COLOR + ">" + ORANGE_COLOR + "]" + RESET_COLOR
	BUTTON_FORWARD   = "[>]"
	BUTTON_PLAYING   = "  ||  "
	BUTTON_PAUSED    = YELLOW_COLOR + "  ||  " + RESET_COLOR
	HELP_MENU        = "[q]uit  pause[spacebar, k]  backwards[j, <]  forward[l, >]  goto[0-9]"
)

const PREFIX_TEXT = "VideoPlayer:"
const PREFIX string = YELLOW_COLOR + PREFIX_TEXT + RESET_COLOR
const BUFFER_SIZE int = 15
const BUFFER_OFFSET int = 30
const SKIP_AMOUNT_S int = 10

var CURRENT_VIDEO Video
var TERMINAL_WIDTH int
var TERMINAL_HEIGHT int
var PLAYING bool
var PAUSED bool
var DIM_CHANGE_DURING_PAUSE bool = false
var SKIP_BACKWARD bool = false
var SKIP_FORWARD bool = false
var GOTO bool = false
var GOTOPOS int = 0

func main() {
	if len(os.Args) < 2 {
		fmt.Println()
		fmt.Println(PREFIX, "Run 'play <video_path>' to play a video,")
		fmt.Println(strings.Repeat(" ", len(PREFIX_TEXT)), "for example: 'play video.mp4'.")
		return
	}

	if os.Args[1] == "test" {
		runTests(os.Args[2])
		return
	}

	if _, err := os.Stat(os.Args[1]); err != nil {
		fmt.Println(PREFIX, "File '"+os.Args[1]+"' could not be found.")
		return
	}
	CURRENT_VIDEO = loadVideo(os.Args[1], BUFFER_OFFSET*2)
	if CURRENT_VIDEO.fps == 0 {
		fmt.Println(PREFIX, "'"+os.Args[1]+"' is not a valid video.")
		return
	}
	playVideo()
}

func playVideo() {
	bufferVideo(&CURRENT_VIDEO, 0, BUFFER_OFFSET)
	setTerminalDimensions()
	go handleInput()
	PLAYING = true
	PAUSED = false

	frame, _ := getFrame(&CURRENT_VIDEO)
	oldFrame := processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
	printFrame(oldFrame)
	shiftBuffer(&CURRENT_VIDEO)
	drawMenu()

	for PLAYING {
		startFrameTime := time.Now()
		dimChanged := setTerminalDimensions()
		frame, exists := getFrame(&CURRENT_VIDEO)
		if !PAUSED {
			if exists {
				newFrame := processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
				if dimChanged || DIM_CHANGE_DURING_PAUSE {
					printFrame(newFrame)
					DIM_CHANGE_DURING_PAUSE = false
				} else {
					frameDiff := getFrameDiff(oldFrame, newFrame)
					printFrame(&frameDiff)
				}
				oldFrame = newFrame

				shiftBuffer(&CURRENT_VIDEO)
				if CURRENT_VIDEO.currentFrame%BUFFER_SIZE == 0 {
					go bufferVideo(&CURRENT_VIDEO, CURRENT_VIDEO.currentFrame+BUFFER_OFFSET, BUFFER_SIZE)
				}
				CURRENT_VIDEO.currentFrame++
			} else {
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}
		if PAUSED && dimChanged {
			frame, _ := getFrame(&CURRENT_VIDEO)
			oldFrame := processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
			printFrame(oldFrame)
			DIM_CHANGE_DURING_PAUSE = true
		}

		drawMenu()
		handleGoto()
		handleSkip()

		var deltaTime time.Duration = time.Now().Sub(startFrameTime)
		time.Sleep((time.Second / time.Duration(CURRENT_VIDEO.fps)) - deltaTime)
		if CURRENT_VIDEO.currentFrame == CURRENT_VIDEO.totalFrames {
			PLAYING = false
		}
	}
	exit()
}

func drawMenu() {
	var runtime = int(CURRENT_VIDEO.duration.Seconds())
	var currentTime = int(((time.Second / time.Duration(CURRENT_VIDEO.fps)) * time.Duration(CURRENT_VIDEO.currentFrame)).Seconds())
	currentMinutes := currentTime / 60
	currentSeconds := currentTime % 60
	endMinutes := runtime / 60
	endSeconds := runtime % 60

	var currentTimeWidth int = len(fmt.Sprintf("[%d:%02d]", currentMinutes, currentSeconds))
	var currentTimeText string = fmt.Sprintf(BLUE_COLOR+"["+CYAN_COLOR+"%d:%02d"+BLUE_COLOR+"]"+RESET_COLOR, currentMinutes, currentSeconds)
	var endTimeWidth int = len(fmt.Sprintf("[%02d:%02d]", endMinutes, endSeconds))
	var endTime string = fmt.Sprintf(BLUE_COLOR+"["+CYAN_COLOR+"%02d:%02d"+BLUE_COLOR+"]"+RESET_COLOR, endMinutes, endSeconds)

	var buttonsWidth int = 12
	var buttons string = ""
	if SKIP_BACKWARD {
		buttons += BUTTON_BACK_H
	} else {
		buttons += BUTTON_BACK
	}
	if PAUSED {
		buttons += BUTTON_PAUSED
	} else {
		buttons += BUTTON_PLAYING
	}
	if SKIP_FORWARD {
		buttons += BUTTON_FORWARD_H
	} else {
		buttons += BUTTON_FORWARD
	}
	var spacingWidth float64 = float64(TERMINAL_WIDTH-currentTimeWidth-endTimeWidth-buttonsWidth) / 2
	var oddSpacing bool = spacingWidth-math.Floor(spacingWidth) >= 0.5
	var spacing string = strings.Repeat(" ", int(spacingWidth))

	var progressProcent float64 = float64(currentTime) / float64(runtime)
	var progressChars int = int((float64(TERMINAL_WIDTH) - 2) * progressProcent)
	var progressbar string = BLUE_COLOR + "[" + CYAN_COLOR

	for i := 0; i < TERMINAL_WIDTH-2; i++ {
		if progressChars >= i {
			progressbar += "■"
		} else {
			progressbar += "□"
		}
	}
	gotoPos := fmt.Sprintf("\033[%d;0H", TERMINAL_HEIGHT-1)
	var menubar string
	if oddSpacing {
		menubar = currentTimeText + spacing + buttons + spacing + " " + endTime
	} else {
		menubar = currentTimeText + spacing + buttons + spacing + endTime
	}
	progressbar += BLUE_COLOR + "]" + RESET_COLOR

	fmt.Print(gotoPos + menubar + gotoCharacter(0, TERMINAL_HEIGHT) + progressbar + "\033[0;0H")
}

func handleInput() {
	for {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			exit()
			return
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		b := make([]byte, 1)
		r, err := os.Stdin.Read(b)

		if err != nil || r == 0 {
			continue
		}

		if b[0] == 32 || b[0] == 107 { // PAUSING: SPACEBAR, k
			PAUSED = !PAUSED
		}
		if b[0] >= 48 && b[0] <= 58 { // GOTO: 0-9
			GOTO = true
			GOTOPOS = int(b[0]) - 48
		}
		if b[0] == 106 || b[0] == 60 { // SKIP BACK: j, <
			SKIP_BACKWARD = true
		}
		if b[0] == 108 || b[0] == 62 { // SKIP FORWARD: l, >
			SKIP_FORWARD = true
		}
		if b[0] == 3 || b[0] == 113 { // EXIT: ctrl-c, q
			exit()
			return
		}
	}
}

func handleSkip() {
	if SKIP_FORWARD {
		stepForward(&CURRENT_VIDEO)
		frame, _ := getFrame(&CURRENT_VIDEO)
		previewFrame := processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
		printFrame(previewFrame)
		SKIP_FORWARD = false
	}
	if SKIP_BACKWARD {
		stepBackward(&CURRENT_VIDEO)
		frame, _ := getFrame(&CURRENT_VIDEO)
		previewFrame := processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
		printFrame(previewFrame)
		SKIP_BACKWARD = false
	}
}
func handleGoto() {
	if !GOTO {
		return
	}
	targetFrame := (CURRENT_VIDEO.totalFrames / 10) * GOTOPOS

	setFrame(&CURRENT_VIDEO, targetFrame)
	frame, _ := getFrame(&CURRENT_VIDEO)
	previewFrame := processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
	printFrame(previewFrame)
	GOTO = false

}
func exit() {
	// clear screen before exiting
	fmt.Printf("\033[0;0H")
	for i := 0; i < TERMINAL_HEIGHT; i++ {
		fmt.Println("\n")
	}
	fmt.Println(RESET_COLOR)
	os.Exit(1)
}

func setTerminalDimensions() bool {
	fd := int(os.Stdout.Fd())
	width, height, err := term.GetSize(int(fd))
	if err != nil {
		fmt.Println(PREFIX, "Error getting terminal dimensions:", err)
		exit()
	}
	widthChanged := width != TERMINAL_WIDTH
	heightChanged := height != TERMINAL_HEIGHT

	TERMINAL_WIDTH = width
	TERMINAL_HEIGHT = height
	return widthChanged || heightChanged
}

// Converts a video frame to ASCII
func processFrame(frameptr *Frame, width int, height int, channels int) *string {
	frame := *frameptr
	var characters string = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/()1{}[]?-_+~<>i!lI;:,^`'. "

	// TODO: Preserve aspect ratio.

	var frameWidth = TERMINAL_WIDTH
	var frameHeight = TERMINAL_HEIGHT - 3
	var gamma float64 = 0.7

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
					var r = int(frame[localIndex])
					var g = int(frame[localIndex+1])
					var b = int(frame[localIndex+2])
					var brightness = 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
					brightnessSum += int(brightness)
				}
			}

			var averageBrightness float64 = float64(brightnessSum / (int(pixelHeight) * int(pixelWidth) * 3))
			var normalizedBrightness float64 = averageBrightness / 255.0
			var gammaCorrectedBrightness = math.Pow(normalizedBrightness, gamma)

			var charIndex = int(gammaCorrectedBrightness * float64(len(characters)-1))
			screen += string(characters[charIndex])
		}
	}
	return &screen
}
func printFrame(frame *string) {
	fmt.Print("\033[K\033[1G" + gotoCharacter(0, 0) + HELP_MENU + gotoCharacter(0, 1) + *frame)
}

func generateGaussianKernel(sigma float64) [][]float64 {
	radius := int(math.Ceil(3 * sigma)) // Kernel size of 7 '(7 - 1) / 2 = 3'
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

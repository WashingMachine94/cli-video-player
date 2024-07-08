package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

const YELLOW_COLOR string = "\033[33m"
const RED_COLOR string = "\033[31m"
const RESET_COLOR string = "\033[0m"

const PREFIX string = YELLOW_COLOR + "VideoPlayer:" + RESET_COLOR

var CURRENT_VIDEO Video
var TERMINAL_WIDTH int
var TERMINAL_HEIGHT int
var PLAYING bool

func main() {
	if len(os.Args) < 2 {
		fmt.Println(PREFIX, "Usage: go run main.go <video_path>")
		os.Exit(1)
	}

	if _, err := os.Stat(os.Args[1]); err != nil {
		fmt.Println(PREFIX, "File '"+os.Args[1]+"' could not be found.")
		os.Exit(1)
	}

	playVideo(os.Args[1])
}

func playVideo(path string) {
	CURRENT_VIDEO = loadVideo(path)
	go bufferVideo(&CURRENT_VIDEO)

	setTerminalDimensions()

	PLAYING = true

	fmt.Println()

	for PLAYING {
		startFrameTime := time.Now()

		frame, exists := getFrame(&CURRENT_VIDEO, CURRENT_VIDEO.currentFrame)
		if exists {
			processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
			CURRENT_VIDEO.currentFrame += 2
		} else {
			// Wait for buffering if frame is not yet available
			time.Sleep(10 * time.Millisecond)
			continue
		}
		var deltaTime time.Duration = time.Now().Sub(startFrameTime)
		time.Sleep((time.Second / time.Duration(CURRENT_VIDEO.fps/2)) - deltaTime)
	}
}

func drawMenu() {
	var runtime = int(CURRENT_VIDEO.duration.Seconds())

	var currentTime = int(((time.Second / time.Duration(CURRENT_VIDEO.fps)) * time.Duration(CURRENT_VIDEO.currentFrame)).Seconds())

	startMinutes := currentTime / 60
	startSeconds := currentTime % 60
	var startTime string = fmt.Sprintf("[%d:%02d]", startMinutes, startSeconds)
	var buttons string = "<- ▶ ->"

	endMinutes := runtime / 60
	endSeconds := runtime % 60
	var endTime string = fmt.Sprintf("[%d:%02d]", endMinutes, endSeconds)
	var spacing string = strings.Repeat(" ", (TERMINAL_WIDTH-len(startTime)-len(endTime)-len(buttons))/2+1)

	var progressProcent float64 = float64(currentTime) / float64(runtime)
	var progressChars int = int((float64(TERMINAL_WIDTH) - 2) * progressProcent)
	var progressbar string = "["

	for i := 0; i < TERMINAL_WIDTH-2; i++ {
		if progressChars > i {
			progressbar += "="
		} else {
			progressbar += "·"
		}
	}
	progressbar += "]"

	fmt.Println(startTime + spacing + buttons + spacing + endTime)
	fmt.Print(progressbar)
}

func setTerminalDimensions() {
	fd := windows.Handle(windows.Stdout)
	width, height, err := term.GetSize(int(fd))
	if err != nil {
		fmt.Println(PREFIX, "Error getting terminal dimensions:", err)
		os.Exit(1)
	}
	TERMINAL_WIDTH = width
	TERMINAL_HEIGHT = height
}

func processFrame(frame []byte, width int, height int, channels int) {
	var characters string = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/()1{}[]?-_+~<>i!lI;:,^`'. "

	var frameWidth = TERMINAL_WIDTH
	var frameHeight = TERMINAL_HEIGHT - 3

	var pixelWidth int = int(float32(width) / float32(frameWidth))
	var pixelHeight int = int(float32(height) / float32(frameHeight))

	var screen string = ""

	for row := 0; row < frameHeight; row++ {
		screen += "\n"
		for col := 0; col < frameWidth; col++ {
			var x int = pixelWidth * col
			var y int = pixelHeight * row

			var brightness int

			for pixelRow := 0; pixelRow < pixelHeight; pixelRow++ {
				for pixelCol := 0; pixelCol < pixelWidth; pixelCol++ {
					var localIndex int = (((y + pixelRow) * width) + x + pixelCol) * channels
					brightness += int(frame[localIndex+1])
				}
			}
			brightness /= (pixelHeight * pixelWidth * channels)
			var charIndex int
			if brightness == 0 {
				charIndex = len(characters) - 1
			} else {
				charIndex = (len(characters) - 1) / brightness
			}
			screen += string(characters[charIndex])
		}
	}
	fmt.Printf("\033[0;0H")
	fmt.Print(screen)
	os.Stdout.Sync()
	drawMenu()
}

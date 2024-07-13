package main

import (
	"fmt"
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
)
const (
	BUTTON_BACK_H    = ORANGE_COLOR + "[" + YELLOW_COLOR + "<" + ORANGE_COLOR + "]" + RESET_COLOR
	BUTTON_BACK      = "[<]"
	BUTTON_FORWARD_H = ORANGE_COLOR + "[" + YELLOW_COLOR + ">" + ORANGE_COLOR + "]" + RESET_COLOR
	BUTTON_FORWARD   = "[>]"
	BUTTON_PLAYING   = "  ||  "
	BUTTON_PAUSED    = "  ▶  "
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
var SKIP_BACKWARD bool = false
var SKIP_FORWARD bool = false

func main() {
	if len(os.Args) != 2 {
		fmt.Println()
		fmt.Println(PREFIX, "Run 'play <video_path>' to play a video,")
		fmt.Println(strings.Repeat(" ", len(PREFIX_TEXT)), "for example: 'play video.mp4'.")
		return
	}

	if _, err := os.Stat(os.Args[1]); err != nil {
		fmt.Println(PREFIX, "File '"+os.Args[1]+"' could not be found.")
		return
	}

	playVideo(os.Args[1])
}

func playVideo(path string) {
	CURRENT_VIDEO = loadVideo(path, BUFFER_OFFSET*2)
	if CURRENT_VIDEO.fps == 0 {
		fmt.Println(PREFIX, "'"+path+"' is not a valid video.")
		return
	}
	bufferVideo(&CURRENT_VIDEO, 0, BUFFER_OFFSET)
	setTerminalDimensions()
	go handleInput()

	PLAYING = true
	PAUSED = false

	drawMenu()

	for PLAYING {
		startFrameTime := time.Now()

		frame, exists := getFrame(&CURRENT_VIDEO)
		if !PAUSED {
			if exists {
				setTerminalDimensions()
				processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
				shiftBuffer(&CURRENT_VIDEO)
				if CURRENT_VIDEO.currentFrame%BUFFER_SIZE == 0 {
					go bufferVideo(&CURRENT_VIDEO, CURRENT_VIDEO.currentFrame+BUFFER_OFFSET, BUFFER_SIZE)
				}
				CURRENT_VIDEO.currentFrame++
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		}

		drawMenu()
		handleSkip()

		var deltaTime time.Duration = time.Now().Sub(startFrameTime)
		time.Sleep((time.Second / time.Duration(CURRENT_VIDEO.fps)) - deltaTime)
		if CURRENT_VIDEO.currentFrame == CURRENT_VIDEO.totalFrames {
			PLAYING = false
		}
	}
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
	var endTimeWidth int = len(fmt.Sprintf("[%d:%02d]", endMinutes, endSeconds))
	var endTime string = fmt.Sprintf(BLUE_COLOR+"["+CYAN_COLOR+"%d:%02d"+BLUE_COLOR+"]"+RESET_COLOR, endMinutes, endSeconds)

	var buttonsWidth int = 0
	var buttons string = ""
	if SKIP_BACKWARD {
		buttons += BUTTON_BACK_H
	} else {
		buttons += BUTTON_BACK
	}
	if PAUSED {
		buttonsWidth = 11
		buttons += BUTTON_PAUSED
	} else {
		buttonsWidth = 12
		buttons += BUTTON_PLAYING
	}
	if SKIP_FORWARD {
		buttons += BUTTON_FORWARD_H
	} else {
		buttons += BUTTON_FORWARD
	}

	var spacing string = strings.Repeat(" ", (TERMINAL_WIDTH-currentTimeWidth-endTimeWidth-buttonsWidth)/2)

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
	progressbar += BLUE_COLOR + "]" + RESET_COLOR

	fmt.Printf("\033[%d;0H", TERMINAL_HEIGHT-1)
	fmt.Println(currentTimeText + spacing + buttons + spacing + endTime)
	fmt.Print(progressbar)
	fmt.Printf("\033[0;0H")
}

func handleInput() {
	for {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
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
		if b[0] == 106 || b[0] == 60 { // SKIP BACK: j, <
			SKIP_BACKWARD = true
		}
		if b[0] == 108 || b[0] == 62 { // SKIP FORWARD: l, >
			SKIP_FORWARD = true
		}
		if b[0] == 3 || b[0] == 113 { // EXIT: ctrl-c, q
			fmt.Printf("\033[0;0H")
			for i := 0; i < TERMINAL_HEIGHT; i++ {
				fmt.Println("\n")
			}
			os.Exit(1)
		}
	}
}

func handleSkip() {
	if SKIP_FORWARD {
		stepForward(&CURRENT_VIDEO)
		frame, _ := getFrame(&CURRENT_VIDEO)
		processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
		SKIP_FORWARD = false
	}
	if SKIP_BACKWARD {
		stepBackward(&CURRENT_VIDEO)
		frame, _ := getFrame(&CURRENT_VIDEO)
		processFrame(frame, CURRENT_VIDEO.width, CURRENT_VIDEO.height, 3)
		SKIP_BACKWARD = false
	}
}

func setTerminalDimensions() {
	fd := int(os.Stdout.Fd())
	width, height, err := term.GetSize(int(fd))
	if err != nil {
		fmt.Println(PREFIX, "Error getting terminal dimensions:", err)
		os.Exit(1)
	}
	TERMINAL_WIDTH = width
	TERMINAL_HEIGHT = height
}

func processFrame(frameptr *Frame, width int, height int, channels int) {
	frame := *frameptr
	var characters string = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/()1{}[]?-_+~<>i!lI;:,^`'. "

	var frameWidth = TERMINAL_WIDTH
	var frameHeight = TERMINAL_HEIGHT - 3

	var pixelWidth float32 = float32(width) / float32(frameWidth)
	var pixelHeight float32 = float32(height) / float32(frameHeight)

	// var screen string = "[q]uit [h]elp"
	var screen string = "[q]uit pause:[spacebar, k], backwards: [j, <]. forward: [l, >]"

	screen += "\n"
	for row := 0; row < frameHeight; row++ {
		for col := 0; col < frameWidth; col++ {
			var x int = int(pixelWidth * float32(col))
			var y int = int(pixelHeight * float32(row))

			var brightness int
			for pixelRow := 0; pixelRow < int(pixelHeight); pixelRow++ {
				for pixelCol := 0; pixelCol < int(pixelWidth); pixelCol++ {
					var localIndex int = (((y + pixelRow) * width) + x + pixelCol) * channels
					brightness += int(frame[localIndex+1])
				}
			}

			brightness /= int(pixelHeight) * int(pixelWidth) * channels
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
	fmt.Print(screen, RESET_COLOR)
	os.Stdout.Sync()
}

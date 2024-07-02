package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

const PREFIX string = "VideoPlayer:"

func main() {

	switch os.Args[1] {
	case "play":
		if _, err := os.Stat(os.Args[2]); err != nil {
			fmt.Println(PREFIX, "File '"+os.Args[2]+"' could not be found.")
			os.Exit(1)
		}
		playVideo(os.Args[2])
	default:
		fmt.Println(PREFIX, "Unknown command 'Skill Issue'.")
		os.Exit(1)
	}
}

func playVideo(path string) {
	cmd := exec.Command("./ffmpeg", "-i", path, "-f", "image2pipe", "-vcodec", "rawvideo", "-pix_fmt", "rgb24", "-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg command: %v", err)
	}

	buf := new(bytes.Buffer)
	done := make(chan error, 1)

	go func() {
		_, err := io.Copy(buf, stdout)
		done <- err
	}()

	width, height, err := getVideoDimensions(path)
	if err != nil {
		log.Fatalf("Failed to get video dimensions: %v", err)
	}
	frameSize := width * height * 3

	for {
		if buf.Len() >= frameSize {
			frame := buf.Next(frameSize)
			processFrame(frame, width, height, 3)
			time.Sleep(time.Second / 30)

		} else {
			select {
			case err := <-done:
				if err != nil {
					log.Fatalf("Failed to read FFmpeg output: %v", err)
				}
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func getVideoDimensions(path string) (int, int, error) {
	// FFmpeg command to get video stream information
	cmd := exec.Command("./ffmpeg", "-i", path)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		// FFmpeg writes the information to stderr
		output := stderr.String()

		// Use regex to extract width and height
		re := regexp.MustCompile(`, (\d+)x(\d+)[, ]`)
		matches := re.FindStringSubmatch(output)
		if len(matches) == 3 {
			width, err := strconv.Atoi(matches[1])
			if err != nil {
				return 0, 0, fmt.Errorf("failed to parse width: %v", err)
			}
			height, err := strconv.Atoi(matches[2])
			if err != nil {
				return 0, 0, fmt.Errorf("failed to parse height: %v", err)
			}
			return width, height, nil
		}
		return 0, 0, fmt.Errorf("failed to find video dimensions in output: %s", output)
	}
	return 0, 0, fmt.Errorf("ffmpeg command failed")
}

func processFrame(frame []byte, width int, height int, channels int) {
	var characters string = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/()1{}[]?-_+~<>i!lI;:,^`'. "

	var maxWidth int = 160
	var maxHeight int = 45
	// var maxWidth int = 100
	// var maxHeight int = 32
	var pixelWidth int = int(float32(width) / float32(maxWidth))
	var pixelHeight int = int(float32(height) / float32(maxHeight))

	var screen string = ""

	for row := range maxHeight {
		for col := range maxWidth {
			var x int = pixelWidth * col
			var y int = pixelHeight * row

			var brightness int

			for pixelRow := range pixelHeight {
				for pixelCol := range pixelWidth {
					var localIndex int = (((y + pixelRow) * width) + x + pixelCol) * channels
					brightness += int(frame[localIndex+1])
					// for color := range channels {
					// 	brightness += int(frame[localIndex+color])
					// }
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
		screen += "\n"
	}
	clearTerminal()
	fmt.Println(screen)

}

func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Run()
}
func clearTerminal() {
	switch runtime.GOOS {
	case "darwin":
		runCmd("clear")
	case "linux":
		runCmd("clear")
	case "windows":
		runCmd("cmd", "/c", "cls")
	default:
		runCmd("clear")
	}
}

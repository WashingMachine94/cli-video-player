package main

import (
	"bytes"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Video struct {
	filepath     string
	duration     time.Duration
	width        int
	height       int
	fps          float64
	currentFrame int
}

func loadVideo(filepath string) Video {
	// FFmpeg get video stream information
	cmd := exec.Command("./ffmpeg", "-i", filepath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var duration time.Duration
	var fps float64
	var width, height int

	if err := cmd.Run(); err != nil {
		output := stderr.String()

		// Use regex to extract video information
		re := regexp.MustCompile(`, (\d+)x(\d+)[, ]`)
		matches := re.FindStringSubmatch(output)
		width, _ = strconv.Atoi(matches[1])
		height, _ = strconv.Atoi(matches[2])

		re = regexp.MustCompile(`\b(\d+)\s*fps\b`)
		matches = re.FindStringSubmatch(output)
		fps, _ = strconv.ParseFloat(matches[1], 64)

		var durationStr string
		re = regexp.MustCompile(`Duration:\s+(\d{2}:\d{2}:\d{2}\.\d{2})`)
		matches = re.FindStringSubmatch(output)
		durationStr = matches[1]

		timeParts := strings.Split(durationStr, ":")
		hours, _ := strconv.Atoi(timeParts[0])
		minutes, _ := strconv.Atoi(timeParts[1])
		secondsAndMs := timeParts[2]

		secParts := strings.Split(secondsAndMs, ".")
		seconds, _ := strconv.Atoi(secParts[0])
		milliseconds, _ := strconv.Atoi(secParts[1])

		totalMilliseconds := (hours*3600+minutes*60+seconds)*1000 + milliseconds
		duration = time.Duration(totalMilliseconds) * time.Millisecond

		return Video{filepath, duration, int(width), int(height), fps, 0}
	}
	return Video{}
}

func play(video *Video) {
	cmd := exec.Command("./ffmpeg", "-i", video.filepath, "-f", "image2pipe", "-vcodec", "rawvideo", "-pix_fmt", "rgb24", "-")

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

	if err != nil {
		log.Fatalf("Failed to get video dimensions: %v", err)
	}
	frameSize := video.width * video.height * 3

	for {
		if buf.Len() >= frameSize {
			frame := buf.Next(frameSize)
			processFrame(frame, video.width, video.height, 3)
			time.Sleep(time.Second / time.Duration(video.fps))
			video.currentFrame++

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

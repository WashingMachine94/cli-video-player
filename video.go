package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Video struct {
	filepath       string
	duration       time.Duration
	width          int
	height         int
	fps            float64
	currentFrame   int
	bufferedFrames map[int][]byte
	bufferMutex    sync.Mutex
	bufferComplete bool
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

		return Video{
			filepath:       filepath,
			duration:       duration,
			width:          int(width),
			height:         int(height),
			fps:            fps,
			bufferedFrames: make(map[int][]byte),
		}
	}
	return Video{}
}

func bufferVideo(video *Video) {
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
	var bufferFrame int = 0

	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				log.Fatalf("Failed to copy data from ffmpeg: %v", err)
			}
			// Process remaining buffered data if any
			for buf.Len() >= frameSize {
				frame := buf.Next(frameSize)
				video.bufferMutex.Lock()
				video.bufferedFrames[bufferFrame] = frame
				bufferFrame++
				video.bufferMutex.Unlock()
			}
			video.bufferComplete = true
			log.Println("Buffering complete")
			return
		default:
			if buf.Len() >= frameSize {
				frame := buf.Next(frameSize)
				video.bufferMutex.Lock()
				video.bufferedFrames[bufferFrame] = frame
				if bufferFrame == 150 {
					fmt.Println(video.bufferedFrames[bufferFrame])

					video.bufferMutex.Unlock()
					return
				}
				bufferFrame++
				video.bufferMutex.Unlock()
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func getFrame(video *Video, frameNumber int) ([]byte, bool) {
	video.bufferMutex.Lock()
	// defer video.bufferMutex.Unlock()
	frame, exists := video.bufferedFrames[frameNumber]
	video.bufferMutex.Unlock()
	return frame, exists
}

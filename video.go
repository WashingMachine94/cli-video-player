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

type Frame []byte

type Video struct {
	filepath       string
	duration       time.Duration
	width          int
	height         int
	fps            float64
	totalFrames    int
	currentFrame   int
	frameBuffer    []Frame
	bufferMutex    sync.Mutex
	bufferComplete bool
}

func loadVideo(filepath string, maxBufferLen int) Video {
	// FFmpeg get video stream information
	cmd := exec.Command("ffmpeg", "-i", filepath)

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
		// if fps isnt found its likely not a video
		if len(matches) == 0 {
			return Video{}
		}
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

		totalFrames := duration.Seconds() * fps

		return Video{
			filepath:    filepath,
			duration:    duration,
			width:       int(width),
			height:      int(height),
			totalFrames: int(totalFrames),
			fps:         fps,
			frameBuffer: make([]Frame, 0, maxBufferLen),
		}
	}
	return Video{}
}

func bufferVideo(video *Video, startFrame int, frameAmount int) {
	// Construct ffmpeg command
	args := []string{
		"-ss", fmt.Sprintf("%.6f", float64(startFrame)/video.fps),
		"-i", video.filepath,
		"-frames:v", strconv.Itoa(frameAmount),
		"-vf", fmt.Sprintf("fps=%.5f,format=gray", video.fps),
		"-f", "image2pipe",
		"-vcodec", "rawvideo",
		"-pix_fmt", "gray",
		"-vsync", "vfr",
		"-",
	}
	cmd := exec.Command("ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg command: %v", err)
	}

	frameSize := video.width * video.height * CHANNELS
	done := make(chan error, 1)

	go func() {
		buf := make(Frame, frameSize)
		for {
			n, err := io.ReadFull(stdout, buf)
			if err == io.EOF {
				done <- nil
				break
			}
			if err != nil {
				done <- err
				break
			}
			if n == frameSize {
				var frame Frame = make(Frame, frameSize)
				copy(frame, buf)

				video.bufferMutex.Lock()
				video.frameBuffer = append(video.frameBuffer, frame)
				video.bufferMutex.Unlock()
			}
			if n != frameSize {
				fmt.Println("Frame size mismatch: expected", frameSize, "got", n)
			}
		}
		stdout.Close()
		close(done)
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Fatalf("Failed during FFmpeg execution: %v", err)
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("FFmpeg command failed: %v", err)
	}
}

func getFrame(video *Video) (*Frame, bool) {
	video.bufferMutex.Lock()
	defer video.bufferMutex.Unlock()
	if len(video.frameBuffer) < 1 {
		return nil, false
	}
	frame := video.frameBuffer[0]
	return &frame, true
}
func stepForward(video *Video) {
	clearBuffer(video)
	video.currentFrame += SKIP_AMOUNT_S * int(video.fps)
	if video.currentFrame > video.totalFrames {
		video.currentFrame = video.totalFrames
	}
	bufferVideo(video, video.currentFrame, BUFFER_OFFSET)
}
func stepBackward(video *Video) {
	clearBuffer(video)
	video.currentFrame -= SKIP_AMOUNT_S * int(video.fps)
	if video.currentFrame < 1 {
		video.currentFrame = 1
	}
	bufferVideo(video, video.currentFrame, BUFFER_OFFSET)
}
func setFrame(video *Video, frame int) {
	clearBuffer(video)
	video.currentFrame = frame
	bufferVideo(video, video.currentFrame, BUFFER_OFFSET)
}
func shiftBuffer(video *Video) {
	video.bufferMutex.Lock()
	defer video.bufferMutex.Unlock()
	if len(video.frameBuffer) > 0 {
		video.frameBuffer = video.frameBuffer[1:]
	}
}
func clearBuffer(video *Video) {
	video.bufferMutex.Lock()
	defer video.bufferMutex.Unlock()
	video.frameBuffer = video.frameBuffer[:0]
}

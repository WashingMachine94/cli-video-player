package main

import (
	"fmt"
	"time"
)

const TEST_BUFFER_SIZE int = 15
const TEST_BUFFER_OFFSET int = 30

var TEST_VIDEO Video

func runTests(filepath string) {
	TEST_VIDEO = loadVideo(filepath, TEST_BUFFER_OFFSET*2)
	setTerminalDimensions()
	// testBufferSpeed(&TEST_VIDEO)
	testGaussianBlur(&TEST_VIDEO)
	// testDrawSpeed(&TEST_VIDEO, 2)
}

func testBufferSpeed(video *Video) {
	bufferOffsets := []int{0, video.totalFrames / 4, video.totalFrames / 2, video.totalFrames / 4 * 3, video.totalFrames - TEST_BUFFER_OFFSET}

	for _, startFrame := range bufferOffsets {
		testBuffer(video, startFrame, TEST_BUFFER_OFFSET)
	}
}
func testBuffer(video *Video, startFrame int, bufferOffset int) {
	startTime := time.Now()
	bufferVideo(video, startFrame, bufferOffset)
	deltaTime := time.Now().Sub(startTime)

	minBufferTime := time.Second * time.Duration(video.fps) / time.Duration(bufferOffset)
	minBufferTimeMillis := minBufferTime.Milliseconds()
	bufferTimeMillis := deltaTime.Milliseconds()

	if bufferTimeMillis > minBufferTimeMillis {
		fmt.Printf(RED_COLOR)
	}
	fmt.Printf("Frame: %d-%d\n", startFrame, startFrame+bufferOffset)
	fmt.Printf("Buffertime: %d/%d ms\n", bufferTimeMillis, minBufferTimeMillis)
	fmt.Printf(RESET_COLOR)
	clearBuffer(video)
}

func testDrawSpeed(video *Video, durationSec int) {
	clearBuffer(video)
	bufferVideo(video, 1000, int(video.fps)*durationSec)

	startTime := time.Now()

	frame, _ := getFrame(video)
	oldFrame := processFrame(frame, video.width, video.height, CHANNELS)
	printFrame(oldFrame)
	shiftBuffer(video)

	for i := 0; i < len(video.frameBuffer); i++ {
		frame, _ := getFrame(video)
		setTerminalDimensions()
		newFrame := processFrame(frame, video.width, video.height, CHANNELS)
		frameDiff := getFrameDiff(oldFrame, newFrame)
		printFrame(&frameDiff)
		oldFrame = newFrame

		shiftBuffer(video)
		video.currentFrame++
	}

	deltaTime := time.Now().Sub(startTime)
	if deltaTime.Milliseconds() > int64(durationSec*1000) {
		fmt.Printf(RED_COLOR)
	}

	fmt.Printf("Frame: %d-%d\n", 0, int(video.fps)*durationSec)
	fmt.Printf("Drawtime: %d/%dms\n", deltaTime.Milliseconds(), time.Second*time.Duration(durationSec))

	fmt.Printf(RESET_COLOR)
	clearBuffer(video)
}

func testGaussianBlur(video *Video) {
	bufferVideo(video, video.totalFrames/9, TEST_BUFFER_OFFSET)
	frame, _ := getFrame(video)

	blurredFrame := gaussianBlur(video.width, video.height, frame, 1, 2)
	blurredFrame1 := gaussianBlur(video.width, video.height, frame, 2, 4)
	frame = subtractFrame(&blurredFrame, &blurredFrame1)
	asciiString := processFrame(frame, video.width, video.height, 1)
	printFrame(asciiString)
}

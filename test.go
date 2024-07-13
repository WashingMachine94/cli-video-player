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
	testBufferSpeed(&TEST_VIDEO)
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
	fmt.Printf("Buffertime: %d ms, MinBufferTime: %d ms\n", bufferTimeMillis, minBufferTimeMillis)
	fmt.Printf(RESET_COLOR)
}

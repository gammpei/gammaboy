package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const FILE_FMT = "%d.png"

type recorder struct {
	wg          *sync.WaitGroup
	frameNumber int
	tmpDir      string // The directory for the pngs.
	dstFile     string // The mp4.
}

func newRecorder() *recorder {
	prefix := time.Now().Format("2006-01-02-15h04m05.000")
	prefix = strings.Replace(prefix, ".", "s", -1)

	tmpDir, err := ioutil.TempDir(".", prefix+"_")
	check(err)

	var wg sync.WaitGroup
	return &recorder{
		wg:          &wg,
		frameNumber: 0,
		tmpDir:      tmpDir,
		dstFile:     prefix + ".mp4",
	}
}

func (recorder *recorder) addFrame(frame [144][160]u32) {
	recorder.wg.Add(1)
	go func(i int) {
		defer recorder.wg.Done()

		img := image.NewRGBA(image.Rect(0, 0, 160, 144))
		for y := 0; y < 144; y++ {
			for x := 0; x < 160; x++ {
				pixel := frame[y][x]
				a := u8(pixel >> 24)
				r := u8(pixel >> 16)
				g := u8(pixel >> 8)
				b := u8(pixel)
				img.Set(x, y, color.RGBA{r, g, b, a})
			}
		}

		filename := filepath.Join(recorder.tmpDir, fmt.Sprintf(FILE_FMT, i))
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0200)
		check(err)
		defer file.Close()

		err = png.Encode(file, img)
		check(err)
	}(recorder.frameNumber)

	recorder.frameNumber++
}

func (recorder *recorder) close() {
	recorder.wg.Wait()

	// 4.19 MHz, 154 * 456 cycles per frame
	const framerate = 4190000. / (154. * 456.)
	assert(59.6 <= framerate && framerate <= 59.7)

	argv := []string{
		"ffmpeg",
		"-r", fmt.Sprint(framerate), // The input framerate.
		"-i", filepath.Join(recorder.tmpDir, FILE_FMT), // The input images.
		"-pix_fmt", "yuv420p", // Makes the video playable in web browsers.
		"-y",             // Overwrite the destination file if it exists.
		recorder.dstFile, // The destination file.
	}
	fmt.Println(strings.Join(argv, " "))
	cmd := exec.Command(argv[0], argv[1:]...)

	err := cmd.Run()
	check(err)

	err = os.RemoveAll(recorder.tmpDir)
	check(err)
}

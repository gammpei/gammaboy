/*
 * gammaboy is a Game Boy emulator.
 * Copyright (C) 2018  gammpei
 *
 * This file is part of gammaboy.
 *
 * gammaboy is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * gammaboy is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with gammaboy.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"unsafe"
)

type gui struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	palette  [4]u32
	recorder *recorder
}

func newGui() *gui {
	err := sdl.Init(sdl.INIT_VIDEO)
	check(err)

	window, err := sdl.CreateWindow(
		"gammaboy",              // title
		sdl.WINDOWPOS_UNDEFINED, // x
		sdl.WINDOWPOS_UNDEFINED, // y
		160, 144, // width, height
		sdl.WINDOW_RESIZABLE, // flags
	)
	check(err)

	renderer, err := sdl.CreateRenderer(
		window,
		-1,     // index of the rendering driver
		0x0000, // flags
	)
	check(err)

	assert(sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, flags.scalingAlg))
	err = renderer.SetLogicalSize(160, 144)
	check(err)

	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_ARGB8888,
		sdl.TEXTUREACCESS_STREAMING,
		160, 144, // w, h
	)
	check(err)

	var palette [4]u32
	argb := func(r, g, b u8) u32 {
		var color u32 = 0x00000000
		const a u8 = 255
		for i, x := range [4]u8{b, g, r, a} {
			color |= u32(x) << uint(i*8)
		}
		return color
	}
	if flags.green {
		// https://upload.wikimedia.org/wikipedia/commons/f/f7/Screen_color_test_Gameboy.png
		palette = [4]u32{
			argb(155, 188, 15),
			argb(139, 172, 15),
			argb(48, 98, 48),
			argb(15, 56, 15),
		}
	} else {
		palette = [4]u32{
			argb(255, 255, 255), // White
			argb(170, 170, 170), // Light grey
			argb(85, 85, 85),    // Dark grey
			argb(0, 0, 0),       // Black
		}
	}

	var recorder *recorder = nil
	if flags.record {
		recorder = newRecorder()
	}

	return &gui{
		window:   window,
		renderer: renderer,
		texture:  texture,
		palette:  palette,
		recorder: recorder,
	}
}

func (gui *gui) drawFrame(st *st) {
	var screen [144][160]u32
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			screen[y][x] = gui.palette[0] // White
		}
	}

	LCDC := st.readMem_u8(0xFF40) // LCD Control
	lcdDisplayEnable := getBit(LCDC, 7)
	if lcdDisplayEnable {
		bgDisplayEnable := getBit(LCDC, 0)
		if bgDisplayEnable {
			gui.drawBackground(st, &screen)
		}
	}

	if gui.recorder != nil {
		gui.recorder.addFrame(screen)
	}

	const pitch = 160 * 4
	err := gui.texture.Update(
		nil, // dst rect
		(*[144 * pitch]u8)(unsafe.Pointer(&screen))[:], // pixels
		pitch,
	)
	check(err)

	err = gui.renderer.Copy(
		gui.texture,
		nil, // srcrect
		nil, // dstrect
	)
	check(err)

	gui.renderer.Present()
}

func (gui *gui) drawBackground(st *st, screen *[144][160]u32) {
	LCDC := st.readMem_u8(0xFF40) // LCD Control

	var bgTileMap u16
	switch getBit(LCDC, 3) {
	case false:
		bgTileMap = 0x9800 // 0x9800-0x9BFF
	case true:
		bgTileMap = 0x9C00 // 0x9C00-0x9FFF
	}

	var tileSet u16
	switch getBit(LCDC, 4) {
	case false:
		tileSet = 0x8800 // 0x8800-0x97FF
	case true:
		tileSet = 0x8000 // 0x8000-0x8FFF
	}

	BGP := st.readMem_u8(0xFF47) // BackGround Palette
	var palette [4]u32
	for i, _ := range palette {
		colorIndex := (BGP >> uint(i*2)) & 0x03
		palette[i] = gui.palette[colorIndex]
	}

	var bg [256][256]u32
	for tileMapY := 0; tileMapY < 32; tileMapY++ {
		for tileMapX := 0; tileMapX < 32; tileMapX++ {
			tileMapIndex := tileMapY*32 + tileMapX

			tileSetIndex := st.readMem_u8(bgTileMap + u16(tileMapIndex))
			switch tileSet {
			case 0x8000:
			case 0x8800:
				tileSetIndex = u8(int(i8(tileSetIndex)) + 128)
			default:
				assert(false)
			}

			for tileY := 0; tileY < 8; tileY++ {
				lineAddr := tileSet + u16(int(tileSetIndex)*16+tileY*2)
				lowBits := st.readMem_u8(lineAddr)
				highBits := st.readMem_u8(lineAddr + 1)

				for tileX := 0; tileX < 8; tileX++ {
					bit := uint(7 - tileX)
					l := getBit(lowBits, bit)
					h := getBit(highBits, bit)

					var paletteIndex int
					switch {
					case !h && !l:
						paletteIndex = 0
					case !h && l:
						paletteIndex = 1
					case h && !l:
						paletteIndex = 2
					case h && l:
						paletteIndex = 3
					}

					x := tileMapX*8 + tileX
					y := tileMapY*8 + tileY
					bg[y][x] = palette[paletteIndex]
				}
			}
		}
	}

	// Update the screen.
	SCX := st.readMem_u8(0xFF43)
	SCY := st.readMem_u8(0xFF42)
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			screen[y][x] = bg[int(SCY)+y][int(SCX)+x]
		}
	}
}

func (gui *gui) processEvents() bool {
	for {
		switch event := sdl.PollEvent().(type) {
		case nil:
			return true
		case *sdl.QuitEvent:
			return false
		case *sdl.WindowEvent:
			if event.WindowID == sdl.WINDOWEVENT_CLOSE {
				return false
			}
		}
	}
}

func (gui *gui) close() {
	if gui.recorder != nil {
		gui.recorder.close()
	}

	gui.texture.Destroy()
	gui.renderer.Destroy()
	gui.window.Destroy()
	sdl.Quit()
}

func getScanline(st *st) u8 {
	// The LCD takes 456 cycles to draw one line.
	// It has 154 lines (144 visible lines + 10 "V-blank lines").
	scanline := (st.cycles / 456) % 154
	assert(0 <= scanline && scanline <= 153)
	return u8(scanline)
}

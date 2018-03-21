package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var hashToRom map[string][]u8 = loadTestRoms()

func TestBlarggsTestRoms(t *testing.T) {
	hashToName := map[string]string{
		"fe61349cbaee10cc384b50f356e541c90d1bc380185716706b5d8c465a03cf89": "01-special",
		"ca553e606d9b9c86fbd318f1b916c6f0b9df0cf1774825d4361a3fdff2e5a136": "03-op sp,hl",
		"7686aa7a39ef3d2520ec1037371b5f94dc283fbbfd0f5051d1f64d987bdd6671": "04-op r,imm",
		"d504adfa0a4c4793436a154f14492f044d38b3c6db9efc44138f3c9ad138b775": "05-op rp",
		"17ada54b0b9c1a33cd5429fce5b765e42392189ca36da96312222ffe309e7ed1": "06-ld r,r",
		"ab31d3daaaa3a98bdbd9395b64f48c1bdaa889aba5b19dd5aaff4ec2a7d228a3": "07-jr,jp,call,ret,rst",
		"974a71fe4c67f70f5cc6e98d4dc8c096057ff8a028b7bfa9f7a4330038cf8b7e": "08-misc instrs",
		"b28e1be5cd95f22bd1ecacdd33c6f03e607d68870e31a47b15a0229033d5ba2a": "09-op r,r",
		"7f5b8e488c6988b5aaba8c2a74529b7c180c55a58449d5ee89d606a07c53514a": "10-bit ops",
		"0ec0cf9fda3f00becaefa476df6fb526c434abd9d4a4beac237c2c2692dac5d3": "11-op a,(hl)",
	}

	var wg sync.WaitGroup
	for hash, name := range hashToName {
		wg.Add(1)
		go func(hash, name string) {
			defer wg.Done()
			rom, ok := hashToRom[hash]
			assert(ok)
			testBlarggTestRom(t, rom, name)
		}(hash, name)
	}
	wg.Wait()
}

func testBlarggTestRom(t *testing.T, rom []u8, name string) {
	gb := newTestGameBoy(rom)
	defer gb.close()

	go gb.run()

	filename := name + ".gb"
	expectedString := name + "\n\n\nPassed\n"
	for _, expectedByte := range []u8(expectedString) {
		var actualByte u8
		select {
		case actualByte = <-gb.st.linkCable:
		case <-time.After(30 * time.Second):
			t.Fatalf(`%q took too long to write to the link cable.`, filename)
		}

		if actualByte != expectedByte {
			t.Fatalf(`Expected 0x%02X=%q, got 0x%02X=%q in %q.`,
				expectedByte, expectedByte, actualByte, actualByte, filename,
			)
		}
	}
}

func loadTestRoms() map[string][]u8 {
	defer stopWatch("loadTestRoms", time.Now())
	hashToRom := map[string][]u8{}
	var mutex sync.Mutex

	var wg sync.WaitGroup
	err := filepath.Walk("testRoms", func(path string, fi os.FileInfo, err error) error {
		check(err)

		if fi.IsDir() || !strings.HasSuffix(path, ".gb") {
			return nil
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			rom, err := ioutil.ReadFile(path)
			check(err)

			mutex.Lock()
			defer mutex.Unlock()
			hashToRom[sha256Hash(rom)] = rom
		}()
		return nil
	})
	check(err)
	wg.Wait()

	return hashToRom
}

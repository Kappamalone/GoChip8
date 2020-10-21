package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

var white uint32 = 0x9a5f61
var black uint32 = 0xd4dcd7
var multiplier int32 = 20

func checkErr(err error, desc string) {
	if err != nil {
		panic(desc)
	}
}

func initWindow() (*sdl.Window, *sdl.Surface, *sdl.Renderer) {

	//Window size vars
	var width int32 = 64 * multiplier
	var height int32 = 32 * multiplier

	//Initialise SDL
	err := sdl.Init(sdl.INIT_EVERYTHING)
	checkErr(err, "SDL initialisation error")

	//Create window
	window, err := sdl.CreateWindow("GoChip-8", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		width, height, sdl.WINDOW_SHOWN)
	checkErr(err, "Window creation error")

	//Get surface
	surface, err := window.GetSurface()
	checkErr(err, "surface creation error")

	//Get renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	checkErr(err, "renderer creation error")

	rect := sdl.Rect{0, 0, width, height}
	renderer.SetDrawColor(uint8((white&0xFF0000)>>16), uint8((white&0x00FF00)>>8), uint8((white & 0x0000FF)), 1)
	renderer.Clear()
	renderer.DrawRect(&rect)
	renderer.Present()

	return window, surface, renderer
}

func drawFromArray(window *sdl.Window, surface *sdl.Surface, renderer *sdl.Renderer, videoArr [32][64]uint8) {
	//Called at 60fps or something of that sorts
	var color uint32
	renderer.Clear()

	//Loop through 64x32 space and draw rects
	//Color determined by value held by the videoArr
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if videoArr[y][x] == 1 {
				color = black
			} else {
				color = white
			}
			renderer.SetDrawColor(uint8((color&0xFF0000)>>16), uint8((color&0x00FF00)>>8), uint8((color & 0x0000FF)), 1)
			rect := sdl.Rect{int32(x) * multiplier, int32(y) * multiplier, multiplier, multiplier}
			renderer.DrawRect(&rect)
			renderer.FillRect(&rect)
		}
	}
	renderer.Present()
}

func main() {
	window, surface, renderer := initWindow()

	//Destroy window and quit SDL subsystems
	defer sdl.Quit()
	defer window.Destroy()

	cpu := initCPU("roms/IBM Logo.ch8")

	for i := 0; i < 400; i++ {
		instructionExecuted, drawBool := cpu.cycle()
		println(instructionExecuted)
		if drawBool {
			drawFromArray(window, surface, renderer, cpu.getDisplay())
		}

	}

	/*
		running := true
		for running {

			sdl.Delay(1000/60) //Run emulator at 60fps

			//Handle keyboard inputs
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch e := event.(type) {
				case *sdl.KeyboardEvent:
					if e.Type == sdl.KEYDOWN {
						switch e.Keysym.Scancode {
						case 30: //implement proper keypress detection from here onwards
							//fmt.Print("woo")
						}
						//30 31 32 33
						//20 26 8 21
						//4 22 7 9
						//29 27 6 25
					} else {
						//println("Hah keyup")
					}
				case *sdl.QuitEvent:
					println("Quit")
					running = false
					break
				}
			}
		}
	*/
}

package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"fmt"
)
//Color vars
var white uint32 = 0x2C2F33
var black uint32 = 0x7289DA
var perimColor uint32 = 0x7289DA

//Window size var
var multiplier int32 = 13
var perim int32 = 3

var debugWidth int32 = (17 * multiplier)
var debugHeight int32 = (12 * multiplier)

var width int32 = (64 * multiplier)
var height int32 = (32 * multiplier)

var screenWidth int32 = (64 * multiplier + (perim * 2)) + debugWidth + perim
var screenHeight int32 = (32 * multiplier + (perim * 2)) + debugHeight + perim

func checkErr(err error, desc string) {
	if err != nil {
		panic(desc)
	}
}

func setRenderColor(renderer *sdl.Renderer, color uint32){
	renderer.SetDrawColor(uint8((color&0xFF0000)>>16), uint8((color&0x00FF00)>>8), uint8((color & 0x0000FF)), 1)
}

func initWindow() (*sdl.Window, *sdl.Surface, *sdl.Renderer) {
	//Initialise SDL
	err := sdl.Init(sdl.INIT_EVERYTHING)
	checkErr(err, "SDL initialisation error")

	//Create window
	window, err := sdl.CreateWindow("GoChip-8", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		screenWidth, screenHeight, sdl.WINDOW_SHOWN)
	checkErr(err, "Window creation error")

	//Get surface
	surface, err := window.GetSurface()
	checkErr(err, "surface creation error")

	//Get renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	checkErr(err, "renderer creation error")

	//Draw borders onto screen
	setRenderColor(renderer,perimColor)
	border1 := sdl.Rect{0,0,screenWidth,screenHeight}
	renderer.FillRect(&border1)
	//border2 := sdl.Rect{0,0,width+(perim*2),height+(perim*2)}
	//renderer.FillRect(&border2)
	return window, surface, renderer
}

func drawFromArray(window *sdl.Window, surface *sdl.Surface, renderer *sdl.Renderer, videoArr [32][64]uint8) {
	//Called at 60fps or something of that sorts
	var color uint32

	//Loop through 64x32 space and draw rects
	//Color determined by value held by the videoArr
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if videoArr[y][x] == 1 {
				color = black
			} else {
				color = white
			}
			setRenderColor(renderer,color)
			pixel := sdl.Rect{int32(x) * multiplier + perim, int32(y) * multiplier + perim, multiplier, multiplier}
			renderer.FillRect(&pixel)
		}
	}
	renderer.Present()
}

func drawDebug(renderer *sdl.Renderer){
	//Function to draw all debug information to screen

	setRenderColor(renderer,white)
	//Right debug panel
	rightPanel := sdl.Rect{width+(perim*2),perim,debugWidth,height+debugHeight}
	renderer.FillRect(&rightPanel)

	//Bottom debug panel
	bottomPanel := sdl.Rect{perim,perim*2+height,width,debugHeight}
	renderer.FillRect(&bottomPanel)

}

func main() {
	window, surface, renderer := initWindow()

	//Destroy window and quit SDL subsystems
	defer sdl.Quit()
	defer window.Destroy()

	cpu := initCPU("roms/IBM Logo.ch8")
	drawDebug(renderer)

	for i := 0; i < 3000; i++ {
		memoryLocation,instructionExecuted, drawBool := cpu.cycle()
		fmt.Printf("%s - %s\n",memoryLocation,instructionExecuted)
		if drawBool {
			drawFromArray(window, surface, renderer, cpu.display)
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

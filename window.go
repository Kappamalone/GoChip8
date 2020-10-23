package main

import (
	"github.com/veandco/go-sdl2/sdl"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"fmt"
	"strings"
)

//Color vars
var white uint32 = 0x2C2F33
var black uint32 = 0x7289DA
var perimColor uint32 = 0x7289DA

//Window size var
var multiplier int32 = 13
var perim int32 = 3

var screenWidth int32 = (64 * multiplier + (perim * 2)) 
var screenHeight int32 = (32 * multiplier + (perim * 2))

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

	//Initialise termui
	err = ui.Init()
	checkErr(err,"Failed to intialise termui")

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

func initDebugging()(*widgets.Paragraph){
	//Initialise termui components

	instructionDebug := widgets.NewParagraph()
	instructionDebug.Title = "Instructions"
	instructionDebug.BorderStyle.Fg = ui.ColorBlue
	instructionDebug.SetRect(0,0,30,30)

	/*
	cpuStateDebug := widgets.NewParagraph()
	cpuStateDebug.Title = "Registers"
	cpuStateDebug.BorderStyle.Fg = ui.ColorMagenta
	cpuStateDebug.SetRect(40,0,30,15)
	*/

	return instructionDebug
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


func appendInstruction(slice *[]string,memoryAndInstruction string){
	//Appends instruction to instruction slice and removes first element
	
	*slice = append(*slice,memoryAndInstruction)
	*slice = (*slice)[1:]
}	

func getCPURegisters(c CPU){
	//Return formatted cpu register data: 4x5 of v0-vf and pc,sp,dt,st and index
}

func getCPUStack(c CPU){
	//Return formatted cpu stack data
}


func main() {
	window, surface, renderer := initWindow()

	//Destroy window, quit SDL subsystems and quit termui
	defer sdl.Quit()
	defer window.Destroy()
	defer ui.Close()

	//Initialise vm
	cpu := initCPU("roms/test_opcode.ch8")

	//Instruction slice
	var instructionSlice = make([]string,14)

	//Setup terminal debugging windows
	instructionDebug := initDebugging()
	execute := 0
	
	running := true
	for running {
		if execute < 100 {
			sdl.Delay(1000/600) //Run emulator at 60fps
			//Get data from execution of a cpu cycle, such as instruction executed at a given memory location
			memoryLocation,instructionExecuted, drawBool := cpu.cycle()
			memoryAndInstruction := fmt.Sprintf("[0x%s](fg:green)    ---    [%s](fg:yellow,)\n",memoryLocation,instructionExecuted)

			//Appends instruction to the instructionSlice to display in the debugging panel
			appendInstruction(&instructionSlice,memoryAndInstruction)

			//Draw to screen if cpu cycle updated screen
			if drawBool {
				drawFromArray(window, surface, renderer, cpu.display)
			}

			//Draw debug text from instructionQueue
			instructionDebug.Text = "\n"+strings.Join(instructionSlice[:],"\n")
			ui.Render(instructionDebug)
			execute++
		}
		

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
	
}

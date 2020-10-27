package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/veandco/go-sdl2/sdl"
	"strings"
	"time"
)

//Color vars
var white uint32 = 0x2C2F33
var black uint32 = 0x7289DA
var perimColor uint32 = 0x7289DA

//Window size var
var multiplier int32 = 15
var perim int32 = 6

var screenWidth int32 = (64*multiplier + (perim * 2))
var screenHeight int32 = (32*multiplier + (perim * 2))

var stepMode int = -1 //Used to check if instruction-by-instruction mode is toggled
var dcounter int = 0  //Used to check if debug should be updated every cycle or every instruction slice
var executing int = 1 //Used to pause cpu
var running bool = true

var speed int = 600
var start time.Time = time.Now()

//Initialise vm,window,surface and renderer
var window, surface, renderer = initWindow()
var cpu = initCPU("roms/PONG2")

//Instruction slice thats rendered on the debug window
var instructionSlice = make([]string, 14)

//Setup terminal debugging windows
var instructionDebug, cpuVDebug, cpuGDebug, cpuStack, debugMode = initDebugging()

func checkErr(err error, desc string) {
	if err != nil {
		panic(desc)
	}
}

func setRenderColor(renderer *sdl.Renderer, color uint32) {
	renderer.SetDrawColor(uint8((color&0xFF0000)>>16), uint8((color&0x00FF00)>>8), uint8((color & 0x0000FF)), 1)
}

func fullCycle(isStepping bool) { //If stepmode, then show debug every cycle
	//Get data from execution of a cpu cycle, such as instruction executed at a given memory location
	memoryLocation, instructionExecuted, drawBool := cpu.cycle()
	memoryAndInstruction := fmt.Sprintf("[0x%s](fg:green)   ---   [%s](fg:yellow,)\n", memoryLocation, instructionExecuted)

	//Appends instruction to the instructionSlice to display in the debugging panel
	appendInstruction(&instructionSlice, memoryAndInstruction)

	//Draw to screen if cpu cycle updated screen
	if drawBool {
		drawFromArray(window, surface, renderer, &cpu.display)
	}

	//Draw debug text from cpu
	instructionDebug.Text = "\n" + strings.Join(instructionSlice[:], "\n")
	cpuVDebug.Text, cpuGDebug.Text, debugMode.Text, cpuStack.Text = getDebugInformation(*cpu, executing, stepMode)

	quickUpdateDebug()

	if isStepping {
		ui.Render(instructionDebug, cpuVDebug, cpuGDebug, cpuStack, debugMode)
	} else {
		//show debug every instruction slice
		if dcounter%14 == 0 {
			ui.Render(instructionDebug, cpuVDebug, cpuGDebug, cpuStack, debugMode)
		}
	}
	dcounter++
}

func initWindow() (*sdl.Window, *sdl.Surface, *sdl.Renderer) {
	//Initialise SDL
	err := sdl.Init(sdl.INIT_EVERYTHING)
	checkErr(err, "SDL initialisation error")

	//Initialise termui
	err = ui.Init()
	checkErr(err, "Failed to intialise termui")

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

	//border2 := sdl.Rect{0,0,width+(perim*2),height+(perim*2)}
	//renderer.FillRect(&border2)
	return window, surface, renderer
}

func initDebugging() (*widgets.Paragraph, *widgets.Paragraph, *widgets.Paragraph, *widgets.Paragraph, *widgets.Paragraph) {
	//Initialise termui components

	instructionDebug := widgets.NewParagraph()
	instructionDebug.Title = "Instructions"
	instructionDebug.BorderStyle.Fg = ui.ColorBlue
	instructionDebug.SetRect(1, 0, 30, 30)

	cpuVRegisters := widgets.NewParagraph()
	cpuVRegisters.Title = "V Registers"
	cpuVRegisters.BorderStyle.Fg = ui.ColorRed
	cpuVRegisters.SetRect(31, 0, 60, 10)

	cpuOtherRegisters := widgets.NewParagraph()
	cpuOtherRegisters.Title = "General Registers"
	cpuOtherRegisters.BorderStyle.Fg = ui.ColorMagenta
	cpuOtherRegisters.SetRect(61, 0, 92, 30)

	cpuStack := widgets.NewParagraph()
	cpuStack.Title = "Stack"
	cpuStack.BorderStyle.Fg = ui.ColorRed
	cpuStack.SetRect(31, 10, 60, 30)

	debugMode := widgets.NewParagraph()
	debugMode.Title = "Debug modes"
	debugMode.BorderStyle.Fg = ui.ColorWhite
	debugMode.SetRect(93, 0, 119, 30)

	return instructionDebug, cpuVRegisters, cpuOtherRegisters, cpuStack, debugMode
}

func drawFromArray(window *sdl.Window, surface *sdl.Surface, renderer *sdl.Renderer, videoArr *[32][64]uint8) {
	renderer.Clear()

	//Called at 60fps or something of that sorts
	var color uint32

	//Draw borders onto screen
	setRenderColor(renderer, perimColor)
	border1 := sdl.Rect{0, 0, screenWidth, screenHeight}
	renderer.FillRect(&border1)

	//Loop through 64x32 space and draw rects
	//Color determined by value held by the videoArr

	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if (*videoArr)[y][x] == 1 {
				color = black
			} else {
				color = white
			}
			setRenderColor(renderer, color)
			pixel := sdl.Rect{int32(x)*multiplier + perim, int32(y)*multiplier + perim, multiplier, multiplier}
			renderer.FillRect(&pixel)
		}
	}
	renderer.Present()
}

func appendInstruction(slice *[]string, memoryAndInstruction string) {
	//Appends instruction to instruction slice and removes first element

	*slice = append(*slice, memoryAndInstruction)
	*slice = (*slice)[1:]
}

func getDebugInformation(c CPU, running int, stepping int) (string, string, string, string) {
	//Return formatted cpu register data: 4x5 of v0-vf and pc,sp,dt,st and index as well as stack and modes
	cpuVFormatted := make([]string, 0)

	for i := 0; i < 8; i++ {
		stringLine := fmt.Sprintf("[V%X](fg:green) = [#%02X](fg:yellow)       [V%X](fg:green) = [#%02X](fg:yellow)", i, c.V[i], i+8, c.V[i+8])
		cpuVFormatted = append(cpuVFormatted, stringLine)
	}

	//May god forgive me for this line of code
	cpuGeneralFormatted := strings.Split(fmt.Sprintf(
		"[PC](fg:green) = [#%04X](fg:yellow)   [SP](fg:green) = [#%02X](fg:yellow),[DT](fg:green) = [#%02X](fg:yellow)   [ST](fg:green) = [#%02X](fg:yellow),[I](fg:green)  = [#%04X](fg:yellow)",
		c.pc, c.stkptr, c.delayTimer, c.soundTimer, c.index), ",")

	modes := make([]string, 0)
	modes = append(modes, fmt.Sprintf(" [Running](fg:yellow): %t", running == 1))
	modes = append(modes, fmt.Sprintf(" [Stepmode](fg:yellow): %t", stepping == 1))
	modes = append(modes, fmt.Sprintf(" [Speed](fg:yellow): %d", speed))

	//Return formatted cpu stack data
	cpuStackFormatted := make([]string, 0)
	for i := 0; i < 16; i++ {
		stringLine := fmt.Sprintf("[S%X](fg:green) = [0x%04X](fg:yellow)", i, c.stack[i])
		cpuStackFormatted = append(cpuStackFormatted, stringLine)
	}

	cpuVFormattedf := strings.Join(cpuVFormatted, "\n")
	cpuGeneralFormattedf := strings.Join(cpuGeneralFormatted, "\n\n")
	modesf := strings.Join(modes, "\n\n")
	cpuStackF := "\n" + strings.Join(cpuStackFormatted, "\n")

	return cpuVFormattedf, cpuGeneralFormattedf, modesf, cpuStackF

}

func main() {
	//draw the initial screen
	drawFromArray(window, surface, renderer, &cpu.display)

	//Destroy window, quit SDL subsystems and quit termui
	defer sdl.Quit()
	defer window.Destroy()
	defer renderer.Destroy()
	defer ui.Close()

	for running {
		if stepMode == 1 {
			//Allow for step by step instruction execution
			pause := true
			for pause {
				for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
					switch e := event.(type) {
					case *sdl.KeyboardEvent:
						if e.Type == sdl.KEYDOWN {
							//Toggle stepmode off; kinda ugly but eh
							switch e.Keysym.Scancode {
							case 12:
								stepMode *= -1
								quickUpdateDebug()
								pause = false
							case 19:
								executing *= -1
								quickUpdateDebug()
							case 18: //press O to step
								fullCycle(true)
								pause = false
							case 47: // [ decreases speed of emulation
								speed -= 10
								limitSpeed(&speed)
								quickUpdateDebug()
							case 48: // ] increases speed of emulation
								speed += 10
								limitSpeed(&speed)
								quickUpdateDebug()
							}
						}
					case *sdl.QuitEvent:
						pause = false
						running = false
						break
					}
				}
			}
		} else {
			if time.Since(start) >= (time.Second)/100 {
				if executing == 1 && stepMode == -1 {
					//Decreasing timers at rate of 100hz rather than 60hz
					//Because decrements of 0.6 aren't possible
					//However this isn't too bad considering that doing the buffered cycles usually takes around 5ms
					if cpu.delayTimer > 0 {
						cpu.delayTimer--
					}
					if cpu.soundTimer > 0 {
						cpu.soundTimer--
					}
					//time1 := time.Now()
					for i := 0; i < speed/100; i++ {
						//execute a certain number of cycles per 1/100th of a second
						fullCycle(false)
					}
					//fmt.Println(time.Since(time1))
				}
				start = time.Now()
			}
			//Handle keyboard inputs
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch e := event.(type) {
				case *sdl.KeyboardEvent:
					if e.Type == sdl.KEYDOWN {
						//fmt.Println(e.Keysym.Scancode)
						switch e.Keysym.Scancode {
						case 12:
							//TODO: also this as a way to run the cpu as a command line thing
							//Toggle stepmode with I
							stepMode *= -1
							quickUpdateDebug()
						case 19:
							//Toggle pause with P
							executing *= -1
							quickUpdateDebug()
						case 47: // [ decreases speed of emulation
							speed -= 10
							limitSpeed(&speed)
							quickUpdateDebug()
						case 48: // ] increases speed of emulation
							speed += 10
							limitSpeed(&speed)
							quickUpdateDebug()
						default:
							cpu.handleKeypress(e.Keysym.Scancode, true)
						}
					} else if e.Type == sdl.KEYUP {
						cpu.handleKeypress(e.Keysym.Scancode, false)
					}
				case *sdl.QuitEvent:
					running = false
					break
				}
			}
		}
	}
}

func quickUpdateDebug() {
	_, _, debugMode.Text, _ = getDebugInformation(*cpu, executing, stepMode)
	ui.Render(debugMode)
}

func limitSpeed(speed *int) {
	//limit speed of emulation
	if *speed > 2000 {
		*speed = 2000
	} else if *speed < 50 {
		*speed = 50
	}
}

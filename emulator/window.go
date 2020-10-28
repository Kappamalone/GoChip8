package emulator

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"os"
	"strings"
	"time"
	"strconv"
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
var executing int = 1 //Used to pause cpu
var running bool = true

var s,_ = strconv.ParseInt(os.Args[2],10,64)
var speed int = int(s)

var timerCounter int = 0 //increment by 1 every 0.1s, is used to decrement timers at 60hz
var start time.Time = time.Now()

//Initialise vm,window,surface and renderer
var window, surface, renderer = initWindow()
var cpu = initCPU(fmt.Sprintf("%s",os.Args[1]))

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

func fullCycle() { //If stepmode, then show debug every cycle
	//Get data from execution of a cpu cycle, such as instruction executed at a given memory location
	memoryLocation, instructionExecuted, drawBool := cpu.cycle()
	memoryAndInstruction := fmt.Sprintf("[0x%s](fg:green)   ---   [%s](fg:yellow,)\n", memoryLocation, instructionExecuted)

	//Appends instruction to the instructionSlice to display in the debugging panel
	appendInstruction(&instructionSlice, memoryAndInstruction)

	//Draw to screen if cpu cycle updated screen
	if drawBool {
		drawFromArray(window, surface, renderer, &cpu.display)
	}

	//Set debug text from cpu
	instructionDebug.Text = "\n" + strings.Join(instructionSlice[:], "\n")
	cpuVDebug.Text, cpuGDebug.Text, debugMode.Text, cpuStack.Text = getDebugInformation(*cpu, executing, stepMode)
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

func initBeep() (beep.StreamSeekCloser, *beep.Ctrl) {
	file, err := os.Open("beep.mp3")
	checkErr(err, "couldn't find beep.mp3")

	streamer, format, err := mp3.Decode(file)
	checkErr(err, "couldn't decode beep.mp3")

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer)}
	return streamer, ctrl
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

func runWindow() {
	//Init beep and related stuff
	streamer, ctrl := initBeep()
	speaker.Lock()
	ctrl.Paused = true
	speaker.Unlock()
	speaker.Play(ctrl)

	//draw the initial screen
	drawFromArray(window, surface, renderer, &cpu.display)

	//Destroy window, quit SDL subsystems,termui and beep
	defer sdl.Quit()
	defer window.Destroy()
	defer renderer.Destroy()
	defer ui.Close()
	defer streamer.Close()

	for running {
		if stepMode == 1 {
			//Prevent sound when stepping
			speaker.Lock()
			ctrl.Paused = true
			speaker.Unlock()
			
			//Allow for step by step instruction execution
			pause := true
			for pause {
				ui.Render(instructionDebug, cpuVDebug, cpuGDebug, cpuStack, debugMode) //Draw debug menu
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
								fullCycle()
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
				start = time.Now()

				//Draw debug console at 100hz
				ui.Render(instructionDebug, cpuVDebug, cpuGDebug, cpuStack, debugMode)

				//Play sound if ST > 0
				if cpu.soundTimer > 0 {
					speaker.Lock()
					ctrl.Paused = false
					speaker.Unlock()
				} else {
					speaker.Lock()
					ctrl.Paused = true
					speaker.Unlock()
				}

				if executing == 1 && stepMode == -1 {
					//Decrease timers at 60hz
					timerCounter++
					if (timerCounter % 100) < 60 {
						if cpu.delayTimer > 0 {
							cpu.delayTimer--
						}
						if cpu.soundTimer > 0 {
							cpu.soundTimer--
						}
					} else if timerCounter == 100 {
						timerCounter = 0
					}

					for i := 0; i < speed/100; i++ {
						//execute a certain number of cycles per 1/100th of a second
						fullCycle()
					}
				} else if executing == -1 {
					//Prevent sound when paused
					speaker.Lock()
					ctrl.Paused = true
					speaker.Unlock()
				}
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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"

	"github.com/blackjack/webcam"
	"github.com/felixge/pidctrl"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
)

func bound(x float64, min float64, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func mapVal(x float64, inmin float64, inmax float64, outmin float64, outmax float64) float64 {
	return bound((x-inmin)*(outmax-outmin)/(inmax-inmin)+outmin, outmin, outmax)
}

func fourcc(str string) webcam.PixelFormat { //get camera four letter formt code
	lc := []rune(str)
	if len(lc) != 4 {
		panic(fmt.Errorf("four letter code is not four letters (got %d)", len(lc)))
	}
	return webcam.PixelFormat(
		uint32(lc[0]) |
			(uint32(lc[1]) << 8) |
			(uint32(lc[2]) << 16) |
			(uint32(lc[3]) << 24),
	)
}

type fss []webcam.FrameSize

func (f fss) Len() int {
	return len(f)
}
func (f fss) area(i int) uint64 {
	return uint64(f[i].MaxHeight) * uint64(f[i].MaxWidth)
}
func (f fss) Less(i, j int) bool {
	return f.area(i) > f.area(j)
}
func (f fss) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func main() {
	app := cli.NewApp()
	app.Name = "bottomside"
	app.Usage = "run bottomsside of robot"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "serial",
			Value: "/dev/ttyACM0",
			Usage: "serial port for arduino",
		},
		cli.StringSliceFlag{
			Name:  "camera",
			Usage: "cameras",
		},
		cli.StringFlag{
			Name:  "http",
			Value: ":8080",
			Usage: "http server listening address",
		},
	}
	app.Action = func(c *cli.Context) error {
		ard, err := ConnectArduino(c.GlobalString("serial"))
		if err != nil {
			return err
		}
		var botstate BotState
		var mots = struct {
			va, vb, vc, vd Motor //Vertical motors
			hl, hr         Motor //Horizontal motors
			clawrot        Motor //claw rotation motor
		}{
			ard.mot(26, 22, 25, 2),
			ard.mot(27, 23, 24, 3),
			ard.mot(32, 28, 31, 4),
			ard.mot(33, 29, 30, 5),
			ard.mot(38, 34, 37, 6),
			ard.mot(39, 35, 36, 7),
			ard.mot(65, 65, 65, 65),
		}
		var clawopenpin uint8 = 65 //TODO: change
		var clawvertpin uint8 = 65 //TODO: change
		var lightpin uint8 = 65    //TODO: change
		var obsoundpin uint8 = 65  //TODO: change
		ard.
			pinOut(clawopenpin).
			pinOut(clawvertpin).
			pinOut(lightpin).
			pinOut(obsoundpin)
		go func() { //run arduino processing in background
			pcx := pidctrl.NewPIDController(1, 64, 4).SetOutputLimits(-1, 1).Set(0)
			pcy := pidctrl.NewPIDController(1, 64, 4).SetOutputLimits(-1, 1).Set(0)
			pcv := pidctrl.NewPIDController(1, 64, 4).SetOutputLimits(-1, 1)
			mpur := ard.procMPU() //set up MPU
			acv := Vec3{mpur.AcX, mpur.AcY, mpur.AcZ}.Unit()
			θx := math.Asin(acv.Y()) //these axes are right - look at a diagram
			θy := math.Asin(acv.X())
			g := acv.Magnitude() //gravity
			pcv.Set(g)           //set vertical target to gravity
			ttx := 0.0           //x tilt target
			tty := 0.0           //y tilt target
			vt := 0.0            //vertical target
			for {
				//read accelerometer
				mpur = ard.procMPU()
				acv = Vec3{mpur.AcX, mpur.AcY, mpur.AcZ}.Unit()
				aθx, aθy := math.Asin(acv.Y()), math.Asin(acv.X())
				ΔθxΔt, ΔθyΔt := mpur.GyX, mpur.GyY
				Δt := mpur.DT.Seconds()
				//use Euler's method
				Δθx := ΔθxΔt * Δt
				Δθy := ΔθyΔt * Δt
				θx += Δθx
				θy += Δθy
				//apply filter to use accelerometer to deal with drift
				θx = θx*0.998 + aθx*0.02
				θy = θy*0.998 + aθy*0.02
				//generate axis unit vectors
				xaxis := Vec3{math.Cos(θy), 0, math.Sin(θy)}
				yaxis := Vec3{0, math.Cos(θx), math.Sin(θx)}
				zaxis := yaxis.CrossP(xaxis).Unit()
				//rotate accleration vector
				realacv := mat3{
					{I.CosAng(xaxis), I.CosAng(yaxis), I.CosAng(zaxis)},
					{J.CosAng(xaxis), J.CosAng(yaxis), J.CosAng(zaxis)},
					{K.CosAng(xaxis), K.CosAng(yaxis), K.CosAng(zaxis)},
				}.multiplyVec(acv)
				//fetch targets
				botstate.Lock()
				tiltx := botstate.TiltX
				tilty := botstate.TiltY
				fwd := botstate.Forward
				turn := botstate.Turn
				vert := botstate.Vertical
				clawopen := botstate.ClawOpen
				clawh := botstate.ClawHorizontal
				clawv := botstate.ClawVert
				ls := botstate.Light
				obs := botstate.OBSSound
				botstate.UpdateCount++
				botstate.Unlock()
				//run PID calculation
				if tiltx != ttx {
					pcx.Set(tiltx)
					ttx = tiltx
				}
				if tilty != tty {
					pcy.Set(tilty)
					tty = tilty
				}
				if vert != vt {
					pcv.Set(g + vert)
					vt = vert
				}
				τx := pcx.Update(θx)
				τy := pcx.Update(θy)
				vf := pcv.Update(realacv.Z())
				//calculate motor target forces
				va := τx + τy + vf
				vb := τx + (-τy) + vf
				vc := (-τx) + (-τy) + vf
				vd := (-τx) + τy + vf
				hl := fwd + turn
				hr := fwd - turn
				//convert forces to motor settings
				va = mapVal(va, -1, 1, -255, 255)
				vb = mapVal(vb, -1, 1, -255, 255)
				vc = mapVal(vc, -1, 1, -255, 255)
				vd = mapVal(vd, -1, 1, -255, 255)
				hl = mapVal(hl, -1, 1, -255, 255)
				hr = mapVal(hr, -1, 1, -255, 255)
				//write to motors
				mots.va.set(int16(va))
				mots.vb.set(int16(vb))
				mots.vc.set(int16(vc))
				mots.vd.set(int16(vd))
				mots.hl.set(int16(hl))
				mots.hr.set(int16(hr))
				//do claw stuff
				mots.clawrot.set(clawh)
				ard.digWrite(clawopenpin, clawopen)
				ard.anaWrite(clawvertpin, clawv)
				//other
				ard.digWrite(lightpin, ls)
				if obs {
					ard.freq(obsoundpin, 6000)
				} else {
					ard.freq(obsoundpin, 0)
				}
				ard.flush()
			}
		}()
		//set up cameras
		for i, campath := range c.GlobalStringSlice("camera") {
			cam, err := webcam.Open(campath) //open it
			if err != nil {
				panic(err)
			}
			pxfmts := cam.GetSupportedFormats() //get formats & search for MJPEG
			pxfmt := webcam.PixelFormat(0)
			for _, f := range []webcam.PixelFormat{fourcc("MJPG"), fourcc("JPEG")} {
				if pxfmts[f] != "" {
					pxfmt = f
				}
			}
			if pxfmt == 0 {
				panic(errors.New("No supported pixel format detected"))
			}
			fsizes := fss(cam.GetSupportedFrameSizes(pxfmt)) //select biggest frame size
			sort.Sort(fsizes)
			_, _, _, err = cam.SetImageFormat(pxfmt, fsizes[0].MaxWidth, fsizes[0].MaxHeight) //set format
			if err != nil {
				panic(err)
			}
			err = cam.StartStreaming() //start camera
			if err != nil {
				panic(err)
			}
			framein := make(chan []byte)  //camera worker sends frames here
			frameout := make(chan []byte) //read from this to get frames
			go func() {                   //distributor: takes frames and sends them to web handlers
				f := <-framein
				for {
					select {
					case f = <-framein:
					case frameout <- f:
					}
				}
			}()
			go func() { //read frames from camera
				for {
					//wait up to 5 seconds for a frame to be ready
					err := cam.WaitForFrame(5)
					if err != nil {
						panic(err)
					}
					//read the frame
					dat, err := cam.ReadFrame()
					if err != nil {
						panic(err)
					}
					//send it to the distributor
					framein <- dat
				}
			}()
			//setup http handlers
			http.HandleFunc(fmt.Sprintf("/cam/%d/frame.jpg", i), func(w http.ResponseWriter, r *http.Request) {
				//send as a JPEG file
				w.Header().Set("Content-Type", "image/jpeg")
				w.Header().Set("Cache-Control", "no-cache")
				w.Write(<-frameout) //send one frame
			})
			http.HandleFunc(fmt.Sprintf("/cam/%d/stream", i), func(w http.ResponseWriter, r *http.Request) {
				//send as an MJPEG stream
				w.Header().Set("Content-Type", "multipart/x-mixed-replace;boundary=--BOUNDARY")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				for { //send frames
					_, err := w.Write(<-frameout)
					if err != nil {
						return
					}
					_, err = w.Write([]byte("--BOUNDARY"))
					if err != nil {
						return
					}
				}
			})
		}
		//http handler for info about state
		http.HandleFunc("/info.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			w.Header().Set("Cache-Control", "no-cache")
			json.NewEncoder(w).Encode(struct {
				Serial   string
				Cameras  []string
				NCameras int
			}{
				Serial:   c.GlobalString("serial"),
				Cameras:  c.GlobalStringSlice("camera"),
				NCameras: len(c.GlobalStringSlice("camera")),
			})
		})
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		//handle controls
		var xm xMutex
		http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Cache-Control", "no-cache")
			err := xm.Lock()
			if err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			defer xm.Unlock()
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Printf("Error upgrading websocket: %q", err.Error())
				return
			}
			defer ws.Close()
			for {
				var bs BotState
				err = ws.ReadJSON(&bs)
				if err != nil {
					log.Printf("Error reading from control socket: %q", err.Error())
					return
				}
				func() {
					botstate.Lock()
					defer botstate.Unlock()
					botstate.Forward = bs.Forward
					botstate.Turn = bs.Turn
					botstate.Vertical = bs.Vertical
					botstate.TiltX = bs.TiltX
					botstate.TiltY = bs.TiltY
					botstate.Light = bs.Light
					botstate.ClawVert = bs.ClawVert
					botstate.ClawHorizontal = bs.ClawHorizontal
					botstate.OBSSound = bs.OBSSound
				}()
			}
		})
		//start web server
		go func() { panic(http.ListenAndServe(c.GlobalString("http"), nil)) }()
		//wait forever
		select {}
	}
}

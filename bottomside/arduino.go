package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/tarm/serial"
)

//Arduino is a linked arduino
type Arduino struct {
	serout *bufio.Reader
	serin  *bufio.Writer
	PrevT  uint32
	dcache map[uint8]bool
	acache map[uint8]uint8
}

//ConnectArduino connnects to an Arduino
func ConnectArduino(port string) (*Arduino, error) {
	//init bus
	bus, err := serial.OpenPort(&serial.Config{
		Name: port,
		Baud: 115200,
	})
	if err != nil {
		return nil, err
	}
	//init I/O
	a := &Arduino{serout: bufio.NewReader(bus), serin: bufio.NewWriter(bus)}
	//Wait for init message
	ln, err := a.serout.ReadString(byte('\n'))
	if err != nil {
		bus.Close()
		return nil, err
	}
	if ln != "init" {
		return nil, fmt.Errorf("Expected 'init' but got %q", ln)
	}
	//Wait for start message
	ln, err = a.serout.ReadString(byte('\n'))
	if err != nil {
		bus.Close()
		return nil, err
	}
	if ln != "start" {
		return nil, fmt.Errorf("Expected 'start' but got %q", ln)
	}
	return a, nil
}

//send arguments and recieve output
func (a *Arduino) exCmd(in []byte, nout uint8) ([]byte, error) {
	_, err := a.serin.Write(in)
	if err != nil {
		return nil, err
	}
	if nout == 0 { //if there is no return value
		return nil, nil
	}
	out := make([]byte, nout)
	o := out
reread:
	n, err := a.serout.Read(out)
	if err != nil {
		return nil, err
	}
	if n < len(o) {
		o = o[n:]   //shift over
		goto reread //keep reading
	}
	return out, nil
}

//MPUReading is a struct returned by an MPU read
type MPUReading struct {
	AcX, AcY, AcZ float64       //in g (gravities)
	Temp          float64       //celcius
	GyX, GyY, GyZ float64       //in radians/second (must integrate afterward)
	DT            time.Duration //time since last measurement
}

//call procMPU on the arduino
func (a *Arduino) procMPU() MPUReading {
	dat, err := a.exCmd([]byte{1 /*procMPU*/}, 18)
	if err != nil {
		panic(err)
	}
	var mpudat struct {
		AcX, AcY, AcZ int16
		Temp          int16
		GyX, GyY, GyZ int16
		T             uint32
	}
	//decode big edian
	err = binary.Read(bytes.NewReader(dat), binary.BigEndian, &mpudat)
	if err != nil {
		panic(err)
	}
	defer func() {
		a.PrevT = mpudat.T //update time
	}()
	//do conversions
	return MPUReading{
		AcX:  float64(mpudat.AcX) / math.Pow(2, 14),
		AcY:  float64(mpudat.AcY) / math.Pow(2, 14),
		AcZ:  float64(mpudat.AcZ) / math.Pow(2, 14),
		Temp: float64(mpudat.Temp)/340.00 + 36.53,
		GyX:  (float64(mpudat.GyX) / math.Pow(2, 14)) * 100 * (math.Pi / 180),
		GyY:  (float64(mpudat.GyY) / math.Pow(2, 14)) * 100 * (math.Pi / 180),
		GyZ:  (float64(mpudat.GyZ) / math.Pow(2, 14)) * 100 * (math.Pi / 180),
		DT:   time.Millisecond * time.Duration(mpudat.T-a.PrevT),
	}
}

func b2b(v bool) uint8 {
	switch v {
	case false:
		return 0
	case true:
		return 1
	}
	return 65 //because the go compiler is weird
}

//call digWrite on the Arduino
func (a *Arduino) digWrite(pin uint8, val bool) *Arduino {
	if a.dcache[pin] == val {
		return a
	}
	_, err := a.exCmd([]byte{2, pin, b2b(val)}, 0)
	if err != nil {
		panic(err)
	}
	a.dcache[pin] = val
	return a
}

//call anaWrite on the Arduino
func (a *Arduino) anaWrite(pin uint8, val uint8) *Arduino {
	if a.acache[pin] == val {
		return a
	}
	_, err := a.exCmd([]byte{3, pin, val}, 0)
	if err != nil {
		panic(err)
	}
	a.acache[pin] = val
	return a
}

//call pinOut on the Arduino
func (a *Arduino) pinOut(pin uint8) *Arduino {
	_, err := a.exCmd([]byte{4, pin}, 0)
	if err != nil {
		panic(err)
	}
	return a
}

func (a *Arduino) mot(enable uint8, clockwise uint8, counterclockwise uint8, pwm uint8) Motor {
	a.pinOut(enable)
	a.pinOut(clockwise)
	a.pinOut(counterclockwise)
	a.pinOut(pwm)
	m := Motor{
		Ard:    a,
		Enable: enable,
		CW:     clockwise,
		CCW:    counterclockwise,
		PWM:    pwm,
	}
	m.set(0)
	return m
}

func (a *Arduino) freq(pin uint8, freq uint16) *Arduino {
	err := binary.Write(a.serin, binary.BigEndian, struct {
		cmd  uint8
		pin  uint8
		freq uint16
	}{
		cmd:  5,
		pin:  pin,
		freq: freq,
	})
	if err != nil {
		panic(err)
	}
	return a
}

func (a *Arduino) flush() *Arduino {
	err := a.serin.Flush()
	if err != nil {
		panic(err)
	}
	return a
}

type Motor struct {
	Ard     *Arduino
	Enable  uint8
	CW, CCW uint8
	PWM     uint8
}

func (m Motor) set(spd int16) {
	if spd > 255 {
		spd = 255
	} else if spd < -255 {
		spd = -255
	}
	mag := spd
	if spd < 0 {
		mag *= -1
	}
	m.Ard.digWrite(m.Enable, spd != 0)
	m.Ard.digWrite(m.CW, spd > 0)
	m.Ard.digWrite(m.CCW, spd < 0)
	m.Ard.anaWrite(m.PWM, uint8(mag))
}

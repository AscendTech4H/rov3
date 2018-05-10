package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/tarm/serial"
)

//Arduino is a linked arduino
type Arduino struct {
	serout     *bufio.Reader
	serin      *bufio.Writer
	PrevT      uint32
	motcache   map[uint8]int
	servocache map[uint8]uint8
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

func (a *Arduino) flush() *Arduino {
	err := a.serin.Flush()
	if err != nil {
		panic(err)
	}
	return a
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
	//send cmd
	_, err := fmt.Fprintln(a.serin, "3")
	if err != nil {
		panic(err)
	}
	//flush
	a.flush()
	//decode big edian
	var mpudat struct {
		AcX, AcY, AcZ int16
		Temp          int16
		GyX, GyY, GyZ int16
		T             uint32
	}
	err = binary.Read(a.serout, binary.BigEndian, &mpudat)
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

// setMotor sets a motor speed. speed must be in [-1, 1]
func (a *Arduino) setMotor(motnum uint8, speed float64) *Arduino {
	val := int(speed * 255)
	if val > 255 {
		val = 255
	} else if val < -255 {
		val = -255
	}
	oldval, cached := a.motcache[motnum]
	if !cached || val != oldval {
		_, err := fmt.Fprintf(a.serin, "1 %d %d\n", motnum, val)
		if err != nil {
			panic(err)
		}
	}
	a.motcache[motnum] = val
	return a
}

// setServo sets a servo position; pos must be in degrees in [0, 180]
func (a *Arduino) setServo(servonum uint8, pos uint8) *Arduino {
	if pos > 180 {
		pos = 180
	}
	oldval, cached := a.servocache[servonum]
	if !cached || pos != oldval {
		_, err := fmt.Fprintf(a.serin, "2 %d %d\n", servonum, pos)
		if err != nil {
			panic(err)
		}
	}
	a.servocache[servonum] = pos
	return a
}

type Motor struct {
	num uint8
	a   *Arduino
}

func (a *Arduino) mot(motnum uint8) Motor {
	return Motor{
		num: motnum,
		a:   a,
	}
}

func (m *Motor) set(speed float64) {
	m.a.setMotor(m.num, speed)
}

type Servo struct {
	num uint8
	a   *Arduino
}

func (a *Arduino) servo(servonum uint8) Servo {
	return Servo{
		num: servonum,
		a:   a,
	}
}

func (m *Servo) set(pos uint8) {
	m.a.setServo(m.num, pos)
}

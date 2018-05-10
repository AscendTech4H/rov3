package main

import "sync"

//BotState is the state of the robot
type BotState struct {
	sync.Mutex
	Forward        float64 //between -1 and 1
	Turn           float64 //between -1 and 1
	Vertical       float64 //in m/s^2
	TiltX, TiltY   float64 //in radians
	ClawOpen       uint8   //is the claw supposed to be open
	ClawVert       uint8   //claw vertical tilt
	ClawHorizontal float64 //claw horizontal tilt
	UpdateCount    uint64  //number of times the motors have beeen updated
	Light          bool
	OBSSound       bool
}

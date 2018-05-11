#include <Servo.h>
#include <Wire.h>
#include <Tlc5940.h>
#include <tlc_animations.h>
#include <tlc_config.h>
#include <tlc_fades.h>
#include <tlc_progmem_utils.h>
#include <tlc_servos.h>
#include <tlc_shifts.h>

#define modeOut(p) pinMode(p, OUTPUT)

const int servopins[] = {/*todo*/};
const int nservos = sizeof(servopins)/sizeof(int);
Servo servos[nservos];

#define MPU_addr 0x68

void setup() {
	//start serial at 115200
	Serial.begin(115200);
	//attatch servos
	for(int i = 0; i < nservos; i++) {
		servos[i].attach(servopins[i]);
	}
    //TLC5940 setup
    Tlc.init(0);
    Tlc.update();
	//mpu setup
	Wire.begin();
	Wire.beginTransmission(MPU_addr);
	Wire.write(0x6B);   //set PWR_MGMT_1 register to 0 (which turns it on)
	Wire.write(0);
	Wire.write(0x1C);   //set ACCEL_CONFIG sensor to +-2g
	Wire.write(0);
	Wire.endTransmission(true);
	//print "started" to activate controller code
	Serial.println("started");
}

long serRead() {
	while(Serial.available()<1);
	return Serial.parseInt();
}

void cmdMotor() {
	Tlc.set(serRead(), serRead());
}

void cmdServo() {
	servos[serRead()].write(serRead());
}

void cmdUpdate() {
    Tlc.update();
}

//used to convert a value to big edian
#define BE(v, n) ((v >> (sizeof(v)-n)) & 0b11111111)

void cmdMPU() {
	Wire.beginTransmission(MPU_addr);
	Wire.write(0x3B);											//start at register 0x3B
	Wire.endTransmission(false);								//stop sending
	Wire.requestFrom(MPU_addr, 14, true);						//request 14 registers
	unsigned long t = millis();									//get time stamp
	for(int i = 0; i < 14; i++) Serial.write(Wire.read());		//copy to serial
	for(int i = 0; i < sizeof(t); i++) Serial.write(BE(t, i));	//send time stamp
}

void waitStep() {
	while(Serial.available() < 1);
	while(Serial.available() > 0) Serial.read();
}

void loop() {
	switch(serRead()) {
	case 0:
		Serial.println("lol");
		break;
	case 1:
		cmdMotor();
		break;
	case 2:
		cmdServo();
		break;
	case 3:
		cmdMPU();
		break;
    case 4:
        cmdUpdate();
        break;
	default:
		Serial.println("bad command");
	}
  /*for(int i = 0; i < nmotors; i++) {
	Serial.print(i);
	Serial.println("cw");
	setMotor(motors[i], 255);
	waitStep();
	Serial.print(i);
	Serial.println("stop");
	setMotor(motors[i], 0);
	waitStep();
	Serial.print(i);
	Serial.println("ccw");
	setMotor(motors[i], -255);
	waitStep();
	Serial.print(i);
	Serial.println("stop");
	setMotor(motors[i], 0);
	waitStep();
  }*/
}

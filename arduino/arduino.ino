#include <Servo.h>
#include <Wire.h>

struct motor {
  int cwpin;
  int ccwpin;
  int enablepin;
  int pwmpin;
  bool isweird;
};

void setMotor(struct motor mot, int val) {
	digitalWrite(mot.enablepin, val != 0);
	if(mot.isweird) {
		analogWrite(mot.cwpin, (val > 0) ? val : 0);
		analogWrite(mot.cwpin, (val < 0) ? -val : 0);
	} else {
		digitalWrite(mot.cwpin, val > 0);
		digitalWrite(mot.ccwpin, val < 0);
		analogWrite(mot.pwmpin, abs(val));
  	}
}

#define modeOut(p) pinMode(p, OUTPUT)

void setupMotor(struct motor mot) {
  	modeOut(mot.cwpin);
  	modeOut(mot.ccwpin);
  	modeOut(mot.enablepin);
  	if(mot.isweird) modeOut(mot.pwmpin);
}

const struct motor motors[] = {
	{24, A8, 38, 45},
	{49, 26, 34, 5},
	{25, 29, 40, 4},
	{27, 43, 41, 3},
	{12, 44, 47, -1, true},
	{44, 11, 50, -1, true},
	{6, 7, 14, -1, true}
};
const int nmotors = sizeof(motors)/sizeof(struct motor);

const int servopins[] = {/*todo*/};
const int nservos = sizeof(servopins)/sizeof(int);
Servo servos[nservos];

#define MPU_addr 0x68

void setup() {
	//start serial at 115200
	Serial.begin(115200);
	//initialize motors
	for(int i = 0; i < nmotors; i++) {
		setupMotor(motors[i]);
	}
	//attatch servos
	for(int i = 0; i < nservos; i++) {
		servos[i].attach(servopins[i]);
	}
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

int serRead() {
	while(Serial.available()<1);
	return Serial.parseInt();
}

void cmdMotor() {
	setMotor(motors[serRead()], serRead());
}

void cmdServo() {
	servos[serRead()].write(serRead());
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

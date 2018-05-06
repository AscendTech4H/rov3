#include<Wire.h>

#define MPU_addr 0x68

//used to convert a value to big edian
#define BE(v, n) ((v >> (sizeof(v)-n)) & 0b11111111)

//load data from mpu
void procMPU() {
    Wire.beginTransmission(MPU_addr);
    Wire.write(0x3B);                                   //start at register 0x3B
    Wire.endTransmission(false);                        //stop sending
    Wire.requestFrom(MPU_addr, 14, true);               //request 14 registers
    unsigned long t = millis();                         //get time stamp
    for(int i = 0; i < 14; i++) Serial.write(Wire.read());  //copy to serial
    for(int i = 0; i < sizeof(t); i++) Serial.write(BE(t, i)); //send time stamp
}

//digitalWrite command
void digWrite() {
    while(Serial.available()<2);
    digitalWrite(Serial.read(), Serial.read());
}

//analogWrite command
void anaWrite() {
    while(Serial.available()<2);
    analogWrite(Serial.read(), Serial.read());
}

//set pin to out
void pinOut() {
    while(Serial.available()<1);
    pinMode(Serial.read(), OUTPUT);
}

//generate noise on a pin
void noise() {
    while(Serial.available()<3);
    int pin = Serial.read();
    unsigned int freq = (((unsigned int)Serial.read())<<8)|((unsigned int)Serial.read());
    if(freq == 0) {
        noTone(pin);
    } else {
        tone(pin, freq);
    }
}

void loop() {
    while(Serial.available()<1);
    switch(Serial.read()) {
    case 1:
        procMPU();
        break;
    case 2:
        digWrite();
        break;
    case 3:
        anaWrite();
        break;
    case 4:
        pinOut();
        break;
    case 5:
        noise();
        break;
    }
}

void setup() {
    //start serial
    Serial.begin(115200);
    Serial.println("init");
    //mpu setup
    Wire.begin();
    Wire.beginTransmission(MPU_addr);
    Wire.write(0x6B);   //set PWR_MGMT_1 register to 0 (which turns it on)
    Wire.write(0);
    Wire.write(0x1C);   //set ACCEL_CONFIG sensor to +-2g
    Wire.write(0);
    Wire.endTransmission(true);
    //notify master of completion
    Serial.println("start");
}

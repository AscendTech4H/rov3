all: topside.tar.gz bottomside.tar.gz
topside.o: $(glob topside/*.go) godeps
	GOOS=linux GOARCH=arm GOARM=7 go build -v -o topside.o ./topside
bottomside.o: $(glob bottomside/*.go) godeps
	GOOS=linux GOARCH=arm GOARM=7 go buildm -v -o bottomside.o ./bottomside
godeps: $(glob topside/*.go) $(glob bottomside/*.go)
	go get -v ./...
.PHONY: static
static:
	$(MAKE) -C static
topside.tar.gz: topside.o static
	tar -cf topside.tar.gz topside.o static
bottomside.tar.gz: bottomside.o arduino/arduino.ino
	tar -cf bottomside.tar.gz bottomside.o arduino/arduino.ino

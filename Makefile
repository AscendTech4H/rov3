all: topside.o bottomside.o static
topside.o: $(glob topside/*.go) godeps
    GOOS=linux GOARCH=arm GOARM=7 go build -o topside.o ./topside
bottomside.o: $(glob bottomside/*.go) godeps
    go build -o bottomside.o ./bottomside
godeps: $(glob topside/*.go) $(glob bottomside/*.go)
    go get -v ./...
.PHONY: static
static:
    $(MAKE) -C static

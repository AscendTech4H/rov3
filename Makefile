all: topside.tar.gz bottomside.tar.gz

# create custom GOPATH if the user requests it
ifeq ($(USEGOPATH),y)
GBUILDPATH = gopath/src/github.com/jadr2ddude/rov3/
GPTH = github.com/jadr2ddude/rov3
GOPATH = $(shell pwd)/gopath
GB = gb
$(GBUILDPATH):
	mkdir -p $@
$(GBUILDPATH)/bottomside $(GBUILDPATH)/topside: $(GBUILDPATH)
	mkdir $@
$(GBUILDPATH)/bottomside/%.go: bottomside/%.go $(GBUILDPATH)/bottomside
	cp $< $@
$(GBUILDPATH)/topside/topside.go: topside/topside.go $(GBUILDPATH)/topside
	cp $< $@
.PHONY: gb
gb: $(foreach v,$(shell echo bottomside/*.go),$(GBUILDPATH)/$(v)) $(GBUILDPATH)/topside/topside.go
clean-gpath:
	rm -rf gopath
else
GB =
GBUILDPATH = .
GPTH = .
clean-gpath:
endif

# topside/bottomside binaries
topside.o: $(foreach v,$(shell echo topside/*.go),$(GBUILDPATH)/$(v)) godeps
	GOPATH=$(GOPATH) GOOS=linux GOARCH=arm GOARM=7 go build -v -o topside.o $(GPTH)/topside
bottomside.o: $(foreach v,$(shell echo bottomside/*.go),$(GBUILDPATH)/$(v)) godeps
	GOPATH=$(GOPATH) GOOS=linux GOARCH=arm GOARM=7 go build -v -o bottomside.o $(GPTH)/bottomside
# install go packages used
godeps: $(shell echo topside/*.go) $(shell echo bottomside/*.go) $(GB)
	go get -v $(GPTH)/...

# static web files
.PHONY: static
static:
	$(MAKE) -C static

# output tars
topside.tar.gz: topside.o static
	tar -cf topside.tar.gz topside.o static
bottomside.tar.gz: bottomside.o arduino/arduino.ino
	tar -cf bottomside.tar.gz bottomside.o arduino/arduino.ino

# clean targets
clean-static:
	$(MAKE) -C static clean
clean: clean-gpath clean-static
	rm *.o *.tar.gz

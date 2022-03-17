BALLAST := $(shell echo $$RANDOM)
BUILD_OPTS := -ldflags "-X 'github.com/mpkondrashin/demomw/cmd/sw/main.Ballast=$(BALLAST)'"

sw21.iso: sw21.exe autorun.ini
	mkdir -p sw
	cp sw21.exe sw
	cp autorun.ini sw
	rm -f sw21.iso
	hdiutil makehybrid -iso -joliet -o sw21.iso sw

sw21.exe: main.go
	GOOS=windows GOARCH=386 go build $(BUILD_OPTS)

clean:
	rm -rf sw21.iso sw21.exe sw21


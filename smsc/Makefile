all: clean linux64 linux32 freebsd64

clean:
	rm -rf dist/freebsd64
	rm -rf dist/linux64
	rm -rf dist/linux32

freebsd64:
	mkdir -p dist/freebsd64
	GOOS=freebsd GOARCH=amd64 go build -o dist/freebsd64/sms
	strip dist/freebsd64/sms

linux64:
	mkdir -p dist/linux64
	GOOS=linux GOARCH=amd64 go build -o dist/linux64/sms
	strip dist/linux64/sms

linux32:
	mkdir -p dist/linux32
	GOOS=linux GOARCH=386 go build -o dist/linux32/sms
	strip dist/linux32/sms

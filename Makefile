all: clean linux64 linux32 freebsd64

clean:
	rm -rf dist/*
	rm -rf smsq/dist/*
	rm -rf smss/dist/*
	rm -rf smsc/dist/*

freebsd64:
	mkdir -p dist/freebsd64
	mkdir -p smss/dist/freebsd64
	mkdir -p smsq/dist/freebsd64
	mkdir -p smsc/dist/freebsd64
	GOOS=freebsd GOARCH=amd64 go build -o dist/freebsd64/sms
	GOOS=freebsd GOARCH=amd64 go build -o smsq/dist/freebsd64/smsq smsq/smsq.go
	GOOS=freebsd GOARCH=amd64 go build -o smss/dist/freebsd64/smss smss/smss.go
	GOOS=freebsd GOARCH=amd64 go build -o smsc/dist/freebsd64/smsc smsc/smsc.go
	strip dist/freebsd64/sms
	strip smss/dist/freebsd64/smss
	strip smsq/dist/freebsd64/smsq
	strip smsc/dist/freebsd64/smsc

linux64:
	mkdir -p dist/linux64
	mkdir -p smss/dist/linux64
	mkdir -p smsq/dist/linux64
	mkdir -p smsc/dist/linux64
	GOOS=linux GOARCH=amd64 go build -o dist/linux64/sms
	GOOS=linux GOARCH=amd64 go build -o smsq/dist/linux64/smsq smsq/smsq.go
	GOOS=linux GOARCH=amd64 go build -o smss/dist/linux64/smss smss/smss.go
	GOOS=linux GOARCH=amd64 go build -o smsc/dist/linux64/smsc smsc/smsc.go
	strip dist/linux64/sms
	strip smss/dist/linux64/smss
	strip smsq/dist/linux64/smsq
	strip smsc/dist/linux64/smsc

linux32:
	mkdir -p dist/linux32
	mkdir -p smss/dist/linux32
	mkdir -p smsq/dist/linux32
	mkdir -p smsc/dist/linux32
	GOOS=linux GOARCH=386 go build -o dist/linux32/sms
	GOOS=linux GOARCH=amd64 go build -o smsq/dist/linux32/smsq smsq/smsq.go
	GOOS=linux GOARCH=amd64 go build -o smss/dist/linux32/smss smss/smss.go
	GOOS=linux GOARCH=amd64 go build -o smsc/dist/linux32/smsc smsc/smsc.go
	strip dist/linux32/sms
	strip smss/dist/linux32/smss
	strip smsq/dist/linux32/smsq
	strip smsc/dist/linux32/smsc

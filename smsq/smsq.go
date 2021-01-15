package main

import (
	"os"
	"log"
	"flag"
	"time"
	"strings"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type QMessage struct {
	PlainText    bool `json:"plaintext"`
	LowPriority  bool `json:"low_priority"`
	HighPriority bool `json:"high_priority"`
	MessageData  string `json:"data"`
}

type Config struct {
	NatsServer string `json:"nats_server"`
	// NatsQueue string `json:"nats_queue"`
	// NatsProject string `json:"nats_project"`
}

var config Config
var debug bool

func main() {
	var plaintext       bool
	var high_priority   bool
	var low_priority    bool

	// PARSE ARGUMENTS, PERFORM BASIC SANITY CHECKS ********** //
	flag.BoolVar(&debug,  "d", false, "Show some debug info")
	flag.BoolVar(&debug,  "debug",  false, "Show some debug info")

	flag.BoolVar(&plaintext,    "t", false, "Do not use any markup in the message")
	flag.BoolVar(&low_priority, "L", false, "Send silent/low priority message")
	flag.BoolVar(&high_priority,"H", false, "Send unmuted message nevertheless the gently policy")

	flag.BoolVar(&plaintext,    "plaintext",  false, "Do not use any markup in the message")
	flag.BoolVar(&low_priority, "low", false, "Send silent/low priority message")
	flag.BoolVar(&high_priority,"high",   false, "Send unmuted message nevertheless the gently policy")

	flag.Parse()

	args := flag.Args()

	if len(args) < 2 {
		log.Println("usage: smsq [--debug] <subject> <message>")
		os.Exit(15)
	}

	// FIND RIGHT CONFIG FILE TO READ ************************ //
	var filename string
	if _, err := os.Stat("sms.json"); err == nil {
		filename = "sms.json"

        } else if _, err := os.Stat("/spool/sms.json"); err == nil {
		filename = "/spool/sms.json"

        } else if _, err := os.Stat("/etc/sms.json"); err == nil {
		filename = "/etc/sms.json"

        }

	if debug { log.Println("CONFIG: " + filename) }

	// ******************************************************* //

	// READ CONFIG FROM KNOWN FILE *************************** //
	file, err := os.Open(filename)
	if err != nil {
		log.Print("CONFIG OPEN ERROR: ")
		log.Print(err)
		os.Exit(25)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Print("CONFIG PARSE ERROR: ")
		log.Print(err)
		os.Exit(27)
	}
	// ******************************************************* //

	msg_struct := QMessage {
		PlainText:    plaintext,
		LowPriority:  low_priority,
		HighPriority: high_priority,
		MessageData:  strings.Join(args[1:], " ")}

	if debug { log.Println("Encoding the message...") }
	msg_str, err := json.Marshal(msg_struct)
	if err != nil {
		log.Fatal(err)
		os.Exit(55)
	}

	if debug { log.Printf( "MODE: Queue\n" ) }

	opts := []nats.Option{nats.Name("NATS SMSQ Publisher")}
	opts = setupConnOptions(opts)

	// Connect to NATS
	if debug { log.Printf( "NATS [%s] Connecting...\n", config.NatsServer ) }
	nc, err := nats.Connect(config.NatsServer, opts...)
	if err != nil {
		log.Fatal(err)
	}

	defer nc.Close()

	msg_queue := args[0]

	if debug { log.Printf( "POSTing to [%s]...", msg_queue ) }
	nc.Publish(msg_queue, []byte(msg_str))
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
		os.Exit(75)
	}

	if debug { log.Printf( "POSTed\n" ) }

	// That's all folks!
	os.Exit(0)
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectHandler(func(nc *nats.Conn) {
		log.Printf("NATS Disconnected: will attempt reconnects for %.0fm", totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("NATS Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("NATS Exiting: %v", nc.LastError())
	}))
	return opts
}

package main

import (
	"os"
	"log"
	"flag"
	"time"
	"runtime"
	"encoding/json"

	"github.com/nats-io/nats.go"

	"github.com/go-telegram-bot-api/telegram-bot-api"

)

type QMessage struct {
	PlainText    bool `json:"plaintext"`
	LowPriority  bool `json:"low_priority"`
	HighPriority bool `json:"high_priority"`
	MessageData  string `json:"data"`
}

type Config struct {
	TgApitoken string `json:"tg_apitoken"`
	Chats map[string]int64 `json:"chats"`
	WorkingHours []int `json:"working_hours"`
	NatsServer string `json:"nats_server"`
	NatsQueue string `json:"nats_queue"`
	NatsProject string `json:"nats_project"`
}

var config Config
var debug bool

func main() {

	// PARSE ARGUMENTS, PERFORM BASIC SANITY CHECKS ********** //
	flag.BoolVar(&debug,  "d", false, "Show some debug info")
	flag.BoolVar(&debug,  "debug",  false, "Show some debug info")

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		log.Println("usage: smss [--debug] <subject|*>")
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

	if len(config.WorkingHours) != 2 {
		// Working hours are't configured. Using default value: 8-22
		config.WorkingHours[0] = 8
		config.WorkingHours[1] = 22

		if debug { log.Println("No WorkingHours found. Failed to default ones: %d..%d\n", config.WorkingHours[0], config.WorkingHours[1]) }
	}

	var chat_id int64
	if debug { log.Printf( "CONNECTing Tg Server...\n" ) }

	bot, err := tgbotapi.NewBotAPI(config.TgApitoken)
	if err != nil {
		log.Fatal(err)
		os.Exit(15)
	}
	if debug { log.Printf( "CONNECTed!\n" ) }


	if debug { log.Printf( "MODE: Server\n" ) }

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS SMSQ Publisher")}
	opts = setupConnOptions(opts)

	// Connect to NATS
	if debug { log.Printf( "NATS [%s] Connecting...\n", config.NatsServer ) }
	nc, err := nats.Connect(config.NatsServer, opts...)
	if err != nil {
		log.Fatal(err)
	}

	if debug { log.Printf( "NATS CONNECTed!\n" ) }

	msg_queue := getQueue(args[0], config)
	if debug { log.Printf( "QUEUE: %s\n", msg_queue ) }

	nc.Subscribe(msg_queue, func(msg *nats.Msg) {

		if val, ok := config.Chats[ msg.Subject ]; ok {
			chat_id = val

		} else {
			log.Printf("Uknown chat [%s]\n", msg.Subject)
			return
		}

		if debug { log.Printf( "MSG for %s / %d\n", msg.Subject, chat_id ) }

		// Try to decode a message
		var msg_struct QMessage
		err := json.Unmarshal(msg.Data, &msg_struct)
		if err != nil {
			log.Println(err)
			return;
		}

		tg_msg := tgbotapi.NewMessage(chat_id, msg_struct.MessageData)
		tg_msg.DisableNotification = silent_mode(msg_struct.LowPriority, msg_struct.HighPriority)

		if !msg_struct.PlainText {
			tg_msg.ParseMode = "Markdown"
		}

		if debug { log.Print( "POSTing... " ) }

		if _, err := bot.Send(tg_msg); err != nil {
			log.Println(err)
		}

		if debug { log.Print( "DONE!\n" ) }
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()

	if debug { log.Printf( "DISCONNECTED!\n" ) }
}

func silent_mode(Silent bool, Loud bool) bool {
	if Loud {
		if debug { log.Println("LOUD MODE OPTION") }
		return false

	} else if Silent {
		if debug { log.Println("SILENT MODE OPTION") }
		return true

	} else {
		dt := time.Now()
		hr, _, _ := dt.Clock()
		if hr < config.WorkingHours[0] || hr >= config.WorkingHours[1] {
			if debug { log.Println("SILENT MODE POLICY") }
			return true
		}
	}

	return false
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

func getQueue(target string, config Config) string {
	return "sms." + config.NatsProject + "." + config.NatsQueue + "." + target
}

package main

import (
	"os"
	"log"
	"flag"
	"time"
	"strings"
	"encoding/json"

	"github.com/go-telegram-bot-api/telegram-bot-api"

)

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
	var loud bool
	var plain bool
	var silent bool

	flag.BoolVar(&loud,   "l", false, "Force to switch on sound even out of working hours")
	flag.BoolVar(&plain,  "p", false, "Force to do not use Markup in message")
	flag.BoolVar(&debug,  "d", false, "Show some debug info")
	flag.BoolVar(&silent, "s", false, "Send a silent message")

	flag.BoolVar(&loud,   "loud",   false, "Force to switch on sound even out of working hours")
	flag.BoolVar(&plain,  "plain",  false, "Force to do not use Markup in message")
	flag.BoolVar(&debug,  "debug",  false, "Show some debug info")
	flag.BoolVar(&silent, "silent", false, "Send a silent message")


	flag.Parse()

	args := flag.Args()

	if len(args) < 2 {
		log.Println("usage: ./sms [--debug|--loud|--silent] <chat_name> <message>")
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

	msg_str := strings.Join(args[1:], " ")

	if debug { log.Printf( "CONNECTing Tg Server...\n" ) }

	bot, err := tgbotapi.NewBotAPI(config.TgApitoken)
	if err != nil {
		log.Fatal(err)
		os.Exit(15)
	}
	if debug { log.Printf( "CONNECTed!\n" ) }

	if debug { log.Printf( "MODE: Direct\n" ) }

	var chat_id int64
	if val, ok := config.Chats[args[0]]; ok {
		chat_id = val
	} else {
		log.Print("Uknown chat")
		os.Exit(25)
	}

	msg := tgbotapi.NewMessage(chat_id, msg_str)
	msg.DisableNotification = silent_mode(loud, silent, config)

	if !plain {
		msg.ParseMode = "Markdown"
	}

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}

}

func silent_mode(loud bool, silent bool, config Config) bool {
	if loud {
		if debug { log.Println("Loud mode used explicitly") }
		return true

	} else if !silent {
		if len(config.WorkingHours) != 2 {
			// Working hours are't configured. Using default value: 8-22
			config.WorkingHours[0] = 8
			config.WorkingHours[1] = 22

			if debug { log.Println("No WorkingHours found. Failed to default ones: %d..%d\n", config.WorkingHours[0], config.WorkingHours[1]) }
		}

		dt := time.Now()
		hr, _, _ := dt.Clock()
		if hr < config.WorkingHours[0] || hr >= config.WorkingHours[1] {
			if debug { log.Println("SILENT MODE POLICY") }
			return true
		}
	}

	return false
}

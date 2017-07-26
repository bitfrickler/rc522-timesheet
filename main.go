package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	rfid "github.com/firmom/go-rfid-rc522/rfid"
	rc522 "github.com/firmom/go-rfid-rc522/rfid/rc522"
	rpio "github.com/stianeikeland/go-rpio"
	// localdb "github.com/bitfrickler/rc522-timesheet/localdb"
)

var (
	deviceID, _ = os.Hostname()
	apiURL      = "http://10.0.26.106/LazyTimesheet/Timesheet/Post"
	apiKey      = "myApIKey"
	ledPin      = rpio.Pin(16)
	buzzerPin   = rpio.Pin(18)
	nobuzzer    *bool
)

func log(msg string) {
	fmt.Printf("%s: %s\n", time.Now().String(), msg)

	//TODO: Write log file
}

func main() {

	nobuzzer = flag.Bool("nobuzzer", false, "Disable buzzer")
	flag.Parse()

	if *nobuzzer {
		log("BUZZER DISABLED")
	}

	reset()

	startTicker()

	var oldvalue string

	reader, err := rc522.NewRfidReader()
	if err != nil {
		fmt.Println(err)
		return
	}
	readerChan, err := rfid.NewReaderChan(reader)
	if err != nil {
		fmt.Println(err)
		return
	}
	rfidChan := readerChan.GetChan()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log("received an interrupt")
			cleanup()
			exit()
		}
	}()

	openDB()

	log("waiting for card...")

	for {
		select {
		case id := <-rfidChan:
			if id != oldvalue {
				notifyRegisterCard(id)

				te := TimeEntry{TimeStamp: time.Now(), CardID: id, DeviceID: deviceID}
				err := saveLocal(te)

				if err != nil {
					notifyError()
					log(err.Error())
				} else {
					log("committed to local database: " + id)
				}

				oldvalue = id

				ticker := time.NewTicker(time.Second * 10)
				go func() {
					for _ = range ticker.C {
						if oldvalue != "" {
							fmt.Println("removing old id: ", oldvalue)
							oldvalue = ""
						}

						ticker.Stop()
					}
				}()

				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func exit() {
	log("exiting")
	os.Exit(0)
}

func cleanup() {
	closeDB()
}

func reset() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
	}

	defer rpio.Close()

	//ledPin.Low()
	buzzerPin.Low()
}

func notifyRegisterCard(cardID string) {

	log("registered: " + cardID)

	if !*nobuzzer {
		if err := rpio.Open(); err != nil {
			fmt.Println(err)
			//os.Exit(1)
		}

		// Unmap gpio memory when done
		defer rpio.Close()

		// Set pin to output mode
		//ledPin.Output()
		buzzerPin.Output()

		//led_pin.High()
		buzzerPin.High()
		//ledPin.High()
		time.Sleep(time.Second / 20)
		buzzerPin.Low()
		//ledPin.Low()
		time.Sleep(time.Second / 20)
		buzzerPin.High()
		//ledPin.High()
		time.Sleep(time.Second / 20)
		buzzerPin.Low()
		//ledPin.Low()
	}
}

func notifyError() {
	// TODO
}

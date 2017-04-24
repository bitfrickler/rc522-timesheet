package main

import (
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
	DeviceID, _ = os.Hostname()
	APIURL      = "http://10.0.26.106:8000/api/Time"
	APIKey      = "apikey123#"
	led_pin     = rpio.Pin(16)
	buzzer_pin  = rpio.Pin(18)
)

func Log(msg string) {
	fmt.Println(msg)

	//TODO: Write log file
}

func main() {

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("terminating:", s)

	reset()

	StartTicker()

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

	fmt.Println("waiting for card...")

	for {
		select {
		case id := <-rfidChan:
			if id != oldvalue {
				Log("card id: " + id)

				te := TimeEntry{TimeStamp: time.Now(), CardID: id, DeviceID: DeviceID}
				err := saveLocal(te)

				if err != nil {
					notifyError()
					Log(err.Error())
				} else {
					notifySuccess()
					Log("committed to local database")
				}

				oldvalue = id

				ticker := time.NewTicker(time.Second * 10)
				go func() {
					for _ = range ticker.C {
						if oldvalue != "" {
							fmt.Println("removing oldvalue", oldvalue)
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

func reset() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
	}

	defer rpio.Close()

	led_pin.Low()
	buzzer_pin.Low()
}

func notifySuccess() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	led_pin.Output()
	buzzer_pin.Output()

	//led_pin.High()
	buzzer_pin.High()
	led_pin.High()
	time.Sleep(time.Second / 20)
	buzzer_pin.Low()
	led_pin.Low()
	time.Sleep(time.Second / 20)
	buzzer_pin.High()
	led_pin.High()
	time.Sleep(time.Second / 20)
	buzzer_pin.Low()
	led_pin.Low()
}

func notifyError() {
	// TODO
}

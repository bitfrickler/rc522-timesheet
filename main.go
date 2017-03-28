package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	rfid "github.com/firmom/go-rfid-rc522/rfid"
	rc522 "github.com/firmom/go-rfid-rc522/rfid/rc522"
	rpio "github.com/stianeikeland/go-rpio"

	localdb "github.com/bitfrickler/rc522-timesheet/localdb"
)

// JSONTimeEntry struct for time entry to be logged
type JSONTimeEntry struct {
	APIKey   string
	CardID   string
	DeviceID string
}

var (
	DeviceID, _ = os.Hostname()
	APIURL      = "http://10.0.26.106:8000/api/Time"
	APIKey      = "apikey123#"
	led_pin     = rpio.Pin(16)
	buzzer_pin  = rpio.Pin(18)
)

func log(msg string) {
	fmt.Println(msg)

	//TODO: Write log file
}

func saveRemote(apiKey string, cardID string, deviceID string) (error) {
	j := JSONTimeEntry{APIKey: apiKey, CardID: cardID, DeviceID: deviceID}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(j)
	_, err := http.Post(APIURL, "application/json;charset=utf-8", b)

	return err
}

func main() {

	reset()

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
				log("card id: " + id)

				te := localdb.TimeEntry{TimeStamp: time.Now(), CardID: id, DeviceID: DeviceID}
				err := localdb.SaveLocal(te)

				if err != nil {
					log(err.Error())

					notifyError()
				} else {
					notify_success()
				}

				oldvalue = id

				ticker := time.NewTicker(time.Second * 10)
				go func() {
					for _ = range ticker.C {
						if oldvalue != "" {
							fmt.Println("Removing oldvalue", oldvalue)
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

func notify_success() {
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

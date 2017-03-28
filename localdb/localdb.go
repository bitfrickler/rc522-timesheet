package localdb

import (
	_ "bytes"
	"database/sql"
	_ "encoding/json"
	"fmt"
	_ "net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TimeEntry bla
type TimeEntry struct {
	TimeStamp        time.Time
	CardID, DeviceID string
}

func main() {

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

	te := TimeEntry{TimeStamp: time.Now(), CardID: "mycard123", DeviceID: "myhostname123"}
	SaveLocal(te)
}

// StartTicker starts the ticker for recurring data export to the remote JSON API
func StartTicker() {
	ticker := time.NewTicker(time.Second * 60)
	go func() {
		for _ = range ticker.C {

		}
	}()
}

func SaveLocal(te TimeEntry) error {
	db, err := sql.Open("sqlite3", "./timesheet.db")
	checkErr(err)

	stmt, err := db.Prepare("insert into timeentries(timestamp, cardid, deviceid) values(?,?,?)")
	checkErr(err)

	_, err = stmt.Exec(te.TimeStamp, te.CardID, te.DeviceID)

	db.Close()

	return err
}

func transferLocal() {
	db, err := sql.Open("sqlite3", "./timesheet.db")
	checkErr(err)

	rows, err := db.Query("select * from timeentries where transferdate is null order by date(timestamp) asc")
	checkErr(err)

	for rows.Next() {

	}

	rows.Close()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

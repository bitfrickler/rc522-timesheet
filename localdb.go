package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

// JSONTimeEntry struct for time entry to be logged
type JSONTimeEntry struct {
	APIKey    string
	CardID    string
	DeviceID  string
	TimeStamp string
}

// TimeEntry struct for local time entry to be logged
type TimeEntry struct {
	TimeStamp        time.Time
	CardID, DeviceID string
}

// func main() {

// 	// Set up channel on which to send signal notifications.
// 	// We must use a buffered channel or risk missing the signal
// 	// if we're not ready to receive when the signal is sent.
// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt)

// 	// Block until a signal is received.
// 	s := <-c
// 	fmt.Println("Got signal:", s)

// 	te := TimeEntry{TimeStamp: time.Now(), CardID: "mycard123", DeviceID: "myhostname123"}
// 	SaveLocal(te)
// }

// startTicker starts the ticker for recurring data export to the remote JSON API
func startTicker() {
	ticker := time.NewTicker(time.Second * 60)
	go func() {
		for _ = range ticker.C {
			log("exporting to remote API")
			transferLocal()
		}
	}()
}

func openDB() {
	log("opening database...")
	var err error
	db, err = sql.Open("sqlite3", "./timesheet.db")
	//checkErr(err)
	if err == nil {
		log("database open")
	} else {
		log("error opening database")
	}
}

func closeDB() {
	log("closing database")
	db.Close()
	log("database closed")
}

func saveLocal(te TimeEntry) error {
	//db, err := sql.Open("sqlite3", "./timesheet.db")
	//checkErr(err)

	stmt, err := db.Prepare("insert into timeentries(timestamp, cardid, deviceid) values(?,?,?)")
	checkErr(err)

	_, err = stmt.Exec(te.TimeStamp, te.CardID, te.DeviceID)

	//db.Close()

	return err
}

// saveRemote sends a timesheet entry to the remote JSON API
func saveRemote(apiKey string, cardID string, deviceID string, timeStamp string) error {
	j := JSONTimeEntry{APIKey: apiKey, CardID: cardID, DeviceID: deviceID, TimeStamp: timeStamp}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(j)
	_, err := http.Post(apiURL, "application/json;charset=utf-8", b)

	return err
}

func transferLocal() {
	var id int
	var timeStamp string
	var cardID string
	var stmt *sql.Stmt

	//db, err := sql.Open("sqlite3", "./timesheet.db")
	//checkErr(err)

	rows, err := db.Query("select id, timestamp, cardid from timeentries where transferdate is null order by date(timestamp) asc")
	checkErr(err)

	for rows.Next() {

		err = rows.Scan(&id, &timeStamp, &cardID)
		checkErr(err)

		log("sending to remote server: " + timeStamp)
		err = saveRemote(apiKey, cardID, deviceID, timeStamp)
		checkErr(err)

		if err == nil {
			log("sent to remote server: " + timeStamp)

			stmt, err = db.Prepare("update timeentries set transferdate = ? where id = ?")
			checkErr(err)

			_, err = stmt.Exec(time.Now(), id)
			checkErr(err)
		}

	}

	rows.Close()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

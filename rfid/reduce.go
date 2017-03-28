package rfid

import (
	"fmt"
)

type Reducer struct {
	reader   RfidReader
	oldvalue string
}

func NewReducer(reader RfidReader) (RfidReader, error) {
	return RfidReader(&Reducer{
		reader:   reader,
		oldvalue: "",
	}), nil
}

func (r *Reducer) ReadId() (string, error) {
	id, err := r.reader.ReadId()
	if err != nil {
		r.oldvalue = ""
		return id, err
	}
	if id != r.oldvalue {
		r.oldvalue = id
		#return id, fmt.Errorf("RFID: duplicate id echo")
		return id, err
	}
	#r.oldvalue = id
	#return id, err
}

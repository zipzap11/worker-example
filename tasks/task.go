package tasks

import (
	"errors"
	"time"

	"github.com/RichardKnop/machinery/v1/log"
)

func Add(args ...interface{}) (int64, error) {
	var res int64
	for _, v := range args {
		log.INFO.Printf("Adding %d to %d", v, res)
		res += v.(int64)
	}
	return res, nil
}

func PanicTask() (string, error) {
	panic(errors.New("oops"))
}

// LongRunningTask ...
func LongRunningTask() error {
	log.INFO.Print("Long running task started")
	for i := 0; i < 10; i++ {
		log.INFO.Print(10 - i)
		time.Sleep(1 * time.Second)
	}
	log.INFO.Print("Long running task finished")
	return nil
}

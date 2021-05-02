package main

import (
	"fmt"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"

	// replace with e.g. "gitlab.com/gomidi/rtmididrv" for real midi connections
	// driver "gitlab.com/gomidi/midi/testdrv"

	driver "gitlab.com/gomidi/rtmididrv"
)

func main() {
	drv, err := driver.New()
	must(err)

	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	in := ins[1]

	must(in.Open())

	defer in.Close()

	rd := reader.New(
		reader.NoLogger(),
		// write every message to the out port
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			fmt.Printf("got %s\n", msg)
		}),
	)

	err = rd.ListenTo(in)
	must(err)
	select {}

}

func forever() {
	for {
		fmt.Printf("%v+\n", time.Now())
		time.Sleep(time.Second)
	}
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

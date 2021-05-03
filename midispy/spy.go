package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"

	// replace with e.g. "gitlab.com/gomidi/rtmididrv" for real midi connections
	// driver "gitlab.com/gomidi/midi/testdrv"

	driver "gitlab.com/gomidi/rtmididrv"
)

func main() {
	var midiInNumber = flag.Int("midiIn", 0, "The MIDI IN device name to use")
	flag.Parse()

	drv, err := driver.New()
	must(err)

	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	printInPorts(ins)

	if len(ins)-1 < *midiInNumber {
		log.Fatalf("MIDI Port [%d] does not exist", *midiInNumber)
	}

	in := ins[*midiInNumber]

	log.Printf("Using MIDI IN [%s]", in.String())

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

func printPort(port midi.Port) {
	fmt.Printf("[%v] %s\n", port.Number(), port.String())
}

func printInPorts(ports []midi.In) {
	fmt.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
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

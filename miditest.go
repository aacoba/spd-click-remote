package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"

	// replace with e.g. "gitlab.com/gomidi/rtmididrv" for real midi connections
	// driver "gitlab.com/gomidi/midi/testdrv"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

type MidiChannels struct {
	writerChannel *writer.Writer
}

// This example reads from the first input and and writes to the first output port
func main() {
	// you would take a real driver here e.g. rtmididrv.New()
	// drv := driver.New("fake cables: messages written to output port 0 are received on input port 0")
	drv, err := driver.New()
	must(err)

	// make sure to close all open ports at the end
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	in, out := ins[1], outs[1]

	must(in.Open())
	must(out.Open())

	defer in.Close()
	defer out.Close()

	// the writer we are writing to
	wr := writer.New(out)

	// to disable logging, pass mid.NoLogger() as option
	rd := reader.New(
		reader.NoLogger(),
		// write every message to the out port
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			fmt.Printf("got %s\n", msg)
		}),
	)

	printInPorts(ins)
	printOutPorts(outs)

	// listen for MIDI
	err = rd.ListenTo(in)
	must(err)
	time.Sleep(100 * time.Millisecond)

	writer.NoteOffVelocity(wr, 50, 127)

	// Webserver stuff

	channels := &MidiChannels{writerChannel: wr}

	http.HandleFunc("/note/", channels.noteHandler)
	http.HandleFunc("/pc/", channels.programChangeRequestHandler)
	http.HandleFunc("/cc/", channels.controlChangeRequestHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	// err = writer.ProgramChange(wr, 2)
	// must(err)
}

func (mc *MidiChannels) noteHandler(w http.ResponseWriter, r *http.Request) {
	key, err := parameterMustBeUint8(w, *r.URL, "key")
	if err != nil {
		return
	}
	go func() {
		writer.NoteOn(mc.writerChannel, key, 127)
		time.Sleep(2 * time.Second)
		writer.NoteOff(mc.writerChannel, key)
	}()
	log.Printf("Played Note [%d]", key)
	fmt.Fprintln(w, "OK")
}

func (mc *MidiChannels) programChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	program, err := parameterMustBeUint8(w, *r.URL, "program")
	if err != nil {
		return
	}
	go func() {
		writer.ProgramChange(mc.writerChannel, program)
	}()
	log.Printf("Sent PC [%d]", program)
	fmt.Fprintln(w, "OK")
}

func (mc *MidiChannels) controlChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	controller, err := parameterMustBeUint8(w, *r.URL, "controller")
	if err != nil {
		return
	}
	value, err := parameterMustBeUint8(w, *r.URL, "value")
	if err != nil {
		return
	}
	go func() {
		writer.ControlChange(mc.writerChannel, controller, value)
	}()
	log.Printf("Sent CC [%d] [%d]", controller, value)
}

func getParameterAsUint8(u url.URL, name string) (uint8, error) {
	valStr := u.Query().Get(name)
	i64, err := strconv.ParseUint(valStr, 0, 8)
	if err != nil {
		return 0, err
	}
	i8 := uint8(i64)
	return i8, nil
}

func parameterMustBeUint8(w http.ResponseWriter, u url.URL, name string) (uint8, error) {
	parameter, err := getParameterAsUint8(u, name)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Invallid value for parameter [%s]", name)
		return 0, fmt.Errorf("invallid value for parameter [%s]", name)
	}
	return parameter, nil

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

func printOutPorts(ports []midi.Out) {
	fmt.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

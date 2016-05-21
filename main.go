package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// DP832 .
type DP832 struct {
	conn net.Conn
	Addr string // tcp address to device
}

// Measurement .
type Measurement struct {
	Channel Channel
	Voltage float64
	Current float64
	Power   float64
}

func (m Measurement) String() string {
	return fmt.Sprintf("%s: %fv %fA %f...", m.Channel, m.Voltage, m.Current, m.Power)
}

// Instrument .
type Instrument struct {
	Manufacturer string
	Model        string
	Serial       string
	Version      string
}

type Channel int

const (
	ChCur Channel = iota
	Ch1
	Ch2
	Ch3
)

var (
	chStrMap = map[Channel]string{
		ChCur: "",
		Ch1:   "CH1",
		Ch2:   "CH2",
		Ch3:   "CH3",
	}
)

func (c Channel) String() string {
	switch c {
	case ChCur:
		return "ChC"
	case Ch1:
		return "Ch1"
	case Ch2:
		return "Ch2"
	case Ch3:
		return "Ch3"
	default:
		panic("Invalid value")
	}
}

type chRange struct {
	VoltageMin float64
	VoltageMax float64
	CurrentMin float64
	CurrentMax float64
}

var chRanges = map[Channel]chRange{
	Ch1: {
		VoltageMin: 00.000,
		VoltageMax: 32.000,
		CurrentMin: 00.000,
		CurrentMax: 3.200,
	},
	Ch2: {
		VoltageMin: 00.000,
		VoltageMax: 32.000,
		CurrentMin: 00.000,
		CurrentMax: 3.200,
	},
	Ch3: {
		VoltageMin: 00.000,
		VoltageMax: 5.300,
		CurrentMin: 00.000,
		CurrentMax: 3.200,
	},
}

func (d *DP832) command(command string) (string, error) {
	_, err := fmt.Fprintf(d.conn, fmt.Sprintf("%s\n", command)) //
	if err != nil {
		return "", err
	}
	result, err := bufio.NewReader(d.conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	result = result[:len(result)-2]
	return result, nil
}

func (d *DP832) Connect() error {
	conn, err := net.Dial("tcp", d.Addr)
	if err != nil {
		return err
	}
	d.conn = conn
	i, err := d.IDN()
	if err != nil {
		return err
	}
	if i.Model != "DP832" {
		return fmt.Errorf("Epected model DP832, found %s", i.Model)
	}
	return nil
}

func (d *DP832) IDN() (Instrument, error) {
	var i Instrument
	s, err := d.command("*IDN?")
	if err != nil {
		return i, err
	}
	a := strings.Split(s, ",")
	i.Manufacturer = a[0]
	i.Model = a[1]
	i.Serial = a[2]
	i.Version = a[3]
	return i, nil
}

func (d *DP832) Measure(ch Channel) (Measurement, error) {
	var m Measurement
	cmd := fmt.Sprintf("MEAS:ALL? %s", chStrMap[ch])

	s, err := d.command(cmd)
	if err != nil {
		return m, err
	}

	a := strings.Split(s, ",")

	{
		f, err := strconv.ParseFloat(a[0], 64)
		if err != nil {
			return m, err
		}
		m.Current = f
	}

	{
		f, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			return m, err
		}
		m.Voltage = f
	}

	{
		f, err := strconv.ParseFloat(a[2], 64)
		if err != nil {
			return m, err
		}
		m.Power = f
	}
	m.Channel = ch

	return m, nil
}

func main() {
	var addrFlag = flag.String("addr", "192.168.0.200:5555", "tcp hostport")
	flag.Parse()
	dp := DP832{
		Addr: *addrFlag,
	}
	err := dp.Connect()
	if err != nil {
		log.Fatal(err)
	}

	for range time.Tick(100 * time.Millisecond) {
		for _, CH := range []Channel{Ch1, Ch2, Ch3} {
			m, err := dp.Measure(CH)
			if err != nil {
				panic(err)
			}
			log.Printf("%v", m)
		}
	}
}

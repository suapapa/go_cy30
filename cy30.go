package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/serial"
)

const (
	cmdStartSingleMeasurement      = "$00022123&"
	cmdLights                      = "$0003260130&"
	cmdOpenContinuousMeasurement   = "$00022426&"
	cmdStopContinousMeasurement    = "$0003260029&"
	cmdInstructionConfirmation     = "$00023335&"
	cmdClosedContinuousMeasurement = "$00022123&"
	cmdTurnOffTheLaser             = "$00022123&"

	errCodeDistanceTooShort        = "$00023335&$0006210000001542&"
	errCodeNoEcho                  = "$00023335&$0006210000001643&"
	errCodeReflectionIsTooStrong   = "$00023335&$0006210000001744&"
	errCodeAmbientLightIsTooStrong = "$00023335&$0006210000001845&"

	// errCodeDistanceTooShort = "$00162499990000001500000000000000053&"
	// errCodeNoEcho           = "$00162499990000001600000000000000054&"
	// errCodeRangingError     = "$001624000100118297001182970011829711&"
)

var (
	ErrWrongFormat = errors.New("wrong data format")
	ErrTooClose    = errors.New("distance too close")
	ErrNoSignal    = errors.New("no signal from sensor")
	ErrNoConfirm   = errors.New("no instruction confirmation")
)

type cy30 struct {
	ser     *serial.Port
	bufRead *bufio.Reader
}

func newCy30(port string) *cy30 {
	ser, err := serial.OpenPort(&serial.Config{
		Name:        port,
		Baud:        115200,
		ReadTimeout: 3 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	ser.Flush()

	return &cy30{
		ser:     ser,
		bufRead: bufio.NewReader(ser),
	}
}

func (c *cy30) Close() error {
	return c.ser.Close()
}

func (c *cy30) SingleDistance() (float64, error) {
	retry := 0
loop:

	_, err := c.ser.Write([]byte(cmdStartSingleMeasurement))
	if err != nil {
		return -1, err
	}

	// $00023335&
	first, err := c.bufRead.ReadBytes('&')
	if err != nil {
		return -1, err
	}
	if string(first) != cmdInstructionConfirmation {
		log.Println(string(first))
		return -1, ErrNoConfirm
	}

	//$0006210 00 0027908 &
	second, err := c.bufRead.ReadBytes('&')
	if err == io.EOF && retry < 3 { // Too long??
		// log.Println("retry 2,", retry)
		retry++
		goto loop
	}
	if err != nil {
		return -1, err
	}

	// log.Println(string(first), string(second))

	return convertBytesToDistance(second)
}

func (c *cy30) ContinuousDistance() (chan string, error) {
	ch := make(chan string)
	// TODO
	return ch, nil
}

func convertBytesToDistance(buf []byte) (float64, error) {
	str := string(buf)
	if str[0] != '$' || str[len(buf)-1] != '&' {
		return 0, ErrWrongFormat
	}

	if strings.HasSuffix(str, "0001643&") {
		return 0, ErrNoSignal
	}

	if strings.HasSuffix(str, "0001542&") {
		return 0, ErrTooClose
	}

	last7str := str[len(str)-1-1-7 : len(str)-1-1]
	num, err := strconv.Atoi(last7str)
	if err != nil {
		return 0, err
	}

	return float64(num) * 1e-4, nil
}

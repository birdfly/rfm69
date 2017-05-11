package rfm69

import (
	"bytes"
	"time"

	"github.com/ecc1/gpio"
	"github.com/ecc1/radio"
)

const (
	hwVersion = 0x0204
)

type hwFlavor struct{}

// Name returns the radio's name.
func (f hwFlavor) Name() string {
	return "RFM69HCW"
}

// SPIDevice returns the pathname of the radio's SPI device.
func (f hwFlavor) SPIDevice() string {
	return spiDevice
}

// Speed returns the radio's SPI speed.
func (f hwFlavor) Speed() int {
	return spiSpeed
}

// CustomCS returns the GPIO pin number to use as a custom chip-select for the radio.
func (f hwFlavor) CustomCS() int {
	return customCS
}

// InterruptPin returns the GPIO pin number to use for receive interrupts.
func (f hwFlavor) InterruptPin() int {
	return interruptPin
}

// ReadSingleAddress returns the (identity) encoding of an address for SPI read operations.
func (f hwFlavor) ReadSingleAddress(addr byte) byte {
	return addr
}

// ReadBurstAddress returns the (identity) encoding of an address for SPI burst-read operations.
func (f hwFlavor) ReadBurstAddress(addr byte) byte {
	return addr
}

// WriteSingleAddress returns the encoding of an address for SPI write operations.
func (f hwFlavor) WriteSingleAddress(addr byte) byte {
	return SPIWriteMode | addr
}

// WriteBurstAddress returns the encoding of an address for SPI burst-write operations.
func (f hwFlavor) WriteBurstAddress(addr byte) byte {
	return SPIWriteMode | addr
}

// Radio represents an open radio device.
type Radio struct {
	hw            *radio.Hardware
	resetPin      gpio.OutputPin
	receiveBuffer bytes.Buffer
	stats         radio.Statistics
	err           error
}

// Open opens the radio device.
func Open() *Radio {
	r := &Radio{hw: radio.Open(hwFlavor{})}
	v := r.Version()
	if r.Error() != nil {
		return r
	}
	if v != hwVersion {
		r.hw.Close() // nolint
		r.SetError(radio.HardwareVersionError{Actual: v, Expected: hwVersion})
		return r
	}
	r.resetPin, r.err = gpio.Output(resetPin, false)
	if r.Error() != nil {
		r.hw.Close() // nolint
		return r
	}
	return r
}

// Close closes the radio device.
func (r *Radio) Close() {
	r.setMode(SleepMode)
	r.hw.Close()
}

// Version returns the radio's hardware version.
func (r *Radio) Version() uint16 {
	v := r.hw.ReadRegister(RegVersion)
	return uint16(v>>4)<<8 | uint16(v&0xF)
}

// Reset resets the radio device.  See section 7.2.2 of data sheet.
func (r *Radio) Reset() {
	if r.Error() != nil {
		return
	}
	r.resetPin.Write(true) // nolint
	time.Sleep(100 * time.Microsecond)
	r.err = r.resetPin.Write(false)
	time.Sleep(5 * time.Millisecond)
}

// Init initializes the radio device.
func (r *Radio) Init(frequency uint32) {
	r.Reset()
	r.InitRF(frequency)
	r.setMode(SleepMode)
}

// Statistics returns the byte and packet counts for the radio device.
func (r *Radio) Statistics() radio.Statistics {
	return r.stats
}

// Error returns the error state of the radio device.
func (r *Radio) Error() error {
	err := r.hw.Error()
	if err != nil {
		return err
	}
	return r.err
}

// SetError sets the error state of the radio device.
func (r *Radio) SetError(err error) {
	r.hw.SetError(err)
	r.err = err
}

// Hardware returns the radio's hardware information.
func (r *Radio) Hardware() *radio.Hardware {
	return r.hw
}

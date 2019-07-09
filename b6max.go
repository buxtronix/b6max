// Package b6max provides for communication to SkyRC iMAX and compatible battery chargers.
//
// Tested on SkyRC iMAX B6V2. May work with others.
package b6max

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/logger"
	"github.com/karalabe/usb"
)

const (
	vendorId  = 0x0000
	productId = 0x0001
)

var startTime time.Time

type packet struct {
	bytes.Buffer
	command byte
	data    []byte
}

func (p *packet) init(pType byte) {
	p.Reset()
	p.command = pType
}

func (p *packet) writeData(data interface{}) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	p.setData(buf.Bytes())
}

func (p *packet) setData(data []byte) {
	p.Reset()
	p.WriteByte(0x0f)
	p.WriteByte(byte(len(data) + 2))
	p.WriteByte(p.command)
	p.Write(data)
	p.WriteByte(p.checksum())
	p.Write([]byte{0xff, 0xff})
}

func (p *packet) checksum() byte {
	var sum byte
	d := p.Bytes()
	for i := byte(2); i < d[1]+1; i++ {
		sum += d[i]
	}
	return sum
}

func (p *packet) checksumOk() error {
	d := p.Bytes()
	if d[0] == 0xf0 && d[1] == 0xff && d[2] == 0xff {
		return nil // Ack packet, no checksum.
	}
	wantSum := d[d[1]+1]
	if got, want := p.checksum(), wantSum; got != want {
		return fmt.Errorf("checksum mismatch calc=%d want=%d", got, want)
	}
	return nil
}

// B6Max represents a B6Max device.
type B6Max struct {
	// UsbInfo is the underlying USB information.
	UsbInfo usb.DeviceInfo
	// UsbDevice is the opened USB connection.
	UsbDevice usb.Device
}

// Find searches for connected B6Max devices.
func Find() ([]*B6Max, error) {
	devices := []*B6Max{}
	hids, err := usb.EnumerateHid(vendorId, productId)
	if err != nil {
		return nil, fmt.Errorf("cannot enumerate HID devices: %v", err)
	}

	for _, hid := range hids {
		devices = append(devices, &B6Max{UsbInfo: hid})
	}
	return devices, nil
}

// Open opens the USB connection.
func (b *B6Max) Open() error {
	d, err := b.UsbInfo.Open()
	if err != nil {
		return fmt.Errorf("error opening hid device '%s': %v", b.UsbInfo.Path, err)
	}
	b.UsbDevice = d
	if startTime.IsZero() {
		startTime = time.Now()
	}
	return nil
}

// OpenFirst finds and opens the first available device.
func OpenFirst() (*B6Max, error) {
	all, err := Find()
	if err != nil {
		return nil, err
	}
	if len(all) < 1 {
		return nil, errors.New("No B6Max devices found")
	}
	if err := all[0].Open(); err != nil {
		return nil, err
	}
	return all[0], nil
}

const (
	replyLength  = 64
	sendAttempts = 3
)

func (b *B6Max) writeCommand(cmd *packet) ([]byte, error) {
	rCh := make(chan []byte)
	go func() {
		resp := make([]byte, replyLength)
		if _, err := b.UsbDevice.Read(resp); err != nil {
			logger.Infof("read error: %v", err)
			return
		}
		logger.Infof("Read: %#v", resp)
		rCh <- resp
	}()
	defer close(rCh)
	for attempts := 0; attempts < sendAttempts; attempts++ {
		logger.Infof("Write: %#v", cmd.Bytes())
		_, err := cmd.WriteTo(b.UsbDevice)
		if err != nil {
			return nil, err
		}
		select {
		case data := <-rCh:
			return data, nil
		case <-time.After(time.Second):
			fmt.Printf("Didn't get reply, trying again...")
		}
	}
	return nil, errors.New("timeout in command")
}

// Close closes the USB connection.
func (b *B6Max) Close() error {
	return b.UsbDevice.Close()
}

// ReadSysInfo reads the system config information from the device.
func (b *B6Max) ReadSysInfo() (*SysInfo, error) {
	si := &SysInfo{}
	resp, err := b.writeCommand(si.encode())
	if err != nil {
		return nil, err
	}
	if err := si.parse(resp); err != nil {
		return nil, err
	}
	return si, nil
}

// ReadDevInfo reads device hardware and software parameters.
func (b *B6Max) ReadDevInfo() (*DevInfo, error) {
	di := &DevInfo{}
	resp, err := b.writeCommand(di.encode())
	if err != nil {
		return nil, err
	}
	if err := di.parse(resp); err != nil {
		return nil, err
	}
	return di, nil
}

// ReadProgramState reads the current (dis)charging status.
func (b *B6Max) ReadProgramState() (*ProgramState, error) {
	ps := &ProgramState{}
	resp, err := b.writeCommand(ps.encode())
	if err != nil {
		return nil, err
	}
	if err := ps.parse(resp); err != nil {
		return nil, err
	}
	return ps, nil
}

// Stop sends a stop command to the device.
func (b *B6Max) Stop() (*ProgramStop, error) {
	ps := &ProgramStop{}
	resp, err := b.writeCommand(ps.encode())
	if err != nil {
		return nil, err
	}
	if err := ps.parse(resp); err != nil {
		return nil, err
	}
	return ps, nil
}

// Start tells the device to start a program per the provided parameters.
func (b *B6Max) Start(p *ProgramStart) error {
	resp, err := b.writeCommand(p.encode())
	if err != nil {
		return err
	}
	return p.parse(resp)
}

func getInt(d []byte, n int) int {
	return int(d[n])*256 + int(d[n+1])
}

func setInt(d []byte, v int) {
	d[0], d[1] = byte(v/256), byte(v%256)
}

// SysInfo represents system configuration information.
type SysInfo struct {
	_ uint8 // Start.
	_ uint8 // Length.
	_ uint8 // Command.
	_ uint8 // Pad.
	// CycleTime is the rest time between charge->discharge cycles (minutes, max 60).
	CycleTime uint8
	// TimeLimitOn is whether the time limit is enabled.
	TimeLimitOn uint8
	// TimeLimit is the maximum duration of a program (minutes).
	TimeLimit uint16
	// CapLimitOn is whether the Cap limit is enabled.
	CapLimitOn uint8
	// CapLimit is the maximum charge mAh for a program.
	CapLimit uint16
	// KeyBuzz is whether the keypress sound is enabled.
	KeyBuzz uint8
	// SysBuzz is whether system sounds are enabled.
	SysBuzz uint8
	// InDCLow is the lower DC input voltage limit.
	InDCLowLimit uint16
	// TempLimit is the maximum allowed operating temperature (Celcius).
	TemperatureLimit uint8
	// Voltage represents the current voltage of ??
	Millivolts uint16
	// Cells are the current voltages of the cells at the balance port.
	Cells [6]uint16
}

func (s *SysInfo) parse(data []byte) error {
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, s); err != nil {
		return err
	}
	return nil
}

// Encode returns data to request SysInfo.
func (s *SysInfo) encode() *packet {
	p := &packet{}
	p.init(0x5a)
	p.setData([]byte{0x0})
	return p
}

// String returns the string format of SysInfo returned data.
func (s *SysInfo) String() string {
	ret := []string{
		fmt.Sprintf("Time Limit     : %dm (enabled: %v)", s.TimeLimit, s.TimeLimitOn == 0x1),
		fmt.Sprintf("Capacity Limit : %dmAh (enabled: %v)", s.CapLimit, s.CapLimitOn == 0x1),
		fmt.Sprintf("Beeps          : key (%v), system (%v)", s.KeyBuzz == 0x1, s.SysBuzz == 0x1),
		fmt.Sprintf("Vin low limit  : %2.2fv", float64(s.InDCLowLimit)/1000),
		fmt.Sprintf("Temp limit     : %dC", s.TemperatureLimit),
		fmt.Sprintf("Current voltage: %2.2fv", float64(s.Millivolts)/1000),
		fmt.Sprintf("Cell voltages  : C1:%2.2fv C2:%2.2fv C3:%2.2fv C4:%2.2fv C5:%2.2fv C6:%2.2fv", float64(s.Cells[0])/1000, float64(s.Cells[1])/1000, float64(s.Cells[2])/1000, float64(s.Cells[3])/1000, float64(s.Cells[4])/1000, float64(s.Cells[5])/1000),
	}
	return strings.Join(ret, "\n")
}

// DevInfo represents device hardware configuration.
type DevInfo struct {
	_ uint8 // Start.
	_ uint8 // Length.
	_ uint8 // Command.
	_ uint8 // Pad.
	// CoreType is the core type of the device.
	CoreType [6]byte
	// UpgradeType is ??
	UpgradeType uint8
	// IsEncrypt is ??
	IsEncrypt uint8
	// CustomerID is ??
	CustomerID uint16
	// LanguageID is ??
	LanguageID uint8
	// SwVersion is the current software version.
	SwVersion uint16
	// HwVersion is the current hardware version.
	HwVersion uint8
}

func (d *DevInfo) parse(data []byte) error {
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, d); err != nil {
		return err
	}
	return nil
}

// Encode returns data to request DevInfo.
func (d *DevInfo) encode() *packet {
	p := &packet{}
	p.init(0x57)
	p.setData([]byte{0x0})
	return p
}

// String returns the string format of DevInfo returned data.
func (d *DevInfo) String() string {
	ret := []string{
		fmt.Sprintf("Core Type : %s", d.CoreType),
		fmt.Sprintf("Versions  : hw: %d    sw: %d", d.HwVersion, d.SwVersion),
	}
	return strings.Join(ret, "\n")
}

// State is the current operating state.
type State uint8

const (
	_ State = iota
	StateRunning
	StateIdle
	StateFinish
	StateError
)

var stateString = map[State]string{
	StateRunning: "running",
	StateIdle:    "idle",
	StateFinish:  "finished",
	StateError:   "error",
}

func (s State) String() string {
	return stateString[s]
}

// ProgramState represents current program info.
type ProgramState struct {
	_ uint8 // Start.
	_ uint8 // Length.
	_ uint8 // Command.
	_ uint8 // Pad.
	// WorkState is the current device state.
	WorkState State
	// ErrorCode is the error code, if WorkState is StateError.
	// If not errored, the total (dis)charge current at this moment.
	ErrorCodeMah uint16
	// Number of seconds the program has been running.
	Time uint16
	// The voltage at this moment.
	MilliVolt uint16
	// The current at this moment.
	MilliAmp uint16
	// Temperature of the external sensor.
	TemperatureExternal uint8
	// Temperature of the internal sensor.
	TemperatureInternal uint8
	// Impedance of the attached battery.
	Impedance uint16
	// Voltage of the attached cells at the balance port.
	Cells [6]uint16
}

func (p *ProgramState) parse(data []byte) error {
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, p); err != nil {
		return err
	}
	return nil
}

// Encode encodes the ProgramState request.
func (p *ProgramState) encode() *packet {
	pk := &packet{}
	pk.init(0x55)
	pk.setData([]byte{0x0})
	return pk
}

// Header returns a header string for displaying ProgramState.
func (p *ProgramState) Header() string {
	return "State      Time    mAh       Voltage  Current   TempExt  TempInt  Imp    C1     C2      C3      C4     C5     C6"
}

// String returns a string of the current ProgramState.
func (p *ProgramState) String() string {
	if p.WorkState == StateError {
		return fmt.Sprintf("ERROR CODE %d\n", p.ErrorCodeMah)
	}
	return fmt.Sprintf("%-9s  %5d   %5d    %.3fv     %2.2fA     %dC       %dC     %2d\u2126   %.3fv  %.3fv  %.3fv  %.3fv  %.3fv  %.3fv\n",
		p.WorkState, time.Now().Sub(startTime)/time.Second,
		p.ErrorCodeMah, float64(p.MilliVolt)/1000.0, float64(p.MilliAmp)/1000.0,
		p.TemperatureExternal, p.TemperatureInternal, p.Impedance,
		float64(p.Cells[0])/1000, float64(p.Cells[1])/1000, float64(p.Cells[2])/1000, float64(p.Cells[3])/1000, float64(p.Cells[4])/1000, float64(p.Cells[5])/1000)
}

// BatteryType is the battery type connected.
type BatteryType uint8

const (
	LiPo BatteryType = iota
	LiIo
	LiFe
	// LiHv slots in here for newer chargers (changes const values)
	NiMh
	NiCd
	Pb
)

var batteryTypeStrings = map[BatteryType]string{
	LiPo: "LiPo",
	LiIo: "LiIo",
	LiFe: "LiFe",
	NiMh: "NiMh",
	NiCd: "NiCd",
	Pb:   "Pb",
}

func (b BatteryType) String() string {
	return batteryTypeStrings[b]
}

// ProgramType is the requested program type.
type ProgramType uint8

const (
	Charge ProgramType = iota
	Discharge
	Storage
	FastCharge
	Balance
	AutoCharge
	RePeak
	Cycle
)

var programTypeStrings = map[ProgramType]string{
	Charge:     "charge",
	Discharge:  "discharge",
	Storage:    "storage",
	FastCharge: "fastcharge",
	Balance:    "balance",
	AutoCharge: "autocharge",
	RePeak:     "repeak",
	Cycle:      "cycle",
}

func (p ProgramType) String() string {
	return programTypeStrings[p]
}

// Cycle types.
type CycleType uint8

const (
	ChargeDischarge = iota
	DischargeCharge
)

var cycleTypeStrings = map[CycleType]string{
	ChargeDischarge: "charge-discharge",
	DischargeCharge: "discharge-charge",
}

func (c CycleType) String() string {
	return cycleTypeStrings[c]
}

var programTypeMap = map[BatteryType]map[ProgramType]uint8{
	LiPo: map[ProgramType]uint8{
		Charge:     0,
		Discharge:  1,
		Storage:    2,
		FastCharge: 3,
		Balance:    4,
	},
	LiIo: map[ProgramType]uint8{
		Charge:     0,
		Discharge:  1,
		Storage:    2,
		FastCharge: 3,
		Balance:    4,
	},
	LiFe: map[ProgramType]uint8{
		Charge:     0,
		Discharge:  1,
		Storage:    2,
		FastCharge: 3,
		Balance:    4,
	},
	NiMh: map[ProgramType]uint8{
		Charge:     0,
		AutoCharge: 1,
		Discharge:  2,
		RePeak:     3,
		Cycle:      4,
	},
	NiCd: map[ProgramType]uint8{
		Charge:     0,
		AutoCharge: 1,
		Discharge:  2,
		RePeak:     3,
		Cycle:      4,
	},
	Pb: map[ProgramType]uint8{
		Charge:    0,
		Discharge: 1,
	},
}

// ProgramStart is a request to start a program.
type ProgramStart struct {
	_ uint8 // Pad.
	// The battery type connected.
	BatteryType BatteryType
	// Number of cells.
	Cells byte
	// Program type requested.
	PwmMode ProgramType
	// Maximum charge current requested.
	ChargeCurrent uint16
	// Maximum discharge current requested.
	DischargeCurrent uint16
	// Discharge cutoff voltage, millivolts.
	DischargeCutoff uint16
	// Cutoff voltage for charging.
	// For Ni* batteries, this is the delta-v end-of-charge detection (typically 4mv)
	ChargeCutoff uint16
	// Thie field depends on the charge mode. If RePeak, it's the number of
	// cycles to perform. If Cycle, it's the cycle type.
	RePeakCycleInfo uint8
	// Number of charge/discharge cycles in Cycle mode. 1-5
	CycleCount uint8
	// Trickle charge (mA).
	Trickle uint16
}

// Encode encodes a ProgramStart request.
func (p *ProgramStart) encode() *packet {
	p1 := *p
	p1.PwmMode = ProgramType(programTypeMap[p.BatteryType][p.PwmMode])
	pk := &packet{}
	pk.init(0x05)
	pk.writeData(&p1)
	return pk
}

// Parse parses the response to a request.
func (p *ProgramStart) parse(data []byte) error {
	if len(data) != replyLength {
		return fmt.Errorf("parse(): data length got=%d want=%d bytes", len(data), replyLength)
	}
	return nil
}

// ProgramStop implements a stop program command.
type ProgramStop struct {
}

// Parse parses the command response.
func (p *ProgramStop) parse(data []byte) error {
	if len(data) != replyLength {
		return fmt.Errorf("parse(): data length got=%d want=%d bytes", len(data), replyLength)
	}
	return nil
}

// Encode encodes a command request.
func (p *ProgramStop) encode() *packet {
	pk := &packet{}
	pk.init(0xfe)
	pk.setData([]byte{0x0})
	return pk
}

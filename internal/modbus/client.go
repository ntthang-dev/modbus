package modbus

import (
	"encoding/binary"
	"time"

	"github.com/goburrow/modbus"
)

// Client đại diện cho một client Modbus
type Client struct {
	handler *modbus.RTUClientHandler
	client  modbus.Client
}

// NewClient tạo một client Modbus mới
func NewClient(port string, baudRate int, dataBits int, stopBits int, parity string, slaveID byte) (*Client, error) {
	handler := modbus.NewRTUClientHandler(port)
	handler.BaudRate = baudRate
	handler.DataBits = dataBits
	handler.StopBits = stopBits
	handler.Parity = parity
	handler.Timeout = 2 * time.Second
	handler.SlaveId = slaveID

	err := handler.Connect()
	if err != nil {
		return nil, err
	}

	return &Client{
		handler: handler,
		client:  modbus.NewClient(handler),
	}, nil
}

// Close đóng kết nối
func (c *Client) Close() error {
	return c.handler.Close()
}

// ReadHoldingRegisters đọc các thanh ghi giữ
func (c *Client) ReadHoldingRegisters(address uint16, quantity uint16) ([]byte, error) {
	return c.client.ReadHoldingRegisters(address, quantity)
}

// WriteSingleRegister ghi một thanh ghi
func (c *Client) WriteSingleRegister(address uint16, value uint16) error {
	_, err := c.client.WriteSingleRegister(address, value)
	return err
}

// WriteMultipleRegisters ghi nhiều thanh ghi
func (c *Client) WriteMultipleRegisters(address uint16, values []uint16) error {
	// Chuyển đổi []uint16 thành []byte
	data := make([]byte, len(values)*2)
	for i, v := range values {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	_, err := c.client.WriteMultipleRegisters(address, uint16(len(values)), data)
	return err
}

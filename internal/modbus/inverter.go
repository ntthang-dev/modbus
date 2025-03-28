package modbus

import (
	"encoding/json"
	"time"
)

// InverterData đại diện cho dữ liệu của inverter
type InverterData struct {
	Timestamp time.Time `json:"timestamp"`

	// Tín hiệu kết nối bắt buộc
	ConnectionStatus uint16 `json:"connection_status"` // 0: Mất kết nối, 1: Kết nối
	DeviceStatus     uint16 `json:"device_status"`     // 0: Lỗi, 1: Hoạt động
	ErrorCode        uint16 `json:"error_code"`        // Mã lỗi

	// Tín hiệu giám sát bắt buộc
	ActivePower   float64 `json:"active_power"`   // kW
	ReactivePower float64 `json:"reactive_power"` // kVar
	PowerFactor   float64 `json:"power_factor"`   // Hệ số công suất
	Frequency     float64 `json:"frequency"`      // Hz
	Voltage       float64 `json:"voltage"`        // V
	Current       float64 `json:"current"`        // A
	Temperature   float64 `json:"temperature"`    // °C

	// Tín hiệu giám sát khuyến nghị
	DailyEnergy float64 `json:"daily_energy"` // kWh
	TotalEnergy float64 `json:"total_energy"` // kWh
	Efficiency  float64 `json:"efficiency"`   // %
}

// InverterService xử lý giao tiếp với inverter
type InverterService struct {
	client *Client
}

// NewInverterService tạo một service mới
func NewInverterService(client *Client) *InverterService {
	return &InverterService{
		client: client,
	}
}

// ReadData đọc dữ liệu từ inverter
func (s *InverterService) ReadData() (*InverterData, error) {
	// Đọc tất cả các thanh ghi
	data, err := s.client.ReadHoldingRegisters(0, 13)
	if err != nil {
		return nil, err
	}

	// Chuyển đổi dữ liệu
	inverterData := &InverterData{
		Timestamp: time.Now(),
	}

	// Tín hiệu kết nối bắt buộc
	inverterData.ConnectionStatus = uint16(data[0])<<8 | uint16(data[1])
	inverterData.DeviceStatus = uint16(data[2])<<8 | uint16(data[3])
	inverterData.ErrorCode = uint16(data[4])<<8 | uint16(data[5])

	// Tín hiệu giám sát bắt buộc
	inverterData.ActivePower = float64(uint16(data[6])<<8|uint16(data[7])) / 100.0   // kW
	inverterData.ReactivePower = float64(uint16(data[8])<<8|uint16(data[9])) / 100.0 // kVar
	inverterData.PowerFactor = float64(uint16(data[10])<<8|uint16(data[11])) / 100.0
	inverterData.Frequency = float64(uint16(data[12])<<8|uint16(data[13])) / 10.0   // Hz
	inverterData.Voltage = float64(uint16(data[14])<<8|uint16(data[15])) / 10.0     // V
	inverterData.Current = float64(uint16(data[16])<<8|uint16(data[17])) / 10.0     // A
	inverterData.Temperature = float64(uint16(data[18])<<8|uint16(data[19])) / 10.0 // °C

	// Tín hiệu giám sát khuyến nghị
	inverterData.DailyEnergy = float64(uint16(data[20])<<8|uint16(data[21])) / 10.0 // kWh
	inverterData.TotalEnergy = float64(uint16(data[22])<<8|uint16(data[23])) / 10.0 // kWh
	inverterData.Efficiency = float64(uint16(data[24])<<8|uint16(data[25])) / 100.0 // %

	return inverterData, nil
}

// ToJSON chuyển đổi dữ liệu sang JSON
func (d *InverterData) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

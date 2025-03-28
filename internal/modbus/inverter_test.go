package modbus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// InverterRegisters định nghĩa các thanh ghi của inverter theo yêu cầu EVN
type InverterRegisters struct {
	// Tín hiệu kết nối bắt buộc
	ConnectionStatus uint16 // Thanh ghi 0: Trạng thái kết nối (0: Mất kết nối, 1: Kết nối)
	DeviceStatus     uint16 // Thanh ghi 1: Trạng thái thiết bị (0: Lỗi, 1: Hoạt động)
	ErrorCode        uint16 // Thanh ghi 2: Mã lỗi

	// Tín hiệu giám sát bắt buộc
	ActivePower   uint16 // Thanh ghi 3: Công suất tác dụng (kW)
	ReactivePower uint16 // Thanh ghi 4: Công suất phản kháng (kVar)
	PowerFactor   uint16 // Thanh ghi 5: Hệ số công suất
	Frequency     uint16 // Thanh ghi 6: Tần số (Hz)
	Voltage       uint16 // Thanh ghi 7: Điện áp (V)
	Current       uint16 // Thanh ghi 8: Dòng điện (A)
	Temperature   uint16 // Thanh ghi 9: Nhiệt độ (°C)

	// Tín hiệu giám sát khuyến nghị
	DailyEnergy uint16 // Thanh ghi 10: Điện năng ngày (kWh)
	TotalEnergy uint16 // Thanh ghi 11: Điện năng tổng (kWh)
	Efficiency  uint16 // Thanh ghi 12: Hiệu suất (%)
}

// TestInverterCommunication kiểm tra giao tiếp với inverter
func TestInverterCommunication(t *testing.T) {
	// Khởi tạo client với simulator
	client, err := NewClient("COM3", 9600, 8, 1, "N", 1)
	if err != nil {
		t.Skipf("Bỏ qua test vì không thể kết nối đến simulator: %v", err)
	}
	defer client.Close()

	// Đợi một chút để đảm bảo kết nối ổn định
	time.Sleep(1 * time.Second)

	// Test 1: Kiểm tra kết nối
	t.Run("Kiểm tra kết nối", func(t *testing.T) {
		// Đọc trạng thái kết nối
		data, err := client.ReadHoldingRegisters(0, 1)
		if err != nil {
			t.Fatalf("Lỗi đọc trạng thái kết nối: %v", err)
		}
		assert.Equal(t, uint16(1), uint16(data[0])<<8|uint16(data[1]), "Trạng thái kết nối phải là 1")
	})

	// Test 2: Đọc tất cả các tín hiệu bắt buộc
	t.Run("Đọc tín hiệu bắt buộc", func(t *testing.T) {
		// Đọc 10 thanh ghi đầu tiên
		data, err := client.ReadHoldingRegisters(0, 10)
		if err != nil {
			t.Fatalf("Lỗi đọc tín hiệu bắt buộc: %v", err)
		}

		// Kiểm tra dữ liệu đọc được
		registers := InverterRegisters{
			ConnectionStatus: uint16(data[0])<<8 | uint16(data[1]),
			DeviceStatus:     uint16(data[2])<<8 | uint16(data[3]),
			ErrorCode:        uint16(data[4])<<8 | uint16(data[5]),
			ActivePower:      uint16(data[6])<<8 | uint16(data[7]),
			ReactivePower:    uint16(data[8])<<8 | uint16(data[9]),
			PowerFactor:      uint16(data[10])<<8 | uint16(data[11]),
			Frequency:        uint16(data[12])<<8 | uint16(data[13]),
			Voltage:          uint16(data[14])<<8 | uint16(data[15]),
			Current:          uint16(data[16])<<8 | uint16(data[17]),
			Temperature:      uint16(data[18])<<8 | uint16(data[19]),
		}

		// Kiểm tra các giá trị hợp lệ
		assert.Equal(t, uint16(1), registers.ConnectionStatus, "Trạng thái kết nối phải là 1")
		assert.Equal(t, uint16(1), registers.DeviceStatus, "Trạng thái thiết bị phải là 1")
		assert.Equal(t, uint16(0), registers.ErrorCode, "Không có lỗi")
		assert.Greater(t, registers.ActivePower, uint16(0), "Công suất tác dụng phải lớn hơn 0")
		assert.Greater(t, registers.ReactivePower, uint16(0), "Công suất phản kháng phải lớn hơn 0")
		assert.Greater(t, registers.PowerFactor, uint16(0), "Hệ số công suất phải lớn hơn 0")
		assert.Equal(t, uint16(50), registers.Frequency, "Tần số phải là 50 Hz")
		assert.Greater(t, registers.Voltage, uint16(0), "Điện áp phải lớn hơn 0")
		assert.Greater(t, registers.Current, uint16(0), "Dòng điện phải lớn hơn 0")
		assert.Greater(t, registers.Temperature, uint16(0), "Nhiệt độ phải lớn hơn 0")
	})

	// Test 3: Đọc tín hiệu khuyến nghị
	t.Run("Đọc tín hiệu khuyến nghị", func(t *testing.T) {
		// Đọc 3 thanh ghi khuyến nghị
		data, err := client.ReadHoldingRegisters(10, 3)
		if err != nil {
			t.Fatalf("Lỗi đọc tín hiệu khuyến nghị: %v", err)
		}

		// Kiểm tra dữ liệu đọc được
		dailyEnergy := uint16(data[0])<<8 | uint16(data[1])
		totalEnergy := uint16(data[2])<<8 | uint16(data[3])
		efficiency := uint16(data[4])<<8 | uint16(data[5])

		// Kiểm tra các giá trị hợp lệ
		assert.GreaterOrEqual(t, dailyEnergy, uint16(0), "Điện năng ngày phải >= 0")
		assert.GreaterOrEqual(t, totalEnergy, uint16(0), "Điện năng tổng phải >= 0")
		assert.Greater(t, efficiency, uint16(0), "Hiệu suất phải > 0")
		assert.LessOrEqual(t, efficiency, uint16(100), "Hiệu suất phải <= 100%")
	})
}

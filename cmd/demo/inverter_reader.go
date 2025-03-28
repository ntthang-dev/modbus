package main

import (
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

// Cấu trúc dữ liệu inverter
type InverterData struct {
	// Các thông số cơ bản
	ActivePower   float64 // Công suất tác dụng (kW)
	ReactivePower float64 // Công suất phản kháng (kVar)
	PowerFactor   float64 // Hệ số công suất
	Frequency     float64 // Tần số (Hz)
	Voltage       float64 // Điện áp (V)
	Current       float64 // Dòng điện (A)
	Temperature   float64 // Nhiệt độ (°C)
	DailyEnergy   float64 // Sản lượng ngày (kWh)
	TotalEnergy   float64 // Sản lượng tổng (kWh)
}

// Hàm kết nối với inverter
func connectInverter(port string, baudRate int, slaveID byte) (modbus.Client, error) {
	// Cấu hình kết nối
	handler := modbus.NewRTUClientHandler(port)
	handler.BaudRate = baudRate
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = slaveID
	handler.Timeout = 2 * time.Second

	// Thử kết nối
	err := handler.Connect()
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối: %v", err)
	}

	return modbus.NewClient(handler), nil
}

// Hàm đọc dữ liệu từ inverter
func readInverterData(client modbus.Client) (*InverterData, error) {
	// Đọc khối thanh ghi (giả sử bắt đầu từ địa chỉ 3000)
	results, err := client.ReadHoldingRegisters(3000, 9)
	if err != nil {
		return nil, fmt.Errorf("lỗi đọc dữ liệu: %v", err)
	}

	// Chuyển đổi dữ liệu
	data := &InverterData{
		ActivePower:   float64(uint16(results[0])<<8|uint16(results[1])) / 10.0,   // kW
		ReactivePower: float64(uint16(results[2])<<8|uint16(results[3])) / 10.0,   // kVar
		PowerFactor:   float64(uint16(results[4])<<8|uint16(results[5])) / 1000.0, // PF
		Frequency:     float64(uint16(results[6])<<8|uint16(results[7])) / 100.0,  // Hz
		Voltage:       float64(uint16(results[8])<<8|uint16(results[9])) / 10.0,   // V
		Current:       float64(uint16(results[10])<<8|uint16(results[11])) / 10.0, // A
		Temperature:   float64(uint16(results[12])<<8 | uint16(results[13])),      // °C
		DailyEnergy:   float64(uint16(results[14])<<8|uint16(results[15])) / 10.0, // kWh
		TotalEnergy:   float64(uint16(results[16])<<8|uint16(results[17])) / 10.0, // kWh
	}

	return data, nil
}

func main() {
	// Thông số kết nối
	port := "COM3"     // Thay đổi theo cổng COM thực tế
	baudRate := 9600   // Thay đổi theo tốc độ baud của inverter
	slaveID := byte(1) // Thay đổi theo địa chỉ slave của inverter

	// Kết nối với inverter
	fmt.Printf("Đang kết nối đến inverter qua cổng %s...\n", port)
	client, err := connectInverter(port, baudRate, slaveID)
	if err != nil {
		log.Fatalf("Không thể kết nối: %v", err)
	}

	// Vòng lặp đọc dữ liệu
	for {
		// Đọc dữ liệu
		data, err := readInverterData(client)
		if err != nil {
			log.Printf("Lỗi đọc dữ liệu: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Hiển thị dữ liệu
		fmt.Printf("\n=== Thông số Inverter ===\n")
		fmt.Printf("Công suất tác dụng: %.1f kW\n", data.ActivePower)
		fmt.Printf("Công suất phản kháng: %.1f kVar\n", data.ReactivePower)
		fmt.Printf("Hệ số công suất: %.3f\n", data.PowerFactor)
		fmt.Printf("Tần số: %.1f Hz\n", data.Frequency)
		fmt.Printf("Điện áp: %.1f V\n", data.Voltage)
		fmt.Printf("Dòng điện: %.1f A\n", data.Current)
		fmt.Printf("Nhiệt độ: %.1f °C\n", data.Temperature)
		fmt.Printf("Sản lượng ngày: %.1f kWh\n", data.DailyEnergy)
		fmt.Printf("Sản lượng tổng: %.1f kWh\n", data.TotalEnergy)

		time.Sleep(1 * time.Second)
	}
}

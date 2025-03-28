package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goburrow/modbus"
)

func main() {
	// Khởi tạo logger
	logger := log.New(os.Stdout, "[InverterSimulator] ", log.LstdFlags)

	// Cấu hình kết nối Modbus
	handler := modbus.NewRTUClientHandler("COM1")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.StopBits = 1
	handler.Parity = "N"
	handler.Timeout = 2 * time.Second

	// Kết nối
	err := handler.Connect()
	if err != nil {
		logger.Fatalf("Không thể kết nối đến cổng COM1: %v", err)
	}
	defer handler.Close()

	// Tạo client Modbus
	client := modbus.NewClient(handler)

	// Khởi tạo dữ liệu mẫu
	registers := make([]byte, 26) // 13 thanh ghi * 2 bytes
	registers[0] = 0x00
	registers[1] = 0x01 // ConnectionStatus = 1
	registers[2] = 0x00
	registers[3] = 0x01 // DeviceStatus = 1
	registers[4] = 0x00
	registers[5] = 0x00 // ErrorCode = 0
	registers[6] = 0x00
	registers[7] = 0x05 // ActivePower = 5 kW
	registers[8] = 0x00
	registers[9] = 0x02 // ReactivePower = 2 kVar
	registers[10] = 0x00
	registers[11] = 0x96 // PowerFactor = 0.95
	registers[12] = 0x00
	registers[13] = 0x32 // Frequency = 50 Hz
	registers[14] = 0x00
	registers[15] = 0xF0 // Voltage = 240 V
	registers[16] = 0x00
	registers[17] = 0x14 // Current = 20 A
	registers[18] = 0x00
	registers[19] = 0x28 // Temperature = 40°C
	registers[20] = 0x00
	registers[21] = 0x32 // DailyEnergy = 50 kWh
	registers[22] = 0x00
	registers[23] = 0xFA // TotalEnergy = 250 kWh
	registers[24] = 0x00
	registers[25] = 0x64 // Efficiency = 100%

	logger.Println("Đã khởi động Inverter Simulator")
	logger.Println("Cấu hình:")
	logger.Println("- Cổng: COM1")
	logger.Println("- Baud rate: 9600")
	logger.Println("- Data bits: 8")
	logger.Println("- Stop bits: 1")
	logger.Println("- Parity: N")

	// Vòng lặp chính để cập nhật dữ liệu
	go func() {
		for {
			// Cập nhật dữ liệu mẫu
			// TODO: Implement logic cập nhật dữ liệu theo thời gian thực
			time.Sleep(1 * time.Second)
		}
	}()

	// Xử lý tín hiệu dừng
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Chờ tín hiệu dừng
	<-sigChan
	logger.Println("Đang dừng simulator...")
}

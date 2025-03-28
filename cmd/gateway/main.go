package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"modbus_inverter/internal/modbus"
)

func main() {
	// Khởi tạo logger
	logger := log.New(os.Stdout, "[Gateway] ", log.LstdFlags)

	// Khởi tạo Modbus client
	client, err := modbus.NewClient("COM7", 9600, 8, 1, "N", 1)
	if err != nil {
		logger.Fatalf("Lỗi khởi tạo client: %v", err)
	}
	defer client.Close()

	// Khởi tạo Inverter service
	inverterService := modbus.NewInverterService(client)

	logger.Println("Đã khởi động Gateway")
	logger.Println("Cấu hình:")
	logger.Println("- Cổng: COM1")
	logger.Println("- Baud rate: 9600")
	logger.Println("- Data bits: 8")
	logger.Println("- Stop bits: 1")
	logger.Println("- Parity: N")
	logger.Println("- Địa chỉ inverter: 1")

	// Vòng lặp chính để đọc dữ liệu
	go func() {
		for {
			// Đọc dữ liệu từ inverter
			data, err := inverterService.ReadData()
			if err != nil {
				logger.Printf("Lỗi đọc dữ liệu: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Chuyển đổi sang JSON
			jsonData, err := data.ToJSON()
			if err != nil {
				logger.Printf("Lỗi chuyển đổi JSON: %v", err)
				continue
			}

			// In dữ liệu
			logger.Printf("Dữ liệu từ inverter: %s", string(jsonData))
			time.Sleep(1 * time.Second)
		}
	}()

	// Xử lý tín hiệu dừng
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Chờ tín hiệu dừng
	<-sigChan
	logger.Println("Đang dừng gateway...")
}

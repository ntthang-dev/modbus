package main

import (
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

func main() {
	// Cấu hình kết nối Modbus TCP
	handler := modbus.NewTCPClientHandler("127.0.0.1:502")
	handler.Timeout = 2 * time.Second
	handler.SlaveId = 1

	// Kết nối
	fmt.Println("Đang kết nối đến Modbus TCP server...")
	err := handler.Connect()
	if err != nil {
		log.Fatalf("Không thể kết nối đến server: %v", err)
	}
	defer handler.Close()

	// Tạo client Modbus
	client := modbus.NewClient(handler)

	fmt.Println("Đã kết nối thành công đến simulator!")

	// Vòng lặp đọc dữ liệu
	for {
		// Đọc 3 thanh ghi
		data, err := client.ReadHoldingRegisters(0, 3)
		if err != nil {
			log.Printf("Lỗi đọc dữ liệu: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// In thông tin
		fmt.Printf("\n[%s] Thông tin Inverter:\n", time.Now().Format("15:04:05"))
		fmt.Printf("Công suất: %d W\n", uint16(data[0])<<8|uint16(data[1]))
		fmt.Printf("Điện áp: %.1f V\n", float64(uint16(data[2])<<8|uint16(data[3]))/10)
		fmt.Printf("Tần số: %d Hz\n", uint16(data[4])<<8|uint16(data[5]))

		time.Sleep(1 * time.Second)
	}
}

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goburrow/modbus"
)

// Cấu trúc cấu hình Modbus
type ModbusConfig struct {
	ComPort   string
	BaudRate  int
	DataBits  int
	Parity    string
	StopBits  int
	Timeout   time.Duration
	SlaveId   byte
	StartAddr uint16
	NumRegs   uint16
}

// Hàm nhập cấu hình từ bàn phím
func getConfig() *ModbusConfig {
	reader := bufio.NewReader(os.Stdin)
	config := &ModbusConfig{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   "N",
		StopBits: 1,
		Timeout:  2 * time.Second,
		SlaveId:  1,
	}

	// Nhập cổng COM
	for {
		fmt.Print("Nhập cổng COM (ví dụ: COM3): ")
		comPort, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Lỗi đọc input: %v", err)
			continue
		}
		config.ComPort = strings.TrimSpace(comPort)
		if config.ComPort != "" {
			break
		}
		fmt.Println("Cổng COM không được để trống!")
	}

	// Nhập địa chỉ bắt đầu
	for {
		fmt.Print("Nhập địa chỉ bắt đầu (0-65535): ")
		addrStr, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Lỗi đọc input: %v", err)
			continue
		}
		addr, err := strconv.ParseUint(strings.TrimSpace(addrStr), 10, 16)
		if err != nil {
			fmt.Println("Địa chỉ không hợp lệ!")
			continue
		}
		config.StartAddr = uint16(addr)
		break
	}

	// Nhập số lượng thanh ghi
	for {
		fmt.Print("Nhập số lượng thanh ghi (1-125): ")
		numStr, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Lỗi đọc input: %v", err)
			continue
		}
		num, err := strconv.ParseUint(strings.TrimSpace(numStr), 10, 16)
		if err != nil || num < 1 || num > 125 {
			fmt.Println("Số lượng thanh ghi không hợp lệ (1-125)!")
			continue
		}
		config.NumRegs = uint16(num)
		break
	}

	// Nhập Slave ID
	for {
		fmt.Print("Nhập Slave ID (1-247): ")
		idStr, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Lỗi đọc input: %v", err)
			continue
		}
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 8)
		if err != nil || id < 1 || id > 247 {
			fmt.Println("Slave ID không hợp lệ (1-247)!")
			continue
		}
		config.SlaveId = byte(id)
		break
	}

	return config
}

// Hàm đọc dữ liệu từ thanh ghi
func readRegisters(client modbus.Client, startAddr, numRegs uint16) ([]byte, error) {
	return client.ReadHoldingRegisters(startAddr, numRegs)
}

func main() {
	// Nhập cấu hình
	fmt.Println("=== Cấu hình kết nối Modbus RTU ===")
	config := getConfig()

	// Cấu hình kết nối Modbus RTU
	handler := modbus.NewRTUClientHandler(config.ComPort)
	handler.BaudRate = config.BaudRate
	handler.DataBits = config.DataBits
	handler.Parity = config.Parity
	handler.StopBits = config.StopBits
	handler.Timeout = config.Timeout
	handler.SlaveId = config.SlaveId

	// Kết nối
	fmt.Printf("\nĐang kết nối đến thiết bị qua %s...\n", config.ComPort)
	err := handler.Connect()
	if err != nil {
		log.Fatalf("Không thể kết nối đến thiết bị: %v\nVui lòng kiểm tra:\n1. Cổng COM đã đúng chưa\n2. Cáp RS485 đã kết nối đúng chưa\n3. Thiết bị đã bật nguồn chưa", err)
	}
	defer handler.Close()

	// Tạo client Modbus
	client := modbus.NewClient(handler)

	fmt.Println("Đã kết nối thành công!")

	// Vòng lặp đọc dữ liệu
	for {
		// Đọc dữ liệu
		data, err := readRegisters(client, config.StartAddr, config.NumRegs)
		if err != nil {
			log.Printf("Lỗi đọc dữ liệu: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// In thông tin
		fmt.Printf("\n[%s] Dữ liệu từ địa chỉ %d:\n", time.Now().Format("15:04:05"), config.StartAddr)
		for i := 0; i < len(data); i += 2 {
			addr := config.StartAddr + uint16(i/2)
			value := uint16(data[i])<<8 | uint16(data[i+1])
			fmt.Printf("Địa chỉ %d: %d (0x%04X)\n", addr, value, value)
		}

		time.Sleep(1 * time.Second)
	}
}

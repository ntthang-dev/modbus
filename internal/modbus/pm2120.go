// internal/modbus/pm2120.go
package modbus

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
)

// --- Định nghĩa địa chỉ thanh ghi và kiểu dữ liệu cho PM2120 ---
// ==========================>>> LƯU Ý QUAN TRỌNG <<<==========================
//
// Các địa chỉ (addr...) dưới đây được lấy từ cột 'Register' trong file
// 'Public_PM2xxx_PMC Register List_v1001.txt' và đã được TRỪ ĐI 1
// để chuyển thành địa chỉ 0-based mà thư viện Modbus cần.
//
// Kiểu dữ liệu (FLOAT32, INT64,...) và số lượng thanh ghi (quantity)
// được lấy từ cột 'Data Type' và 'Size (INT16)' trong file đó.
//
// >>> BẠN NÊN KIỂM TRA LẠI CÁC GIÁ TRỊ NÀY VỚI THIẾT BỊ THỰC TẾ <<<
//
// Đặc biệt chú ý:
// 1. Scale Factor: Filekhông cung cấp Scale Factor. Code này đọc giá trị
//    gốc. Bạn cần tự thêm phép nhân/chia Scale Factor sau khi giải mã nếu
//    cần thiết để có đơn vị đúng (ví dụ: đọc Wh, muốn kWh thì chia 1000).
// 2. Thứ tự Byte: Mặc định là BigEndian. Nếu thiết bị dùng LittleEndian,
//    sửa lại các hàm bytesTo... (thay binary.BigEndian -> binary.LittleEndian).
// 3. Power Factor: Kiểu '4Q_FP_PF' không chuẩn, tạm xử lý như FLOAT32.
//    Kiểm tra lại nếu giá trị P
//
// ========================

const (
	// --- Dòng điện (Current) - FLOAT32, 2 registers ---
	addrCurrentA   uint16 = 2999 // Register 3000 trong file 
	addrCurrentB   uint16 = 3001 // Register 3002
	addrCurrentC   uint16 = 3003 // Register 3004
	addrCurrentN   uint16 = 3005 // Register 3006 (Kiểm tra nếu PM2120 có)
	addrCurrentAvg uint16 = 3009 // Register 3010

	// --- Điện áp (Voltage) - FLOAT32, 2 registers ---
	addrVoltageAB   uint16 = 3019 // Register 3020
	addrVoltageBC   uint16 = 3021 // Register 3022
	addrVoltageCA   uint16 = 3023 // Register 3024
	addrVoltageLLAvg uint16 = 3025 // Register 3026
	addrVoltageAN   uint16 = 3027 // Register 3028
	addrVoltageBN   uint16 = 3029 // Register 3030
	addrVoltageCN   uint16 = 3031 // Register 3032
	addrVoltageLNAvg uint16 = 3035 // Register 3036

	// --- Công suất (Power) - FLOAT32, 2 registers ---
	addrActivePowerA    uint16 = 3053 // Register 3054 (kW)
	addrActivePowerB    uint16 = 3055 // Register 3056 (kW)
	addrActivePowerC    uint16 = 3057 // Register 3058 (kW)
	addrActivePowerTotal uint16 = 3059 // Register 3060 (kW)
	addrReactivePowerA   uint16 = 3061 // Register 3062 (kVAR)
	addrReactivePowerB   uint16 = 3063 // Register 3064 (kVAR)
	addrReactivePowerC   uint16 = 3065 // Register 3066 (kVAR)
	addrReactivePowerTotal uint16 = 3067 // Register 3068 (kVAR)
	addrApparentPowerA   uint16 = 3069 // Register 3070 (kVA)
	addrApparentPowerB   uint16 = 3071 // Register 3072 (kVA)
	addrApparentPowerC   uint16 = 3073 // Register 3074 (kVA)
	addrApparentPowerTotal uint16 = 3075 // Register 3076 (kVA)

	// --- Power Factor - Kiểu 4Q_FP_PF (Tạm xử lý như FLOAT32, 2 registers) ---
	// >>> Cần kiểm tra lại cách mã hóa thực tế <<<
	addrPowerFactorA    uint16 = 3077 // Register 3078
	addrPowerFactorB    uint16 = 3079 // Register 3080
	addrPowerFactorC    uint16 = 3081 // Register 3082
	addrPowerFactorTotal uint16 = 3083 // Register 3084

	// --- Tần số (Frequency) - FLOAT32, 2 registers ---
	addrFrequency uint16 = 3109 // Register 3110 (Hz)

	// --- Năng lượng (Energy) - INT64, 4 registers, đơn vị Wh ---
	// >>> Cần chia 1000 nếu muốn đổi sang kWh <<<
	addrActiveEnergyDelivered   uint16 = 3203 // Register 3204 (Wh)
	addrActiveEnergyReceived    uint16 = 3207 // Register 3208 (Wh)
	addrReactiveEnergyDelivered uint16 = 3219 // Register 3220 (VARh)
	addrReactiveEnergyReceived  uint16 = 3223 // Register 3224 (VARh)
	addrApparentEnergyDelivered uint16 = 3235 // Register 3236 (VAh)
	addrApparentEnergyReceived  uint16 = 3239 // Register 3240 (VAh)

	// --- THD (Total Harmonic Distortion) - FLOAT32, 2 registers ---
	addrTHDCurrentA uint16 = 21299 // Register 21300 (%)
	addrTHDCurrentB uint16 = 21301 // Register 21302 (%)
	addrTHDCurrentC uint16 = 21303 // Register 21304 (%)
	addrTHDVoltageAB uint16 = 21321 // Register 21322 (%)
	addrTHDVoltageBC uint16 = 21323 // Register 21324 (%)
	addrTHDVoltageCA uint16 = 21325 // Register 21326 (%)
	addrTHDVoltageAN uint16 = 21329 // Register 21330 (%)
	addrTHDVoltageBN uint16 = 21331 // Register 21332 (%)
	addrTHDVoltageCN uint16 = 21333 // Register 21334 (%)

)

// PM2120Data cấu trúc để lưu trữ dữ liệu đọc được từ PM2120
// Sử dụng con trỏ để phân biệt giá trị 0 và lỗi đọc (nil)
type PM2120Data struct {
	// Dòng điện (A)
	CurrentA   *float32 `json:"current_a,omitempty"`
	CurrentB   *float32 `json:"current_b,omitempty"`
	CurrentC   *float32 `json:"current_c,omitempty"`
	CurrentN   *float32 `json:"current_n,omitempty"` // Nếu có
	CurrentAvg *float32 `json:"current_avg,omitempty"`

	// Điện áp (V)
	VoltageAB   *float32 `json:"voltage_ab,omitempty"`
	VoltageBC   *float32 `json:"voltage_bc,omitempty"`
	VoltageCA   *float32 `json:"voltage_ca,omitempty"`
	VoltageLLAvg *float32 `json:"voltage_ll_avg,omitempty"`
	VoltageAN   *float32 `json:"voltage_an,omitempty"`
	VoltageBN   *float32 `json:"voltage_bn,omitempty"`
	VoltageCN   *float32 `json:"voltage_cn,omitempty"`
	VoltageLNAvg *float32 `json:"voltage_ln_avg,omitempty"`

	// Công suất
	ActivePowerA    *float32 `json:"active_power_a,omitempty"`    // kW
	ActivePowerB    *float32 `json:"active_power_b,omitempty"`    // kW
	ActivePowerC    *float32 `json:"active_power_c,omitempty"`    // kW
	ActivePowerTotal *float32 `json:"active_power_total,omitempty"` // kW
	ReactivePowerA   *float32 `json:"reactive_power_a,omitempty"`   // kVAR
	ReactivePowerB   *float32 `json:"reactive_power_b,omitempty"`   // kVAR
	ReactivePowerC   *float32 `json:"reactive_power_c,omitempty"`   // kVAR
	ReactivePowerTotal *float32 `json:"reactive_power_total,omitempty"`// kVAR
	ApparentPowerA   *float32 `json:"apparent_power_a,omitempty"`   // kVA
	ApparentPowerB   *float32 `json:"apparent_power_b,omitempty"`   // kVA
	ApparentPowerC   *float32 `json:"apparent_power_c,omitempty"`   // kVA
	ApparentPowerTotal *float32 `json:"apparent_power_total,omitempty"`// kVA

	// Power Factor (PF)
	PowerFactorA    *float32 `json:"power_factor_a,omitempty"`
	PowerFactorB    *float32 `json:"power_factor_b,omitempty"`
	PowerFactorC    *float32 `json:"power_factor_c,omitempty"`
	PowerFactorTotal *float32 `json:"power_factor_total,omitempty"`

	// Tần số (Hz)
	Frequency *float32 `json:"frequency,omitempty"`

	// Năng lượng (Đơn vị gốc: Wh, VARh, VAh) - Cần Scale Factor để ra kWh, kVARh, kVAh
	ActiveEnergyDelivered   *int64 `json:"active_energy_delivered_wh,omitempty"`   // Wh
	ActiveEnergyReceived    *int64 `json:"active_energy_received_wh,omitempty"`    // Wh
	ReactiveEnergyDelivered *int64 `json:"reactive_energy_delivered_varh,omitempty"` // VARh
	ReactiveEnergyReceived  *int64 `json:"reactive_energy_received_varh,omitempty"`  // VARh
	ApparentEnergyDelivered *int64 `json:"apparent_energy_delivered_vah,omitempty"`  // VAh
	ApparentEnergyReceived  *int64 `json:"apparent_energy_received_vah,omitempty"`   // VAh

	// THD (%)
	THDCurrentA *float32 `json:"thd_current_a_percent,omitempty"`
	THDCurrentB *float32 `json:"thd_current_b_percent,omitempty"`
	THDCurrentC *float32 `json:"thd_current_c_percent,omitempty"`
	THDVoltageAB *float32 `json:"thd_voltage_ab_percent,omitempty"`
	THDVoltageBC *float32 `json:"thd_voltage_bc_percent,omitempty"`
	THDVoltageCA *float32 `json:"thd_voltage_ca_percent,omitempty"`
	THDVoltageAN *float32 `json:"thd_voltage_an_percent,omitempty"`
	THDVoltageBN *float32 `json:"thd_voltage_bn_percent,omitempty"`
	THDVoltageCN *float32 `json:"thd_voltage_cn_percent,omitempty"`

}

// readAndDecodeFloat32 đọc và giải mã Float32 (2 registers)
func readAndDecodeFloat32(client *RTUClient, slaveID byte, addr uint16, fieldName string) (*float32, error) {
	results, err := client.ReadHoldingRegisters(slaveID, addr, 2)
	if err != nil {
		log.Printf("Lỗi đọc thanh ghi %s (addr: %d): %v", fieldName, addr, err)
		return nil, fmt.Errorf("đọc %s lỗi: %w", fieldName, err)
	}
	if len(results) != 4 {
		log.Printf("Lỗi đọc thanh ghi %s (addr: %d): độ dài kết quả không đúng (%d bytes, cần 4)", fieldName, addr, len(results))
		return nil, fmt.Errorf("đọc %s lỗi: độ dài kết quả không đúng", fieldName)
	}
	val := bytesToFloat32(results) // Giả định BigEndian
	// >>> TODO: Áp dụng Scale Factor nếu cần dựa vào tài liệu <<<
	// Ví dụ: scaleFactor := 0.1; scaledVal := val * float32(scaleFactor); return &scaledVal, nil
	return &val, nil
}

// readAndDecodeInt64 đọc và giải mã Int64 (4 registers)
func readAndDecodeInt64(client *RTUClient, slaveID byte, addr uint16, fieldName string) (*int64, error) {
	results, err := client.ReadHoldingRegisters(slaveID, addr, 4)
	if err != nil {
		log.Printf("Lỗi đọc thanh ghi %s (addr: %d): %v", fieldName, addr, err)
		return nil, fmt.Errorf("đọc %s lỗi: %w", fieldName, err)
	}
	if len(results) != 8 {
		log.Printf("Lỗi đọc thanh ghi %s (addr: %d): độ dài kết quả không đúng (%d bytes, cần 8)", fieldName, addr, len(results))
		return nil, fmt.Errorf("đọc %s lỗi: độ dài kết quả không đúng", fieldName)
	}
	val := bytesToInt64(results) // Giả định BigEndian
	// >>> TODO: Áp dụng Scale Factor nếu cần dựa vào tài liệu <<<
	// Ví dụ: scaleFactor := 0.001; // Wh to kWh
	// scaledVal := float64(val) * scaleFactor;
	// // Cần quyết định trả về kiểu gì (float64 hay làm tròn về int64?)
	// // Hoặc chỉ trả về int64 gốc và xử lý scale ở nơi khác.
	return &val, nil
}

// readAndDecodeUint16 đọc và giải mã Uint16 (1 register)
func readAndDecodeUint16(client *RTUClient, slaveID byte, addr uint16, fieldName string) (*uint16, error) {
	results, err := client.ReadHoldingRegisters(slaveID, addr, 1)
	if err != nil {
		log.Printf("Lỗi đọc thanh ghi %s (addr: %d): %v", fieldName, addr, err)
		return nil, fmt.Errorf("đọc %s lỗi: %w", fieldName, err)
	}
	if len(results) != 2 {
		log.Printf("Lỗi đọc thanh ghi %s (addr: %d): độ dài kết quả không đúng (%d bytes, cần 2)", fieldName, addr, len(results))
		return nil, fmt.Errorf("đọc %s lỗi: độ dài kết quả không đúng", fieldName)
	}
	val := bytesToUint16(results) // Giả định BigEndian
	// >>> TODO: Áp dụng Scale Factor nếu cần <<<
	return &val, nil
}

// ReadPM2120Data đọc dữ liệu từ thiết bị PM2120 dựa trên file TXT
func ReadPM2120Data(client *RTUClient, slaveID byte) (*PM2120Data, error) {
	data := &PM2120Data{}
	var err error
	var errAccumulator error // Biến tích lũy lỗi không nghiêm trọng

	// --- Đọc Dòng điện ---
	data.CurrentA, err = readAndDecodeFloat32(client, slaveID, addrCurrentA, "CurrentA")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.CurrentB, err = readAndDecodeFloat32(client, slaveID, addrCurrentB, "CurrentB")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.CurrentC, err = readAndDecodeFloat32(client, slaveID, addrCurrentC, "CurrentC")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	// data.CurrentN, err = readAndDecodeFloat32(client, slaveID, addrCurrentN, "CurrentN") // Bỏ comment nếu cần đọc
	// if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.CurrentAvg, err = readAndDecodeFloat32(client, slaveID, addrCurrentAvg, "CurrentAvg")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }

	// --- Đọc Điện áp ---
	data.VoltageAB, err = readAndDecodeFloat32(client, slaveID, addrVoltageAB, "VoltageAB")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.VoltageBC, err = readAndDecodeFloat32(client, slaveID, addrVoltageBC, "VoltageBC")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.VoltageCA, err = readAndDecodeFloat32(client, slaveID, addrVoltageCA, "VoltageCA")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.VoltageLLAvg, err = readAndDecodeFloat32(client, slaveID, addrVoltageLLAvg, "VoltageLLAvg")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.VoltageAN, err = readAndDecodeFloat32(client, slaveID, addrVoltageAN, "VoltageAN")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.VoltageBN, err = readAndDecodeFloat32(client, slaveID, addrVoltageBN, "VoltageBN")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.VoltageCN, err = readAndDecodeFloat32(client, slaveID, addrVoltageCN, "VoltageCN")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.VoltageLNAvg, err = readAndDecodeFloat32(client, slaveID, addrVoltageLNAvg, "VoltageLNAvg")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }

	// --- Đọc Công suất ---
	data.ActivePowerA, err = readAndDecodeFloat32(client, slaveID, addrActivePowerA, "ActivePowerA")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ActivePowerB, err = readAndDecodeFloat32(client, slaveID, addrActivePowerB, "ActivePowerB")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ActivePowerC, err = readAndDecodeFloat32(client, slaveID, addrActivePowerC, "ActivePowerC")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ActivePowerTotal, err = readAndDecodeFloat32(client, slaveID, addrActivePowerTotal, "ActivePowerTotal")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ReactivePowerA, err = readAndDecodeFloat32(client, slaveID, addrReactivePowerA, "ReactivePowerA")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ReactivePowerB, err = readAndDecodeFloat32(client, slaveID, addrReactivePowerB, "ReactivePowerB")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ReactivePowerC, err = readAndDecodeFloat32(client, slaveID, addrReactivePowerC, "ReactivePowerC")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ReactivePowerTotal, err = readAndDecodeFloat32(client, slaveID, addrReactivePowerTotal, "ReactivePowerTotal")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ApparentPowerA, err = readAndDecodeFloat32(client, slaveID, addrApparentPowerA, "ApparentPowerA")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ApparentPowerB, err = readAndDecodeFloat32(client, slaveID, addrApparentPowerB, "ApparentPowerB")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ApparentPowerC, err = readAndDecodeFloat32(client, slaveID, addrApparentPowerC, "ApparentPowerC")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ApparentPowerTotal, err = readAndDecodeFloat32(client, slaveID, addrApparentPowerTotal, "ApparentPowerTotal")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }

	// --- Đọc Power Factor ---
	data.PowerFactorA, err = readAndDecodeFloat32(client, slaveID, addrPowerFactorA, "PowerFactorA") // Tạm coi là Float32
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.PowerFactorB, err = readAndDecodeFloat32(client, slaveID, addrPowerFactorB, "PowerFactorB") // Tạm coi là Float32
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.PowerFactorC, err = readAndDecodeFloat32(client, slaveID, addrPowerFactorC, "PowerFactorC") // Tạm coi là Float32
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.PowerFactorTotal, err = readAndDecodeFloat32(client, slaveID, addrPowerFactorTotal, "PowerFactorTotal") // Tạm coi là Float32
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }

	// --- Đọc Tần số ---
	data.Frequency, err = readAndDecodeFloat32(client, slaveID, addrFrequency, "Frequency")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }

	// --- Đọc Năng lượng (INT64, đơn vị Wh/VARh/VAh) ---
	data.ActiveEnergyDelivered, err = readAndDecodeInt64(client, slaveID, addrActiveEnergyDelivered, "ActiveEnergyDelivered")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ActiveEnergyReceived, err = readAndDecodeInt64(client, slaveID, addrActiveEnergyReceived, "ActiveEnergyReceived")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ReactiveEnergyDelivered, err = readAndDecodeInt64(client, slaveID, addrReactiveEnergyDelivered, "ReactiveEnergyDelivered")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ReactiveEnergyReceived, err = readAndDecodeInt64(client, slaveID, addrReactiveEnergyReceived, "ReactiveEnergyReceived")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ApparentEnergyDelivered, err = readAndDecodeInt64(client, slaveID, addrApparentEnergyDelivered, "ApparentEnergyDelivered")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.ApparentEnergyReceived, err = readAndDecodeInt64(client, slaveID, addrApparentEnergyReceived, "ApparentEnergyReceived")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }

	// --- Đọc THD ---
	data.THDCurrentA, err = readAndDecodeFloat32(client, slaveID, addrTHDCurrentA, "THDCurrentA")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.THDCurrentB, err = readAndDecodeFloat32(client, slaveID, addrTHDCurrentB, "THDCurrentB")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.THDCurrentC, err = readAndDecodeFloat32(client, slaveID, addrTHDCurrentC, "THDCurrentC")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.THDVoltageAB, err = readAndDecodeFloat32(client, slaveID, addrTHDVoltageAB, "THDVoltageAB")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.THDVoltageBC, err = readAndDecodeFloat32(client, slaveID, addrTHDVoltageBC, "THDVoltageBC")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.THDVoltageCA, err = readAndDecodeFloat32(client, slaveID, addrTHDVoltageCA, "THDVoltageCA")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.THDVoltageAN, err = readAndDecodeFloat32(client, slaveID, addrTHDVoltageAN, "THDVoltageAN")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.THDVoltageBN, err = readAndDecodeFloat32(client, slaveID, addrTHDVoltageBN, "THDVoltageBN")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }
	data.THDVoltageCN, err = readAndDecodeFloat32(client, slaveID, addrTHDVoltageCN, "THDVoltageCN")
	if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }

	// >>> THÊM CÁC LỆNH GỌI HÀM ĐỌC CHO CÁC THANH GHI KHÁC BẠN ĐÃ THÊM Ở TRÊN <<<
	// Ví dụ:
	// data.SomeOtherValue, err = readAndDecodeUint16(client, slaveID, addrSomeOtherValue, "SomeOtherValue")
	// if err != nil { errAccumulator = fmt.Errorf("%v; %w", errAccumulator, err) }

	return data, errAccumulator
}

// --- Helper Functions for Data Type Conversion ---
// Giả định thứ tự byte là Big Endian. Sửa thành LittleEndian nếu cần.

func bytesToFloat32(b []byte) float32 {
	bits := binary.BigEndian.Uint32(b)
	return math.Float32frombits(bits)
}

func bytesToInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func bytesToUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

// Thêm các hàm bytesToUint64, bytesToInt32, bytesToUint32 nếu cần
func bytesToUint64(b []byte) uint64 {
	if len(b) != 8 {
		log.Printf("Cảnh báo: bytesToUint64 nhận %d bytes, cần 8", len(b))
		return 0
	}
	return binary.BigEndian.Uint64(b)
}

func bytesToUint32(b []byte) uint32 {
	if len(b) != 4 {
		log.Printf("Cảnh báo: bytesToUint32 nhận %d bytes, cần 4", len(b))
		return 0
	}
	return binary.BigEndian.Uint32(b)
}



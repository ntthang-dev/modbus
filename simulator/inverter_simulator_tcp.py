#!/usr/bin/env python3
import asyncio
import logging
import random
from datetime import datetime
from pymodbus.datastore import ModbusSequentialDataBlock
from pymodbus.datastore import ModbusSlaveContext, ModbusServerContext
from pymodbus.server import StartAsyncTcpServer

# Cấu hình logging
logging.basicConfig()
log = logging.getLogger()
log.setLevel(logging.DEBUG)

# Cấu trúc dữ liệu của inverter
class InverterData:
    def __init__(self):
        # Tín hiệu kết nối bắt buộc
        self.connection_status = 1  # Thanh ghi 0
        self.device_status = 1      # Thanh ghi 1
        self.error_code = 0         # Thanh ghi 2

        # Tín hiệu giám sát bắt buộc
        self.active_power = 1000    # Thanh ghi 3: 1.0 kW
        self.reactive_power = 500   # Thanh ghi 4: 0.5 kVar
        self.power_factor = 950     # Thanh ghi 5: 0.95
        self.frequency = 50         # Thanh ghi 6: 50 Hz
        self.voltage = 2200         # Thanh ghi 7: 220V
        self.current = 100          # Thanh ghi 8: 1.0A
        self.temperature = 45       # Thanh ghi 9: 45°C

        # Tín hiệu giám sát khuyến nghị
        self.daily_energy = 5000    # Thanh ghi 10: 5.0 kWh
        self.total_energy = 100000  # Thanh ghi 11: 100.0 kWh
        self.efficiency = 980       # Thanh ghi 12: 98%

    def update_values(self):
        """Cập nhật các giá trị với một chút biến động ngẫu nhiên"""
        # Cập nhật công suất tác dụng (0.8 - 1.2 kW)
        self.active_power = int(1000 + random.uniform(-200, 200))
        
        # Cập nhật công suất phản kháng (0.4 - 0.6 kVar)
        self.reactive_power = int(500 + random.uniform(-100, 100))
        
        # Cập nhật điện áp (215 - 225V)
        self.voltage = int(2200 + random.uniform(-50, 50))
        
        # Cập nhật dòng điện (0.8 - 1.2A)
        self.current = int(100 + random.uniform(-20, 20))
        
        # Cập nhật nhiệt độ (40 - 50°C)
        self.temperature = int(45 + random.uniform(-5, 5))
        
        # Cập nhật điện năng ngày (tăng dần)
        self.daily_energy += int(random.uniform(0, 100))
        
        # Cập nhật điện năng tổng (tăng dần)
        self.total_energy += int(random.uniform(0, 200))
        
        # Cập nhật hiệu suất (95 - 99%)
        self.efficiency = int(980 + random.uniform(-30, 20))

    def get_register_values(self):
        """Trả về danh sách các giá trị thanh ghi"""
        return [
            self.connection_status,
            self.device_status,
            self.error_code,
            self.active_power,
            self.reactive_power,
            self.power_factor,
            self.frequency,
            self.voltage,
            self.current,
            self.temperature,
            self.daily_energy,
            self.total_energy,
            self.efficiency
        ]

async def run_server():
    """Chạy server Modbus TCP"""
    # Khởi tạo dữ liệu inverter
    inverter = InverterData()
    
    # Tạo block dữ liệu Modbus
    block = ModbusSequentialDataBlock(0, inverter.get_register_values())
    
    # Tạo context cho slave
    store = ModbusSlaveContext(
        di=None,    # Discrete Inputs
        co=None,    # Coils
        hr=block,   # Holding Registers
        ir=None     # Input Registers
    )
    
    # Tạo context cho server
    context = ModbusServerContext(slaves=store, single=True)
    
    # Cấu hình server
    server_config = {
        "address": ("127.0.0.1", 502),  # Địa chỉ IP và port
        "timeout": 1                     # Timeout
    }
    
    # Khởi động server
    print(f"Khởi động server Modbus TCP trên {server_config['address']}...")
    server = await StartAsyncTcpServer(
        context=context,
        **server_config
    )
    
    print("Server đã khởi động. Nhấn Ctrl+C để dừng.")
    
    try:
        while True:
            # Cập nhật giá trị mỗi giây
            inverter.update_values()
            block.setValues(0, inverter.get_register_values())
            
            # In thông tin cập nhật
            print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Cập nhật giá trị:")
            print(f"Công suất tác dụng: {inverter.active_power/1000:.2f} kW")
            print(f"Công suất phản kháng: {inverter.reactive_power/1000:.2f} kVar")
            print(f"Điện áp: {inverter.voltage/10:.1f}V")
            print(f"Dòng điện: {inverter.current/100:.2f}A")
            print(f"Nhiệt độ: {inverter.temperature}°C")
            print(f"Điện năng ngày: {inverter.daily_energy/1000:.2f} kWh")
            print(f"Điện năng tổng: {inverter.total_energy/1000:.2f} kWh")
            print(f"Hiệu suất: {inverter.efficiency/10:.1f}%")
            
            await asyncio.sleep(1)
            
    except KeyboardInterrupt:
        print("\nDừng server...")
        server.stop()
        print("Server đã dừng.")

if __name__ == "__main__":
    asyncio.run(run_server()) 
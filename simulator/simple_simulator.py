#!/usr/bin/env python3
import asyncio
import logging
from pymodbus.datastore import ModbusSequentialDataBlock
from pymodbus.datastore import ModbusSlaveContext, ModbusServerContext
from pymodbus.server import StartAsyncTcpServer

# Cấu hình logging
logging.basicConfig()
log = logging.getLogger()
log.setLevel(logging.DEBUG)

async def run_server():
    """Chạy server Modbus TCP đơn giản"""
    # Khởi tạo dữ liệu với 3 giá trị cơ bản
    values = [1000, 2200, 50]  # Công suất (W), Điện áp (V), Tần số (Hz)
    
    # Tạo block dữ liệu Modbus
    block = ModbusSequentialDataBlock(0, values)
    
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
        "address": ("127.0.0.1", 502)  # Địa chỉ IP và port
    }
    
    # Khởi động server
    print(f"Khởi động server Modbus TCP đơn giản trên {server_config['address']}...")
    server = await StartAsyncTcpServer(
        context=context,
        **server_config
    )
    
    print("Server đã khởi động. Nhấn Ctrl+C để dừng.")
    print("Các giá trị mẫu:")
    print("Công suất: 1000W")
    print("Điện áp: 220V")
    print("Tần số: 50Hz")
    
    try:
        while True:
            await asyncio.sleep(1)
            
    except KeyboardInterrupt:
        print("\nDừng server...")
        server.stop()
        print("Server đã dừng.")

if __name__ == "__main__":
    asyncio.run(run_server()) 
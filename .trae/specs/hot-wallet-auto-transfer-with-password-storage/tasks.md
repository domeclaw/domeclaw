# Tasks
- [x] Task 1: เพิ่มการเก็บ password ลงใน pin.json
  - [x] SubTask 1.1: แก้ไข `cmd/picoclaw/internal/wallet/create.go` เพื่อเขียน password ลงใน `workspace/wallets/pin.json`
  - [x] SubTask 1.2: สร้าง logic ในการแยกไฟล์ `pin.json` จาก keystore file
  - [x] SubTask 1.3: ตรวจสอบว่า `pin.json` ถูกสร้างในโฟลเดอร์ `workspace/wallets/`

- [x] Task 2: แก้ไขคำสั่ง transfer ให้ไม่ต้องใส่ PIN และ from address
  - [x] SubTask 2.1: แก้ไข `cmd/picoclaw/internal/wallet/transfer.go` เพื่ออ่าน PIN จาก `pin.json`
  - [x] SubTask 2.2: แก้ไข `cmd/picoclaw/internal/wallet/transfer.go` เพื่อใช้ wallet อันเดียวโดยอัตโนมัติ
  - [x] SubTask 2.3: ลบ parameter from address ออกจากคำสั่ง transfer

- [x] Task 3: แก้ไขคำสั่ง transfertoken ให้ไม่ต้องใส่ PIN และ from address
  - [x] SubTask 3.1: สร้างและแก้ไข `cmd/picoclaw/internal/wallet/transfertoken.go` เพื่ออ่าน PIN จาก `pin.json`
  - [x] SubTask 3.2: แก้ไข `cmd/picoclaw/internal/wallet/transfertoken.go` เพื่อใช้ wallet อันเดียวโดยอัตโนมัติ
  - [x] SubTask 3.3: ลบ parameter from address ออกจากคำสั่ง transfertoken

- [x] Task 4: จำกัดระบบเพื่อใช้ wallet อันเดียว
  - [x] SubTask 4.1: แก้ไข `pkg/wallet/service.go` เพื่อไม่ให้สร้าง wallet หลายอัน
  - [x] SubTask 4.2: เพิ่ม logic ตรวจสอบ wallet อันเดียวในระบบ

- [x] Task 5: ทดสอบระบบ
  - [x] SubTask 5.1: ทดสอบ `wallet create 1234` และตรวจสอบ `pin.json`
  - [x] SubTask 5.2: ทดสอบ `transfer 0xA3570FCDA303F55e0978be450f87F885d80a3758 0.01`
  - [x] SubTask 5.3: ทดสอบ `transfertoken 0xA3570FCDA303F55e0978be450f87F885d80a3758 0.01 0x20c0000000000000000000000000000000000000`

## Task Dependencies
- [Task 2] และ [Task 3] ขึ้นอยู่กับ [Task 1]
- [Task 5] ขึ้นอยู่กับ [Task 1], [Task 2], [Task 3], [Task 4]
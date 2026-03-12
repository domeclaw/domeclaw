# สเปคระบบ Hot Wallet อัตโนมัติ

## Why
เพื่อสร้าง hot wallet ที่ picoclaw สามารถจัดการเองโดยไม่ต้องให้ผู้ใช้ใส่ PIN หรือ from address ทุกครั้ง ปรับปรุงประสิทธิภาพการโอนเงิน

## What Changes
- **BREAKING** เก็บ password ลงในไฟล์ `workspace/wallets/pin.json` หลังจาก `wallet create [PIN]`
- **BREAKING** รีบูทคำสั่ง `transfer` และ `transfertoken` ให้ไม่ต้องใส่ PIN
- **BREAKING** รีบูทคำสั่ง `transfer` และ `transfertoken` ให้ไม่ต้องใส่ from address
- จำกัดระบบเพื่อใช้ wallet อันเดียวเท่านั้น

## Impact
- Affected specs: ความปลอดภัยของ wallet, การจัดการ password
- Affected code: `cmd/picoclaw/internal/wallet/create.go`, `cmd/picoclaw/internal/wallet/transfer.go`, `cmd/picoclaw/internal/wallet/transfertoken.go`, `pkg/wallet/service.go`

## ADDED Requirements
### Requirement: เก็บ password ใน pin.json
ระบบ SHALL เก็บ password ที่ใช้สร้าง wallet ลงในไฟล์ `workspace/wallets/pin.json` หลังจาก `wallet create [PIN]`

#### Scenario: Success case
- **WHEN** ผู้ใช้สั่ง `wallet create 1234`
- **THEN** ระบบสร้าง wallet และเก็บ password "1234" ลงใน `pin.json`

### Requirement: คำสั่ง transfer อัตโนมัติ
ระบบ SHALL อ่าน password จาก `pin.json` และใช้ wallet อันเดียวเมื่อผู้ใช้สั่ง `transfer`

#### Scenario: Success case
- **WHEN** ผู้ใช้สั่ง `transfer 0xA3570FCDA303F55e0978be450f87F885d80a3758 0.01`
- **THEN** ระบบอ่าน PIN จาก `pin.json` ใช้ wallet อันเดียวโอนเงิน

### Requirement: คำสั่ง transfertoken อัตโนมัติ
ระบบ SHALL อ่าน password จาก `pin.json` และใช้ wallet อันเดียวเมื่อผู้ใช้สั่ง `transfertoken`

#### Scenario: Success case
- **WHEN** ผู้ใช้สั่ง `transfertoken 0xA3570FCDA303F55e0978be450f87F885d80a3758 0.01 CLAW`
- **THEN** ระบบอ่าน PIN จาก `pin.json` ใช้ wallet อันเดียวโอนโทเคน

## MODIFIED Requirements
### Requirement: ระบบ wallet
ระบบ SHALL จำกัดการใช้ wallet อันเดียวเท่านั้น

## REMOVED Requirements
### Requirement: การใส่ PIN ใน transfer
**Reason**: ใช้ PIN จาก `pin.json` แทน
**Migration**: ไม่ต้องใส่ PIN ในคำสั่ง transfer อีก

### Requirement: การใส่ from address ใน transfer
**Reason**: ใช้ wallet อันเดียวเท่านั้น
**Migration**: ไม่ต้องใส่ from address ในคำสั่ง transfer อีก
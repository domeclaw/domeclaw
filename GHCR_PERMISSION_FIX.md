# แก้ไขปัญหา Permission Denied เมื่อ Push Docker Image ไปยัง GHCR

## ปัญหา
```
error buildx failed with: ERROR: failed to build: failed to solve: failed to push ghcr.io/domeclaw/domeclaw:0.1.1: denied: permission_denied: write_package
```

## สาเหตุ
1. GitHub Token ไม่มีสิทธิ์เขียน packages
2. Image name ไม่ตรงกับ repository name

## การแก้ไขที่ทำไปแล้ว

### 1. แก้ไข `.goreleaser.yaml`
เปลี่ยน image name จาก `picoclaw` เป็น `domeclaw`:
```yaml
images:
  - "ghcr.io/{{ .Env.GITHUB_REPOSITORY_OWNER }}/domeclaw"
  - "docker.io/{{ .Env.DOCKERHUB_IMAGE_NAME }}"
```

### 2. แก้ไข `.github/workflows/docker-build.yml`
เปลี่ยน image name และเพิ่ม permissions:
```yaml
env:
  GHCR_IMAGE_NAME: ${{ github.repository }}  # ใช้ full repository name

permissions:
  contents: read
  packages: write
  id-token: write
```

## การตั้งค่าที่ต้องทำใน GitHub

### 1. ตั้งค่า GitHub Token Permissions
ไปที่ **Settings** → **Actions** → **General** → **Workflow permissions**
- เลือก **Read and write permissions**
- ติ๊กถูกที่ **Allow GitHub Actions to create and approve pull requests**

### 2. ตั้งค่า Package Permissions (ถ้าจำเป็น)
ไปที่ **Settings** → **Packages** → **General**
- ตรวจสอบว่า **Allow public packages** เปิดใช้งาน (ถ้าเป็น public repo)
- หรือตั้งค่า visibility ให้เหมาะสม

### 3. สร้าง Personal Access Token (ทางเลือก)
ถ้า GITHUB_TOKEN ยังไม่มีสิทธิ์เพียงพอ:

1. ไปที่ **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Generate new token พร้อม scopes:
   - `repo` (Full control of private repositories)
   - `write:packages` (Upload packages to GitHub Package Registry)
   - `delete:packages` (Delete packages from GitHub Package Registry - optional)
3. เก็บ token ไว้ใน Secrets ในชื่อ `GHCR_TOKEN`
4. แก้ไข workflow ให้ใช้ token นี้:

```yaml
- name: Login to GitHub Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GHCR_TOKEN }}
```

## การทดสอบ
1. Push code ขึ้น GitHub
2. ไปที่ **Actions** tab
3. เลือก workflow ที่ต้องการ run
4. ตรวจสอบว่า build และ push สำเร็จ

## ตรวจสอบ Package ที่สร้างแล้ว
ไปที่ **Settings** → **Packages** → **domeclaw**
- ควรเห็น Docker image ที่ push ขึ้นไป
- ตรวจสอบ permissions ของ package

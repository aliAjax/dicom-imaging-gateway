# dicom-imaging-gateway__004

基于 Go 实现的 DICOM 影像交换网关 Web 项目，一款后端服务与运维控制台，完成影像接收、去标识化、路由监控与审计导出。
## 构建镜像

请从**仓库根目录**执行；`benzhi.Dockerfile`、`build_benzhi_docker.sh`、`BENZHI_README.md` 均固定在该目录：

```bash
./build_benzhi_docker.sh <image-name> [linux/amd64|linux/arm64]
```

## 标准命令

```bash
go build ./...     # 编译
go run ./cmd/dicom-gateway   # 启动
go test ./...      # 测试（如有）
```

## 环境

- 基础镜像: golang:1.26
- Go 模块目录: `.`
- 依赖已在镜像构建阶段预下载，容器内离线可用。
- 容器内工作目录: `/app`

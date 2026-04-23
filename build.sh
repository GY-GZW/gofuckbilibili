#!/bin/bash

# 设置项目名称 (可修改)
APP_NAME="gofuckbilibili"

# 输出目录
BUILD_DIR="./build"

# 清理旧的构建文件
if [ -d "$BUILD_DIR" ]; then
    rm -rf "$BUILD_DIR"
    echo "已清理旧的构建目录: $BUILD_DIR"
fi
mkdir -p "$BUILD_DIR"

echo "开始交叉编译..."

# 定义编译函数
# 参数: OS ARCH ARM_VERSION(可选) SUFFIX(可选)
build() {
    local os=$1
    local arch=$2
    local arm_version=${3:-""}
    local suffix=${4:-""}
    
    # 设置文件名
    local filename="${APP_NAME}_${os}_${arch}${suffix}"
    
    # 设置环境变量
    export GOOS=$os
    export GOARCH=$arch
    
    # 如果是 ARM 32位，设置 GOARM
    if [ "$arch" == "arm" ] && [ -n "$arm_version" ]; then
        export GOARM=$arm_version
    else
        unset GOARM
    fi

    echo "正在编译: ${GOOS}/${GOARCH}${GOARM:+ v${GOARM}} -> ${filename}"
    
    # 执行编译
    # -ldflags "-s -w" 用于减小二进制文件大小
    go build -ldflags "-s -w" -o "${BUILD_DIR}/${filename}" .
    
    if [ $? -ne 0 ]; then
        echo "❌ 编译失败: ${GOOS}/${GOARCH}"
    else
        echo "✅ 编译成功: ${filename}"
    fi
}

# --- Linux ---
build linux amd64
build linux 386
build linux arm64
build linux arm 7 # ARM 32位 (v7指令集，兼容性较好)

# --- Windows ---
build windows amd64 "" ".exe"
build windows 386 "" ".exe"
build windows arm64 "" ".exe"
build windows arm 7 ".exe" # Windows on ARM 32位 (较少见，但为了完整性加上)

# --- macOS ---
build darwin amd64
build darwin arm64

echo ""
echo "🎉 所有平台编译完成！文件位于 ${BUILD_DIR} 目录："
ls -lh "$BUILD_DIR"
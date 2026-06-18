#!/bin/bash

echo "=========================================="
echo "   旅游行程预订售后退款系统 - 启动脚本"
echo "=========================================="
echo ""

cd "$(dirname "$0")"

echo "📦 正在安装后端依赖..."
cd backend
go mod tidy
if [ $? -ne 0 ]; then
  echo "❌ 后端依赖安装失败"
  exit 1
fi
echo "✅ 后端依赖安装完成"
cd ..

echo ""
echo "📦 正在安装前端依赖..."
cd frontend
npm install
if [ $? -ne 0 ]; then
  echo "❌ 前端依赖安装失败"
  exit 1
fi
echo "✅ 前端依赖安装完成"
cd ..

echo ""
echo "🚀 启动后端服务 (端口 8080)..."
cd backend
go run main.go &
BACKEND_PID=$!
cd ..

sleep 3

echo "🚀 启动前端服务 (端口 5173)..."
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "=========================================="
echo "   系统启动成功！"
echo "=========================================="
echo ""
echo "🌐 前端地址: http://localhost:5173"
echo "🔌 后端地址: http://localhost:8080"
echo ""
echo "💡 首次使用请在前端页面点击右上角 '初始化测试数据' 按钮"
echo ""
echo "按 Ctrl+C 停止所有服务"

trap "echo ''; echo '🛑 正在停止服务...'; kill $BACKEND_PID 2>/dev/null; kill $FRONTEND_PID 2>/dev/null; wait 2>/dev/null; echo '✅ 服务已停止'; exit 0" INT TERM

wait

# PolyAgent Frontend

基于 React + Vite + TypeScript 构建的现代化AI智能体前端界面。

## 🚀 快速开始

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问应用
# http://localhost:3000
```

## 📋 可用命令

```bash
npm run dev          # 启动开发服务器 (端口3000)
npm run build        # 构建生产版本
npm run preview      # 预览生产版本
npm run lint         # ESLint代码检查
npm run type-check   # TypeScript类型检查
```

## 🔧 配置说明

### API代理配置

前端通过Vite代理将API请求转发到后端:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8082',  // 后端服务地址
      changeOrigin: true,
    }
  }
}
```

### 目录结构

```
src/
├── components/     # React组件
│   ├── chat/      # 聊天相关组件
│   ├── layout/    # 布局组件
│   └── ui/        # 通用UI组件
├── pages/         # 页面组件
├── services/      # API服务
├── stores/        # 状态管理 (Zustand)
└── types/         # TypeScript类型定义
```

## 🎨 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite 4
- **UI组件**: Radix UI + Tailwind CSS
- **状态管理**: Zustand
- **HTTP客户端**: Axios
- **图标**: Lucide React
- **图表**: Recharts

## 🔗 相关文档

- [完整启动指南](../STARTUP.md)
- [后端API文档](../eino-polyagent/README.md)

---

> 💡 **注意**: 前端服务需要后端API服务 (端口8082) 正常运行才能完整工作。
# x-mindflow 前端项目

## 项目简介

x-mindflow 智能发布系统的前端项目，提供拓扑可视化与交互界面。

## 技术栈

- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite
- **图可视化**: AntV G6 v5
- **状态管理**: Pinia
- **样式方案**: Tailwind CSS v4
- **HTTP 客户端**: Axios

## 快速开始

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 项目结构

```
src/
├── components/          # 组件
│   ├── common/         # 通用组件
│   └── business/       # 业务组件
│       └── TopologyGraph.vue
├── views/              # 页面
│   └── TopologyView.vue
├── stores/             # Pinia 状态管理
│   └── topologyStore.ts
├── types/              # TypeScript 类型定义
│   └── topology.ts
├── utils/              # 工具函数
│   └── mockData.ts
├── App.vue
├── main.ts
└── style.css
```

## 功能特性

- ✅ 拓扑图可视化展示
- ✅ 横向树形布局
- ✅ 展示/编辑模式切换
- ✅ 服务节点和 LB 节点渲染
- ✅ 流量百分比自动计算
- ✅ 缩放和平移交互

## 开发规范

请遵循项目根目录的 [CLAUDE.md](../CLAUDE.md) 开发规范。

## 相关文档

- [拓扑可视化原型文档](../docs/prototypes/11-topology-visualization-prototype.md)
- [Issue #11 - 拓扑可视化需求](https://github.com/LiusCraft/x-mindflow/issues/11)

## 许可证

Copyright © 2025 x-mindflow Team

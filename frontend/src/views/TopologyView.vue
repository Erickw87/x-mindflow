<script setup lang="ts">
import { ref } from 'vue';
import { useTopologyStore } from '@/stores/topologyStore';
import TopologyGraph from '@/components/business/TopologyGraph.vue';

const topologyStore = useTopologyStore();

const currentMode = ref<'view' | 'edit'>('view');

function switchMode(mode: 'view' | 'edit') {
  currentMode.value = mode;
  topologyStore.setMode(mode);
}

function handleReset() {
  if (confirm('确定要重置为初始数据吗?')) {
    topologyStore.resetToMockData();
  }
}

function handleSave() {
  alert('保存功能将在后续版本实现\n当前拓扑配置将生成对应的 LB 配置');
}
</script>

<template>
  <div class="topology-view">
    <header class="header">
      <div class="header-left">
        <h1 class="title">拓扑可视化与交互界面</h1>
        <p class="subtitle">智能发布系统 - 原型演示</p>
      </div>
      <div class="header-right">
        <div class="mode-switcher">
          <button
            :class="['mode-btn', { active: currentMode === 'view' }]"
            @click="switchMode('view')"
          >
            <svg
              class="icon"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
            >
              <path
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
            展示模式
          </button>
          <button
            :class="['mode-btn', { active: currentMode === 'edit' }]"
            @click="switchMode('edit')"
          >
            <svg
              class="icon"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
            >
              <path
                d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
            编辑模式
          </button>
        </div>
        <button class="action-btn reset-btn" @click="handleReset">
          <svg
            class="icon"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
          >
            <path
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          重置
        </button>
        <button
          v-if="currentMode === 'edit'"
          class="action-btn save-btn"
          @click="handleSave"
        >
          <svg
            class="icon"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
          >
            <path
              d="M5 13l4 4L19 7"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          保存配置
        </button>
      </div>
    </header>

    <main class="main-content">
      <div class="graph-container">
        <TopologyGraph />
      </div>

      <aside class="info-panel">
        <div class="panel-section">
          <h3 class="panel-title">当前模式</h3>
          <div class="mode-indicator">
            <span
              :class="[
                'mode-badge',
                currentMode === 'view' ? 'view-mode' : 'edit-mode',
              ]"
            >
              {{ currentMode === 'view' ? '展示模式' : '编辑模式' }}
            </span>
          </div>
          <p class="mode-description">
            {{
              currentMode === 'view'
                ? '当前为只读模式,可以查看拓扑结构和流量分布,但不能进行修改。'
                : '当前为编辑模式,可以添加节点、修改 LB 规则、调整流量分配等操作。'
            }}
          </p>
        </div>

        <div class="panel-section">
          <h3 class="panel-title">图例说明</h3>
          <div class="legend-items">
            <div class="legend-item">
              <div class="legend-icon service-icon"></div>
              <span class="legend-text">服务节点</span>
            </div>
            <div class="legend-item">
              <div class="legend-icon lb-icon-small"></div>
              <span class="legend-text">LB 节点(1/4大小)</span>
            </div>
            <div class="legend-item">
              <div class="legend-icon branch-default"></div>
              <span class="legend-text">默认拓扑树(蓝色)</span>
            </div>
            <div class="legend-item">
              <div class="legend-icon branch-v11"></div>
              <span class="legend-text">v1.1 分支(橙色)</span>
            </div>
            <div class="legend-item">
              <div class="legend-icon branch-v12"></div>
              <span class="legend-text">v1.2 分支(绿色)</span>
            </div>
          </div>
        </div>

        <div class="panel-section">
          <h3 class="panel-title">功能说明</h3>
          <ul class="feature-list">
            <li>横向树形布局,支持分支结构</li>
            <li>颜色区分不同拓扑树版本</li>
            <li>LB节点缩小为服务节点的1/4</li>
            <li>流量百分比实时计算</li>
            <li>支持缩放和平移操作</li>
            <li>默认拓扑树始终居中</li>
          </ul>
        </div>
      </aside>
    </main>
  </div>
</template>

<style scoped>
.topology-view {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f8fafc;
  overflow: hidden;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}

.subtitle {
  font-size: 0.875rem;
  color: #64748b;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.mode-switcher {
  display: flex;
  gap: 0.5rem;
  background-color: #f1f5f9;
  padding: 0.25rem;
  border-radius: 8px;
}

.mode-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  background-color: transparent;
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn:hover {
  background-color: #e2e8f0;
}

.mode-btn.active {
  background-color: #ffffff;
  color: #3b82f6;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.icon {
  width: 1.25rem;
  height: 1.25rem;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.reset-btn {
  background-color: #f1f5f9;
  color: #64748b;
}

.reset-btn:hover {
  background-color: #e2e8f0;
}

.save-btn {
  background-color: #3b82f6;
  color: #ffffff;
}

.save-btn:hover {
  background-color: #2563eb;
}

.main-content {
  display: flex;
  flex: 1;
  overflow: hidden;
  gap: 1rem;
  padding: 1rem;
}

.graph-container {
  flex: 1;
  min-width: 0;
  background-color: #ffffff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.info-panel {
  width: 300px;
  background-color: #ffffff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.panel-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.panel-title {
  font-size: 1rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.mode-indicator {
  display: flex;
}

.mode-badge {
  padding: 0.375rem 0.75rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
}

.mode-badge.view-mode {
  background-color: #dbeafe;
  color: #1e40af;
}

.mode-badge.edit-mode {
  background-color: #fef3c7;
  color: #92400e;
}

.mode-description {
  font-size: 0.875rem;
  color: #64748b;
  line-height: 1.5;
  margin: 0;
}

.legend-items {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.legend-icon {
  width: 24px;
  height: 24px;
  border-radius: 4px;
}

.service-icon {
  background-color: #ffffff;
  border: 2px solid #e2e8f0;
}

.lb-icon-small {
  width: 12px;
  height: 12px;
  background-color: #3b82f6;
  border: 2px dashed #2563eb;
}

.branch-default {
  background-color: #3b82f6;
  border: 2px solid #2563eb;
}

.branch-v11 {
  background-color: #f59e0b;
  border: 2px solid #d97706;
}

.branch-v12 {
  background-color: #10b981;
  border: 2px solid #059669;
}

.legend-text {
  font-size: 0.875rem;
  color: #475569;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.feature-list li {
  font-size: 0.875rem;
  color: #64748b;
  padding-left: 1.25rem;
  position: relative;
}

.feature-list li::before {
  content: '•';
  position: absolute;
  left: 0;
  color: #3b82f6;
  font-weight: bold;
}
</style>

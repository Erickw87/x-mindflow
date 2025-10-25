<script setup lang="ts">
import TopologyGraph from '@/components/business/TopologyGraph.vue';
</script>

<template>
  <div class="topology-view">
    <header class="header">
      <div class="header-left">
        <h1 class="title">x-mindflow</h1>
      </div>
    </header>

    <main class="main-content">
      <div class="graph-container">
        <TopologyGraph />
      </div>

      <aside class="info-panel">
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
          <h3 class="panel-title">使用说明</h3>
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

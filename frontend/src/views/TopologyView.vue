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
        
        <!-- 图例说明 - 右下角悬浮 -->
        <div class="legend-overlay">
          <h3 class="legend-title">图例说明</h3>
          <div class="legend-items">
            <div class="legend-item">
              <div class="legend-icon service-icon"></div>
              <span class="legend-text">服务节点</span>
            </div>
            <div class="legend-item">
              <div class="legend-icon lb-icon-small"></div>
              <span class="legend-text">LB 节点</span>
            </div>
            <div class="legend-item">
              <div class="legend-icon status-healthy"></div>
              <span class="legend-text">正常运行</span>
            </div>
            <div class="legend-item">
              <div class="legend-icon status-warning"></div>
              <span class="legend-text">有异常</span>
            </div>
            <div class="legend-item">
              <div class="legend-icon status-error"></div>
              <span class="legend-text">严重异常</span>
            </div>
          </div>
          
          <div class="usage-tips">
            <div class="tip-item">💡 支持缩放和平移操作</div>
            <div class="tip-item">💡 流量百分比实时计算</div>
          </div>
        </div>
      </div>
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
  padding: 1rem;
}

.graph-container {
  flex: 1;
  position: relative;
  min-width: 0;
  background-color: #ffffff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.legend-overlay {
  position: absolute;
  bottom: 20px;
  right: 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 1rem 1.25rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(226, 232, 240, 0.8);
  max-width: 280px;
  z-index: 10;
}

.legend-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 0.75rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.legend-items {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.legend-icon {
  width: 16px;
  height: 16px;
  border-radius: 3px;
  flex-shrink: 0;
}

.service-icon {
  background-color: #ffffff;
  border: 2px solid #64748b;
  border-radius: 50%;
}

.lb-icon-small {
  width: 10px;
  height: 10px;
  background-color: #3b82f6;
  border: 2px solid #2563eb;
  border-radius: 2px;
}

.status-healthy {
  background-color: #10b981;
  border-radius: 50%;
}

.status-warning {
  background-color: #f59e0b;
  border-radius: 50%;
}

.status-error {
  background-color: #ef4444;
  border-radius: 50%;
}

.legend-text {
  font-size: 0.75rem;
  color: #475569;
  line-height: 1.2;
}

.usage-tips {
  padding-top: 0.5rem;
  border-top: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.tip-item {
  font-size: 0.7rem;
  color: #64748b;
  line-height: 1.3;
}
</style>

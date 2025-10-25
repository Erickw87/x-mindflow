<script setup lang="ts">
import { ref, onMounted, watch, onBeforeUnmount } from 'vue';
import { Graph } from '@antv/g6';
import { useTopologyStore } from '@/stores/topologyStore';
import type { ServiceNode, TopologyEdge } from '@/types/topology';

const topologyStore = useTopologyStore();

const containerRef = ref<HTMLDivElement | null>(null);
let graph: Graph | null = null;

// 复制提示相关
const showCopyToast = ref(false);
const copyToastMessage = ref('');

const branchColors = {
  default: {
    node: '#3b82f6',
    nodeStroke: '#2563eb',
    edge: '#60a5fa',
  },
  'v1.1': {
    node: '#f59e0b',
    nodeStroke: '#d97706',
    edge: '#fbbf24',
  },
  'v1.2': {
    node: '#10b981',
    nodeStroke: '#059669',
    edge: '#34d399',
  },
};

const statusColors = {
  healthy: '#10b981',
  warning: '#f59e0b',
  error: '#ef4444',
};

function getBranchColor(branch: string | undefined, type: 'node' | 'nodeStroke' | 'edge' = 'node'): string {
  const branchKey = branch || 'default';
  const colors = branchColors[branchKey as keyof typeof branchColors] || branchColors.default;
  return colors[type];
}

function getStatusColor(status?: string): string {
  return statusColors[status as keyof typeof statusColors] || statusColors.healthy;
}

onMounted(() => {
  if (!containerRef.value) return;

  graph = new Graph({
    container: containerRef.value,
    width: containerRef.value.offsetWidth,
    height: containerRef.value.offsetHeight,
    layout: {
      type: 'dagre',
      rankdir: 'LR',
      nodesep: 80,
      ranksep: 180,
      align: 'UL',
      controlPoints: true,
      sortByCombo: true,
    },
    node: {
      style: {
        labelText: (d: any) => d.id,
        labelPlacement: 'bottom',
        labelFontSize: 14,
        labelOffsetY: 10,
        labelFontWeight: 600,
        labelFill: '#1e293b',
        ports: [],
      },
      palette: {
        type: 'group',
        field: 'cluster',
      },
    },
    edge: {
      style: {
        labelText: (d: any) => d.data?.label || '',
        labelFontSize: 11,
        labelBackground: true,
        endArrow: true,
        lineWidth: 2,
      },
    },
    behaviors: [
      'drag-canvas',
      'zoom-canvas',
      {
        type: 'click-select',
        enable: true,
      },
    ],
    autoFit: 'view',
    padding: [50, 100, 50, 100],
  });

  renderGraph();

  // 添加节点点击事件，支持复制文本
  graph.on('node:click', (event: any) => {
    const nodeData = event.data;
    if (nodeData?.data) {
      const { serviceName, versionLabel, label } = nodeData.data;
      let textToCopy = '';

      if (serviceName) {
        textToCopy = versionLabel ? `${serviceName} (${versionLabel})` : serviceName;
      } else if (label) {
        textToCopy = label;
      }

      if (textToCopy) {
        // 复制到剪贴板
        navigator.clipboard.writeText(textToCopy).then(() => {
          copyToastMessage.value = `已复制: ${textToCopy}`;
          showCopyToast.value = true;
          setTimeout(() => {
            showCopyToast.value = false;
          }, 2000);
        }).catch(err => {
          console.error('复制失败:', err);
          copyToastMessage.value = '复制失败';
          showCopyToast.value = true;
          setTimeout(() => {
            showCopyToast.value = false;
          }, 2000);
        });
      }
    }
  });

  // 添加边点击事件，支持复制流量信息
  graph.on('edge:click', (event: any) => {
    const edgeData = event.data;
    if (edgeData?.data?.label) {
      navigator.clipboard.writeText(edgeData.data.label).then(() => {
        copyToastMessage.value = `已复制: ${edgeData.data.label}`;
        showCopyToast.value = true;
        setTimeout(() => {
          showCopyToast.value = false;
        }, 2000);
      }).catch(err => {
        console.error('复制失败:', err);
      });
    }
  });

  window.addEventListener('resize', handleResize);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize);
  if (graph) {
    graph.destroy();
  }
});

watch(
  () => topologyStore.topologyData,
  () => {
    renderGraph();
  },
  { deep: true }
);

function renderGraph() {
  if (!graph) return;

  const { nodes, edges, combos = [] } = topologyStore.topologyData;

  // 将业务 combos 转换为 G6 可识别的 combos 数据
  const g6Combos = (combos || []).map((combo) => ({
    id: combo.id,
    data: { ...combo },
    style: {
      labelText: combo.label,
      labelPlacement: 'top' as const,
      labelFontSize: 12,
      labelFontWeight: 600,
      fill: 'rgba(148, 163, 184, 0.06)',
      stroke: getBranchColor(combo.branch, 'nodeStroke'),
      lineWidth: 1.5,
      radius: 8,
    },
  }));

  const serviceNodeSize = 90;
  const lbNodeSize = 30; // 缩小为原来的一半

  const g6Nodes = nodes.map((node) => {
    const isService = node.type === 'service';
    const serviceNode = node as ServiceNode;
    const branch = node.branch || 'default';

    if (isService) {
      const versions = serviceNode.versions || [];
      const primaryStatus = versions[0]?.status || 'healthy';
      const primaryColor = getStatusColor(primaryStatus);

      // 从节点 ID 中拆分服务名和版本名
      // 例如: service-a-branch1 → 服务名: service-a, 版本名: branch1
      //      service-b-branch2 → 服务名: service-b, 版本名: branch2
      let serviceName = serviceNode.name; // 默认使用 name 字段
      let versionLabel = '';

      const branchMatch = node.id.match(/^(.+)-branch(\d+)$/);
      if (branchMatch) {
        serviceName = branchMatch[1] as string; // 提取服务名，如 "service-a"
        versionLabel = `branch${branchMatch[2] as string}`; // 提取版本名，如 "branch1"
      }

      return {
        id: node.id,
        type: 'circle', // G6 v5 基础节点类型
        data: {
          ...node,
          serviceName, // 保存服务名
          versionLabel, // 保存版本名
          cluster: 'service',
          branch,
          versions,
        },
        style: {
          size: serviceNodeSize,
          fill: primaryColor,
          stroke: getBranchColor(branch, 'nodeStroke'),
          lineWidth: 3,
          // 节点下方的标签（服务名称）
          labelText: serviceName,
          labelPlacement: 'bottom' as const,
          labelFontSize: 14,
          labelOffsetY: 8,
          labelFill: '#1e293b',
          labelFontWeight: 600,
          // 节点内部的文本（版本名）使用 icon 配置
          iconText: versionLabel,
          iconFontSize: 12,
          iconFill: '#ffffff',
          iconFontWeight: 600,
          // 3D 效果：使用更强烈的阴影和偏移
          shadowColor: 'rgba(0, 0, 0, 0.35)',
          shadowBlur: 15,
          shadowOffsetX: 4,
          shadowOffsetY: 6,
          // 添加渐变效果模拟 3D
          fillOpacity: 0.95,
        },
        // 指定所属 combo（仅在存在时添加该字段）
        ...(serviceNode.comboId ? { combo: serviceNode.comboId } : {}),
      };
    } else {
      // LB 节点：浅蓝色六边形
      return {
        id: node.id,
        type: 'hexagon', // G6 v5 六边形类型
        data: {
          ...node,
          label: 'LB', // LB 标签显示在节点下方
          cluster: 'lb',
          branch,
        },
        style: {
          size: lbNodeSize,
          fill: '#4C9AFF', // 浅蓝色（参考 AWS）
          stroke: '#2684FF', // 深一点的蓝色边框
          lineWidth: 2,
          // 节点下方的标签
          labelText: 'LB',
          labelPlacement: 'bottom' as const,
          labelFontSize: 12,
          labelOffsetY: 6,
          labelFill: '#1e293b',
          labelFontWeight: 600,
          // 节点内部不显示文本
          shadowColor: 'rgba(76, 154, 255, 0.3)',
          shadowBlur: 8,
          shadowOffsetX: 2,
          shadowOffsetY: 2,
        },
        ...(node.comboId ? { combo: node.comboId } : {}),
      };
    }
  });

  const g6Edges = edges.map((edge: TopologyEdge) => {
    const branch = edge.branch || 'default';
    const edgeColor = getBranchColor(branch, 'edge');

    // 格式化流量文本：如果 label 包含 "流量"，保持原样；否则格式化为 "流量：XX%"
    const formattedLabel = edge.label && edge.label.includes('流量')
      ? edge.label.replace(/(\d+)%\s*流量/, '流量：$1%')
      : `流量：${edge.trafficPercentage}%`;

    // 检查是否是隐藏边
    const isHidden = edge.hidden === true;

    return {
      id: edge.id,
      source: edge.source,
      target: edge.target,
      data: {
        ...edge,
        label: formattedLabel,
        branch,
        hidden: isHidden,
      },
      style: {
        labelText: isHidden ? '' : formattedLabel, // 隐藏边不显示标签
        stroke: isHidden ? 'transparent' : edgeColor, // 隐藏边透明
        lineWidth: isHidden ? 0 : 2.5, // 隐藏边宽度为 0
        endArrow: isHidden ? false : true,
        lineDash: isHidden ? [] : [10, 5],
        lineDashOffset: 0,
        shadowColor: isHidden ? 'transparent' : 'rgba(0, 0, 0, 0.1)',
        shadowBlur: isHidden ? 0 : 4,
        opacity: isHidden ? 0 : 1, // 隐藏边完全透明
      },
      // 添加边的动画（隐藏边不需要动画）
      keyShape: {
        lineDashOffset: 0,
      },
    };
  });

  graph.setData({
    nodes: g6Nodes,
    edges: g6Edges,
    combos: g6Combos,
  });

  graph.render();

  // 重新适配视图，使图形居中并占据 80% 宽度
  setTimeout(() => {
    if (!graph) return;

    graph.fitView();

    // 启动连接线的动画效果
    startEdgeAnimation();
  }, 100);
}

// 连接线动画函数
function startEdgeAnimation() {
  if (!graph) return;
  const g = graph as Graph;
  let offset = 0;
  const animate = () => {
    offset += 0.5; // 控制动画速度
    if (offset > 15) offset = 0; // 重置偏移量

    // 更新所有边的虚线偏移
    const edges = g.getEdgeData();
    edges.forEach((edge: any) => {
      g.updateEdgeData([{ id: edge.id, style: { lineDashOffset: -offset } }]);
    });

    requestAnimationFrame(animate);
  };

  animate();
}

function handleResize() {
  if (!graph || !containerRef.value) return;
  const width = containerRef.value.offsetWidth;
  const height = containerRef.value.offsetHeight;
  graph.setSize(width, height);

  // 窗口大小改变后重新适配视图，保持居中
  graph.fitView();
}
</script>

<template>
  <div class="topology-graph-container">
    <div ref="containerRef" class="graph-canvas" />

    <!-- 复制成功提示 Toast -->
    <Transition name="toast">
      <div v-if="showCopyToast" class="copy-toast">
        <span class="toast-icon">✓</span>
        <span class="toast-message">{{ copyToastMessage }}</span>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.topology-graph-container {
  width: 100%;
  height: 100%;
  position: relative;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 8px;
  overflow: hidden;
}

.graph-canvas {
  width: 100%;
  height: 100%;
}

/* 启用 G6 画布内文字选择 */
.graph-canvas :deep(canvas) {
  user-select: text !important;
  -webkit-user-select: text !important;
  -moz-user-select: text !important;
  -ms-user-select: text !important;
}

/* 允许文本节点被选中 */
.graph-canvas :deep(text) {
  user-select: text !important;
  -webkit-user-select: text !important;
  cursor: text !important;
  pointer-events: all !important;
}

/* SVG 元素也启用文本选择 */
.graph-canvas :deep(svg) {
  user-select: text !important;
  -webkit-user-select: text !important;
}

/* 复制提示 Toast */
.copy-toast {
  position: absolute;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(16, 185, 129, 0.95);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  backdrop-filter: blur(10px);
}

.toast-icon {
  font-size: 18px;
  font-weight: bold;
}

.toast-message {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Toast 动画 */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(-50%) translateY(-10px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-10px);
}
</style>

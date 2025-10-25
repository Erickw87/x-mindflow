<script setup lang="ts">
import { ref, onMounted, watch, onBeforeUnmount } from 'vue';
import { Graph } from '@antv/g6';
import { useTopologyStore } from '@/stores/topologyStore';
import type { ServiceNode } from '@/types/topology';

const topologyStore = useTopologyStore();

const containerRef = ref<HTMLDivElement | null>(null);
let graph: Graph | null = null;

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
      ranksep: 200,
      align: 'UL',
      controlPoints: true,
    },
    node: {
      style: {
        labelText: (d: any) => d.id,
        labelPlacement: 'center',
        labelFontSize: 12,
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
    behaviors: ['drag-canvas', 'zoom-canvas'],
    autoFit: 'view',
  });

  renderGraph();

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

  const { nodes, edges } = topologyStore.topologyData;

  const serviceNodeSize = 90;
  const lbNodeSize = 60;

  const g6Nodes = nodes.map((node) => {
    const isService = node.type === 'service';
    const serviceNode = node as ServiceNode;
    const branch = node.branch || 'default';

    if (isService) {
      const versions = serviceNode.versions || [];
      const primaryStatus = versions[0]?.status || 'healthy';
      const primaryColor = getStatusColor(primaryStatus);

      return {
        id: node.id,
        data: {
          ...node,
          label: `${serviceNode.name}\n${serviceNode.totalTrafficPercentage}%`,
          cluster: 'service',
          branch,
          versions,
        },
        style: {
          size: serviceNodeSize,
          fill: primaryColor,
          stroke: getBranchColor(branch, 'nodeStroke'),
          lineWidth: 3,
          labelText: `${serviceNode.name}\n${serviceNode.totalTrafficPercentage}%`,
          labelFontSize: 11,
          labelFill: '#1e293b',
          labelFontWeight: 600,
          shadowColor: 'rgba(0, 0, 0, 0.2)',
          shadowBlur: 8,
          shadowOffsetX: 2,
          shadowOffsetY: 2,
        },
      };
    } else {
      return {
        id: node.id,
        data: {
          ...node,
          label: `LB`,
          cluster: 'lb',
          branch,
        },
        style: {
          size: lbNodeSize,
          fill: getBranchColor(branch, 'node'),
          stroke: getBranchColor(branch, 'nodeStroke'),
          lineWidth: 2,
          lineDash: [5, 5],
          radius: 6,
          labelText: `LB`,
          labelFontSize: 10,
          labelFill: '#ffffff',
          labelFontWeight: 600,
          shadowColor: 'rgba(0, 0, 0, 0.15)',
          shadowBlur: 6,
          shadowOffsetX: 1,
          shadowOffsetY: 1,
        },
      };
    }
  });

  const g6Edges = edges.map((edge) => {
    const branch = edge.branch || 'default';
    const edgeColor = getBranchColor(branch, 'edge');

    return {
      id: edge.id,
      source: edge.source,
      target: edge.target,
      data: {
        ...edge,
        label: edge.label || `${edge.trafficPercentage}%`,
        branch,
      },
      style: {
        labelText: edge.label || `${edge.trafficPercentage}%`,
        stroke: edgeColor,
        lineWidth: 2.5,
        endArrow: true,
        lineDash: [10, 5],
        shadowColor: 'rgba(0, 0, 0, 0.1)',
        shadowBlur: 4,
      },
    };
  });

  graph.setData({
    nodes: g6Nodes,
    edges: g6Edges,
  });

  graph.render();
}

function handleResize() {
  if (!graph || !containerRef.value) return;
  const width = containerRef.value.offsetWidth;
  const height = containerRef.value.offsetHeight;
  graph.setSize(width, height);
}
</script>

<template>
  <div class="topology-graph-container">
    <div ref="containerRef" class="graph-canvas" />
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
</style>

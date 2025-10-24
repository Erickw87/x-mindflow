<script setup lang="ts">
import { ref, onMounted, watch, onBeforeUnmount } from 'vue';
import { Graph } from '@antv/g6';
import { useTopologyStore } from '@/stores/topologyStore';
import type { ServiceNode, LBNode } from '@/types/topology';

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

function getBranchColor(branch: string | undefined, type: 'node' | 'nodeStroke' | 'edge' = 'node'): string {
  const branchKey = branch || 'default';
  const colors = branchColors[branchKey as keyof typeof branchColors] || branchColors.default;
  return colors[type];
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
      nodesep: 60,
      ranksep: 180,
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

  const serviceNodeSize = 180;
  const lbNodeSize = serviceNodeSize / 4;

  const g6Nodes = nodes.map((node) => {
    const isService = node.type === 'service';
    const serviceNode = node as ServiceNode;
    const lbNode = node as LBNode;
    const branch = node.branch || 'default';

    const nodeFill = isService ? '#ffffff' : getBranchColor(branch, 'node');
    const nodeStroke = getBranchColor(branch, 'nodeStroke');

    return {
      id: node.id,
      data: {
        ...node,
        label: isService
          ? `${serviceNode.name}\n${serviceNode.totalTrafficPercentage}% 流量\n${serviceNode.totalInstances} 实例`
          : `LB\n${lbNode.rules.length} 规则`,
        cluster: isService ? 'service' : 'lb',
        branch,
      },
      style: {
        fill: nodeFill,
        stroke: nodeStroke,
        lineWidth: isService ? 2 : 3,
        lineDash: isService ? [] : [5, 5],
        radius: isService ? 8 : 4,
        size: isService ? serviceNodeSize : lbNodeSize,
        labelText: isService
          ? `${serviceNode.name}\n${serviceNode.totalTrafficPercentage}% 流量\n${serviceNode.totalInstances} 实例`
          : `LB\n${lbNode.rules.length} 规则`,
        labelFontSize: isService ? 12 : 10,
        labelFill: isService ? '#1e293b' : '#ffffff',
      },
    };
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
        lineWidth: 2,
        endArrow: true,
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
  background-color: #f8fafc;
  border-radius: 8px;
  overflow: hidden;
}

.graph-canvas {
  width: 100%;
  height: 100%;
}
</style>

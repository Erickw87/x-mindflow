/**
 * 拓扑图状态管理
 */

import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { TopologyData, TopologyMode } from '@/types/topology';
import { generateMockTopologyData } from '@/utils/mockData';

export const useTopologyStore = defineStore('topology', () => {
  const topologyData = ref<TopologyData>(generateMockTopologyData());

  const mode = ref<TopologyMode>('view');

  const isEditMode = computed(() => mode.value === 'edit');

  function setMode(newMode: TopologyMode) {
    mode.value = newMode;
  }

  function updateTopologyData(data: TopologyData) {
    topologyData.value = data;
  }

  function addNode(node: TopologyData['nodes'][0]) {
    topologyData.value.nodes.push(node);
  }

  function removeNode(nodeId: string) {
    topologyData.value.nodes = topologyData.value.nodes.filter(
      (n) => n.id !== nodeId
    );
    topologyData.value.edges = topologyData.value.edges.filter(
      (e) => e.source !== nodeId && e.target !== nodeId
    );
  }

  function addEdge(edge: TopologyData['edges'][0]) {
    topologyData.value.edges.push(edge);
  }

  function removeEdge(edgeId: string) {
    topologyData.value.edges = topologyData.value.edges.filter(
      (e) => e.id !== edgeId
    );
  }

  function resetToMockData() {
    topologyData.value = generateMockTopologyData();
  }

  return {
    topologyData,
    mode,
    isEditMode,
    setMode,
    updateTopologyData,
    addNode,
    removeNode,
    addEdge,
    removeEdge,
    resetToMockData,
  };
});

/**
 * Mock 数据
 * 用于原型演示的模拟数据
 */

import type { TopologyData, ServiceNode, LBNode } from '@/types/topology';

/**
 * 生成示例拓扑数据
 * 场景:服务 A v1.0 → v1.1 的灰度发布,存在分支结构
 */
export function generateMockTopologyData(): TopologyData {
  const nodes: Array<ServiceNode | LBNode> = [
    {
      id: 'entry',
      type: 'service',
      name: '入口网关',
      versions: [
        {
          version: 'v2.0',
          instanceCount: 10,
          color: '#3b82f6',
        },
      ],
      totalTrafficPercentage: 100,
      totalInstances: 10,
      branch: 'default',
    },
    {
      id: 'lb1',
      type: 'lb',
      rules: [
        {
          name: '规则1',
          targetVersions: ['v1.0'],
          trafficPercentage: 70,
          ruleDetails: 'Header: x-version=stable',
        },
        {
          name: '规则2',
          targetVersions: ['v1.1'],
          trafficPercentage: 30,
          ruleDetails: 'Header: x-version=canary',
        },
      ],
      branch: 'default',
    },
    {
      id: 'service-a-branch1',
      type: 'service',
      name: '服务 A',
      versions: [
        {
          version: 'v1.0',
          instanceCount: 14,
          color: '#10b981',
        },
      ],
      totalTrafficPercentage: 70,
      totalInstances: 14,
      branch: 'default',
    },
    {
      id: 'service-a-branch2',
      type: 'service',
      name: '服务 A',
      versions: [
        {
          version: 'v1.1',
          instanceCount: 6,
          color: '#f59e0b',
          isTarget: true,
        },
      ],
      totalTrafficPercentage: 30,
      totalInstances: 6,
      branch: 'v1.1',
    },
    {
      id: 'lb2-branch1',
      type: 'lb',
      rules: [
        {
          name: '规则1',
          targetVersions: ['v2.0'],
          trafficPercentage: 100,
        },
      ],
      branch: 'default',
    },
    {
      id: 'lb2-branch2',
      type: 'lb',
      rules: [
        {
          name: '规则1',
          targetVersions: ['v2.1'],
          trafficPercentage: 100,
        },
      ],
      branch: 'v1.1',
    },
    {
      id: 'service-b-branch1',
      type: 'service',
      name: '服务 B',
      versions: [
        {
          version: 'v2.0',
          instanceCount: 10,
          color: '#8b5cf6',
        },
      ],
      totalTrafficPercentage: 70,
      totalInstances: 10,
      branch: 'default',
    },
    {
      id: 'service-b-branch2',
      type: 'service',
      name: '服务 B',
      versions: [
        {
          version: 'v2.1',
          instanceCount: 5,
          color: '#ef4444',
          isTarget: true,
        },
      ],
      totalTrafficPercentage: 30,
      totalInstances: 5,
      branch: 'v1.1',
    },
    {
      id: 'service-c',
      type: 'service',
      name: '服务 C',
      versions: [
        {
          version: 'v1.0',
          instanceCount: 8,
          color: '#06b6d4',
        },
      ],
      totalTrafficPercentage: 100,
      totalInstances: 8,
      branch: 'default',
    },
  ];

  const edges = [
    {
      id: 'edge-1',
      source: 'entry',
      target: 'lb1',
      trafficPercentage: 100,
      branch: 'default',
    },
    {
      id: 'edge-2',
      source: 'lb1',
      target: 'service-a-branch1',
      trafficPercentage: 70,
      label: '70% 流量',
      branch: 'default',
    },
    {
      id: 'edge-3',
      source: 'lb1',
      target: 'service-a-branch2',
      trafficPercentage: 30,
      label: '30% 流量',
      branch: 'v1.1',
    },
    {
      id: 'edge-4',
      source: 'service-a-branch1',
      target: 'lb2-branch1',
      trafficPercentage: 70,
      branch: 'default',
    },
    {
      id: 'edge-5',
      source: 'service-a-branch2',
      target: 'lb2-branch2',
      trafficPercentage: 30,
      branch: 'v1.1',
    },
    {
      id: 'edge-6',
      source: 'lb2-branch1',
      target: 'service-b-branch1',
      trafficPercentage: 70,
      label: '70% 流量',
      branch: 'default',
    },
    {
      id: 'edge-7',
      source: 'lb2-branch2',
      target: 'service-b-branch2',
      trafficPercentage: 30,
      label: '30% 流量',
      branch: 'v1.1',
    },
    {
      id: 'edge-8',
      source: 'service-b-branch1',
      target: 'service-c',
      trafficPercentage: 70,
      branch: 'default',
    },
    {
      id: 'edge-9',
      source: 'service-b-branch2',
      target: 'service-c',
      trafficPercentage: 30,
      branch: 'v1.1',
    },
  ];

  return { nodes, edges };
}

/**
 * 计算节点的总流量百分比
 */
export function calculateTrafficPercentage(
  nodeId: string,
  data: TopologyData
): number {
  const incomingEdges = data.edges.filter((edge) => edge.target === nodeId);

  if (incomingEdges.length === 0) {
    return 100;
  }

  return incomingEdges.reduce(
    (sum, edge) => {
      const sourceNode = data.nodes.find((n) => n.id === edge.source);
      if (!sourceNode) return sum;

      const sourceTraffic =
        sourceNode.type === 'service'
          ? sourceNode.totalTrafficPercentage
          : 100;

      return sum + (sourceTraffic * edge.trafficPercentage) / 100;
    },
    0
  );
}

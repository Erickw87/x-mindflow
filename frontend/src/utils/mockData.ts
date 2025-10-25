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
  // 定义 Combos（每个 Combo 包含一个 LB 和一个服务节点）
  const combos = [
    { id: 'combo-entry', type: 'combo' as const, label: '入口网关', branch: 'default' },
    { id: 'combo-a-branch1', type: 'combo' as const, label: '服务 A (v1.0)', branch: 'default' },
    { id: 'combo-a-branch2', type: 'combo' as const, label: '服务 A (v1.1)', branch: 'v1.1' },
    { id: 'combo-b-branch1', type: 'combo' as const, label: '服务 B (v2.0)', branch: 'default' },
    { id: 'combo-b-branch2', type: 'combo' as const, label: '服务 B (v2.1)', branch: 'v1.1' },
    { id: 'combo-c', type: 'combo' as const, label: '服务 C', branch: 'default' },
  ];

  const nodes: Array<ServiceNode | LBNode> = [
    // ========== Entry 组合 ==========
    // Entry 的 LB
    {
      id: 'lb-entry',
      type: 'lb',
      rules: [
        {
          name: '入口规则',
          targetVersions: ['v2.0'],
          trafficPercentage: 100,
        },
      ],
      branch: 'default',
      comboId: 'combo-entry', // 归属于 combo-entry
    } as any,
    // Entry 服务
    {
      id: 'entry',
      type: 'service',
      name: '入口网关',
      versions: [
        {
          version: 'v2.0',
          instanceCount: 10,
          color: '#3b82f6',
          status: 'healthy',
        },
      ],
      totalTrafficPercentage: 100,
      totalInstances: 10,
      branch: 'default',
      comboId: 'combo-entry', // 归属于 combo-entry
    } as any,
    // ========== Service A branch1 组合 ==========
    {
      id: 'lb-a-branch1',
      type: 'lb',
      rules: [{ name: '规则1', targetVersions: ['v1.0'], trafficPercentage: 100 }],
      branch: 'default',
      comboId: 'combo-a-branch1',
    } as any,
    {
      id: 'service-a-branch1',
      type: 'service',
      name: '服务 A',
      versions: [{ version: 'v1.0', instanceCount: 14, color: '#10b981', status: 'healthy' }],
      totalTrafficPercentage: 70,
      totalInstances: 14,
      branch: 'default',
      comboId: 'combo-a-branch1',
    } as any,

    // ========== Service A branch2 组合 ==========
    {
      id: 'lb-a-branch2',
      type: 'lb',
      rules: [{ name: '规则2', targetVersions: ['v1.1'], trafficPercentage: 100 }],
      branch: 'v1.1',
      comboId: 'combo-a-branch2',
    } as any,
    {
      id: 'service-a-branch2',
      type: 'service',
      name: '服务 A',
      versions: [{ version: 'v1.1', instanceCount: 6, color: '#f59e0b', isTarget: true, status: 'warning' }],
      totalTrafficPercentage: 30,
      totalInstances: 6,
      branch: 'v1.1',
      comboId: 'combo-a-branch2',
    } as any,

    // ========== Service B branch1 组合 ==========
    {
      id: 'lb-b-branch1',
      type: 'lb',
      rules: [{ name: '规则1', targetVersions: ['v2.0'], trafficPercentage: 100 }],
      branch: 'default',
      comboId: 'combo-b-branch1',
    } as any,
    {
      id: 'service-b-branch1',
      type: 'service',
      name: '服务 B',
      versions: [{ version: 'v2.0', instanceCount: 10, color: '#8b5cf6', status: 'healthy' }],
      totalTrafficPercentage: 70,
      totalInstances: 10,
      branch: 'default',
      comboId: 'combo-b-branch1',
    } as any,

    // ========== Service B branch2 组合 ==========
    {
      id: 'lb-b-branch2',
      type: 'lb',
      rules: [{ name: '规则2', targetVersions: ['v2.1'], trafficPercentage: 100 }],
      branch: 'v1.1',
      comboId: 'combo-b-branch2',
    } as any,
    {
      id: 'service-b-branch2',
      type: 'service',
      name: '服务 B',
      versions: [{ version: 'v2.1', instanceCount: 5, color: '#ef4444', isTarget: true, status: 'error' }],
      totalTrafficPercentage: 30,
      totalInstances: 5,
      branch: 'v1.1',
      comboId: 'combo-b-branch2',
    } as any,

    // ========== Service C 组合 ==========
    {
      id: 'lb-c',
      type: 'lb',
      rules: [{ name: '规则1', targetVersions: ['v1.0'], trafficPercentage: 100 }],
      branch: 'default',
      comboId: 'combo-c',
    } as any,
    {
      id: 'service-c',
      type: 'service',
      name: '服务 C',
      versions: [{ version: 'v1.0', instanceCount: 8, color: '#06b6d4', status: 'healthy' }],
      totalTrafficPercentage: 100,
      totalInstances: 8,
      branch: 'default',
      comboId: 'combo-c',
    } as any,
  ];

  const edges = [
    // === 隐藏边：用于保持 LB 与 Service 的位置关系 ===
    { id: 'hidden-1', source: 'lb-entry', target: 'entry', hidden: true },
    { id: 'hidden-2', source: 'lb-a-branch1', target: 'service-a-branch1', hidden: true },
    { id: 'hidden-3', source: 'lb-a-branch2', target: 'service-a-branch2', hidden: true },
    { id: 'hidden-4', source: 'lb-b-branch1', target: 'service-b-branch1', hidden: true },
    { id: 'hidden-5', source: 'lb-b-branch2', target: 'service-b-branch2', hidden: true },
    { id: 'hidden-6', source: 'lb-c', target: 'service-c', hidden: true },

    // === 实际的流量连接边：Service → LB ===
    // Entry 出来分流到 Service A 的两个分支
    {
      id: 'edge-1',
      source: 'entry',
      target: 'lb-a-branch1',
      trafficPercentage: 70,
      label: '流量：70%',
      branch: 'default',
    },
    {
      id: 'edge-2',
      source: 'entry',
      target: 'lb-a-branch2',
      trafficPercentage: 30,
      label: '流量：30%',
      branch: 'v1.1',
    },
    // Service A 到 Service B
    {
      id: 'edge-3',
      source: 'service-a-branch1',
      target: 'lb-b-branch1',
      trafficPercentage: 70,
      label: '流量：70%',
      branch: 'default',
    },
    {
      id: 'edge-4',
      source: 'service-a-branch2',
      target: 'lb-b-branch2',
      trafficPercentage: 30,
      label: '流量：30%',
      branch: 'v1.1',
    },
    // Service B 汇聚到 Service C
    {
      id: 'edge-5',
      source: 'service-b-branch1',
      target: 'lb-c',
      trafficPercentage: 70,
      label: '流量：70%',
      branch: 'default',
    },
    {
      id: 'edge-6',
      source: 'service-b-branch2',
      target: 'lb-c',
      trafficPercentage: 30,
      label: '流量：30%',
      branch: 'v1.1',
    },
  ];

  return { nodes, edges, combos };
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

/**
 * 拓扑可视化类型定义
 * 基于 Issue #11 的需求设计
 */

/**
 * 服务版本信息
 */
export interface ServiceVersion {
  version: string;
  instanceCount: number;
  color: string;
  isTarget?: boolean;
}

/**
 * 服务节点
 */
export interface ServiceNode {
  id: string;
  type: 'service';
  name: string;
  versions: ServiceVersion[];
  totalTrafficPercentage: number;
  totalInstances: number;
  branch?: string;
}

/**
 * LB 规则
 */
export interface LBRule {
  name: string;
  targetVersions: string[];
  trafficPercentage: number;
  ruleDetails?: string;
}

/**
 * LB 节点
 */
export interface LBNode {
  id: string;
  type: 'lb';
  rules: LBRule[];
  branch?: string;
}

/**
 * 连接边
 */
export interface TopologyEdge {
  id: string;
  source: string;
  target: string;
  trafficPercentage: number;
  label?: string;
  branch?: string;
}

/**
 * 拓扑图数据
 */
export interface TopologyData {
  nodes: Array<ServiceNode | LBNode>;
  edges: TopologyEdge[];
}

/**
 * 拓扑模式
 */
export type TopologyMode = 'view' | 'edit';

/**
 * LB 配置 (用于配置与图的双向转换)
 */
export interface LBConfig {
  serviceName: string;
  rules: Array<{
    name: string;
    conditions: Record<string, string>;
    targets: Array<{
      version: string;
      weight: number;
    }>;
  }>;
}

/**
 * 节点位置信息
 */
export interface NodePosition {
  x: number;
  y: number;
}

/**
 * G6 节点数据 (扩展标准 Node 类型)
 */
export interface G6NodeData {
  id: string;
  type: string;
  data: ServiceNode | LBNode;
  style?: Record<string, unknown>;
}

/**
 * G6 边数据
 */
export interface G6EdgeData {
  id: string;
  source: string;
  target: string;
  data: TopologyEdge;
  style?: Record<string, unknown>;
}

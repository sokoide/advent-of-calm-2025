export interface CalmArchitecture {
  "unique-id": string;
  name: string;
  description: string;
  nodes: CalmNode[];
  relationships: CalmRelationship[];
  flows?: CalmFlow[];
  controls?: Record<string, CalmControl>;
  metadata?: Record<string, any>;
}

export interface CalmControl {
  description: string;
  requirements?: CalmRequirement[];
}

export interface CalmRequirement {
  "requirement-url": string;
  config?: any;
  "config-url"?: string;
}

export type NodeOriginType = 'explicit' | 'loop' | 'function';

export interface NodeOrigin {
  type: NodeOriginType;
  file: string;
  line: number;
  funcName?: string;
  loopIndex?: number;
  loopMax?: number;
  loopVar?: string;
}

export interface CalmNode {
  "unique-id": string;
  "node-type": string;
  name: string;
  description: string;
  owner?: string;
  costCenter?: string;
  metadata?: Record<string, any>;
  interfaces?: CalmInterface[];
  _origin?: NodeOrigin;
}

export interface CalmInterface {
  "unique-id": string;
  protocol: string;
  port?: number;
}

export interface CalmRelationship {
  "unique-id": string;
  description: string;
  "relationship-type": {
    connects?: {
      source: { node: string };
      destination: { node: string };
    };
    interacts?: {
      actor: string;
      nodes: string[];
    };
    "composed-of"?: {
      container: string;
      nodes: string[];
    };
  };
}

export interface CalmFlow {
  "unique-id": string;
  name: string;
  description: string;
  metadata?: Record<string, any>;
  transitions: CalmTransition[];
}

export interface CalmTransition {
  "relationship-unique-id": string;
  "sequence-number": number;
  direction: string;
}

export interface LayoutData {
  nodes: Record<string, { x: number; y: number }>;
  parentMap?: Record<string, string>;
}

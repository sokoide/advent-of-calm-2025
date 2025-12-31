export interface CalmArchitecture {
  "unique-id": string;
  name: string;
  description: string;
  nodes: CalmNode[];
  relationships: CalmRelationship[];
  flows?: CalmFlow[];
  metadata?: Record<string, any>;
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
  transitions: CalmTransition[];
}

export interface CalmTransition {
  "relationship-unique-id": string;
  "sequence-number": number;
  direction: string;
}

export interface LayoutData {
  nodes: Record<string, { x: number; y: number }>;
}

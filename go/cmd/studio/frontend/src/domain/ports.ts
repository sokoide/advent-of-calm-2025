import type { LayoutData } from './calm';

export interface ContentSnapshot {
  goCode: string;
  d2Code: string;
  svg: string;
  json: string;
}

export interface SyncASTRequest {
  action: 'add' | 'update' | 'delete';
  nodeId: string;
  nodeType?: string;
  name?: string;
  desc?: string;
  property?: string;
  value?: string;
}

export interface PatchOrigin {
  file: string;
  line: number;
  loopVar?: string;
  loopMax?: number;
}

export interface PatchOperation {
  type:
  | 'update-node'
  | 'delete-node'
  | 'update-count'
  | 'delete-relationship'
  | 'add-relationship'
  | 'add-interface'
  | 'delete-interface'
  | 'add-flow'
  | 'update-flow'
  | 'delete-flow';
  nodeId?: string;
  origin?: PatchOrigin;
  property?: string;
  value?: any;
  sourceNode?: string; // For add-relationship
  targetNode?: string; // For add-relationship
  isInteracts?: boolean; // For add-relationship
  interfaceId?: string; // For add/delete-interface
  protocol?: string; // For add-interface
  flowId?: string; // For delete-flow, add-flow, update-flow
  flowName?: string; // For add-flow, update-flow
  flowDesc?: string; // For add-flow, update-flow
  flowSteps?: string[]; // For add-flow, update-flow (list of relationship IDs)
}

export interface StudioAPI {
  fetchContent(): Promise<ContentSnapshot>;
  fetchSVG(): Promise<string>;
  fetchLayout(archId: string): Promise<LayoutData>;
  saveLayout(archId: string, layout: LayoutData): Promise<void>;
  syncAST(request: SyncASTRequest): Promise<void>;
  patchAST(ops: PatchOperation[]): Promise<void>;
  updateGo(content: string): Promise<void>;
  previewJSONSync(json: string): Promise<{ newCode?: string; error?: string }>;
}

export interface RealtimeClient {
  connect(onMessage: (msg: string) => void): () => void;
}

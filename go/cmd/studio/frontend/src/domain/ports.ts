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

export interface StudioAPI {
  fetchContent(): Promise<ContentSnapshot>;
  fetchSVG(): Promise<string>;
  fetchLayout(archId: string): Promise<LayoutData>;
  saveLayout(archId: string, layout: LayoutData): Promise<void>;
  syncAST(request: SyncASTRequest): Promise<void>;
  updateGo(content: string): Promise<void>;
  previewJSONSync(json: string): Promise<{ newCode?: string; error?: string }>;
}

export interface RealtimeClient {
  connect(onMessage: (msg: string) => void): () => void;
}

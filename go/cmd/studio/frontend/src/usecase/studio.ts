import type { LayoutData } from '../domain/calm';
import type { StudioAPI, RealtimeClient, SyncASTRequest } from '../domain/ports';

export class StudioUseCase {
  private readonly api: StudioAPI;
  private readonly realtime: RealtimeClient;

  constructor(api: StudioAPI, realtime: RealtimeClient) {
    this.api = api;
    this.realtime = realtime;
  }

  connectRealtime(onRefresh: (msg: string) => void): () => void {
    return this.realtime.connect(onRefresh);
  }

  fetchContent() {
    return this.api.fetchContent();
  }

  fetchSVG() {
    return this.api.fetchSVG();
  }

  fetchLayout(archId: string) {
    return this.api.fetchLayout(archId);
  }

  saveLayout(archId: string, layout: LayoutData) {
    return this.api.saveLayout(archId, layout);
  }

  syncAST(request: SyncASTRequest) {
    return this.api.syncAST(request);
  }

  updateGo(content: string) {
    return this.api.updateGo(content);
  }

  previewJSONSync(json: string) {
    return this.api.previewJSONSync(json);
  }
}

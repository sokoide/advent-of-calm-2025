import axios from 'axios';
import type { StudioAPI, ContentSnapshot, SyncASTRequest } from '../domain/ports';
import type { LayoutData } from '../domain/calm';

export class StudioAPIClient implements StudioAPI {
  private readonly baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async fetchContent(): Promise<ContentSnapshot> {
    const resp = await axios.get(`${this.baseUrl}/content`);
    return resp.data as ContentSnapshot;
  }

  async fetchSVG(): Promise<string> {
    try {
      const resp = await axios.get(`${this.baseUrl}/svg`);
      if (resp.data && typeof resp.data.svg === 'string') {
        return resp.data.svg as string;
      }
      if (typeof resp.data === 'string') {
        return resp.data;
      }
    } catch (err) {
      const resp = await axios.get(`${this.baseUrl}/content`);
      return (resp.data?.svg as string) ?? '';
    }
    return '';
  }

  async fetchLayout(archId: string): Promise<LayoutData> {
    const resp = await axios.get(`${this.baseUrl}/layout?id=${archId}`);
    return resp.data as LayoutData;
  }

  async saveLayout(archId: string, layout: LayoutData): Promise<void> {
    await axios.post(`${this.baseUrl}/layout?id=${archId}`, layout);
  }

  async syncAST(request: SyncASTRequest): Promise<void> {
    await axios.post(`${this.baseUrl}/sync-ast`, request);
  }

  async updateGo(content: string): Promise<void> {
    await axios.post(`${this.baseUrl}/update`, { type: 'go', content });
  }

  async previewJSONSync(json: string): Promise<{ newCode?: string; error?: string }> {
    const resp = await axios.post(`${this.baseUrl}/preview-json-sync`, { json });
    return resp.data as { newCode?: string; error?: string };
  }
}

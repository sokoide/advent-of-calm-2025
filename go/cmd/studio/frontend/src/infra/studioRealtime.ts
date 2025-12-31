import type { RealtimeClient } from '../domain/ports';

export class StudioRealtime implements RealtimeClient {
  private readonly host: string;
  private readonly protocol: string;

  constructor(host: string, protocol: string) {
    this.host = host;
    this.protocol = protocol;
  }

  connect(onMessage: (msg: string) => void): () => void {
    const wsProtocol = this.protocol === 'https:' ? 'wss:' : 'ws:';
    const socket = new WebSocket(`${wsProtocol}//${this.host}/ws`);
    socket.onmessage = (event) => onMessage(event.data as string);
    return () => socket.close();
  }
}

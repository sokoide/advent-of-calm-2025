import { useCallback, useEffect, useState, useRef } from 'react';
import { 
  useNodesState, 
  useEdgesState,
  addEdge,
  type Node,
  type Connection,
} from 'reactflow';
import 'reactflow/dist/style.css';
import axios from 'axios';
import { 
  Monitor, 
  Code2, 
  FileJson, 
  Columns, 
  RefreshCw,
  CheckCircle2,
  Play,
  Save
} from 'lucide-react';
import * as Resizable from 'react-resizable-panels';

import { transformToReactFlow } from './utils/transformer';
import { getLayoutedElements } from './utils/layout';
import type { CalmArchitecture, LayoutData, CalmNode } from './types/calm';
import Sidebar from './components/Sidebar';
import DiagramView from './components/DiagramView';
import CodeEditor from './components/CodeEditor';

const BASE_URL = window.location.origin;

type TabType = 'diagram' | 'go' | 'd2' | 'merged';

function App() {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [loading, setLoading] = useState(true);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('diagram');

  // Clear selection when tab changes
  useEffect(() => {
    setSelectedNode(null);
  }, [activeTab]);

  const ws = useRef<WebSocket | null>(null);
  
  // Content states
  const [goCode, setGoCode] = useState('');
  const [d2Code, setD2Code] = useState('');
  const [archId, setArchId] = useState('');

  // Flag to avoid re-fetching while we are typing
  const isUpdating = useRef(false);

  const saveLayout = useCallback(async (currentNodes: Node[]) => {
    if (!archId) return;
    try {
      const newLayout: LayoutData = { nodes: {} };
      currentNodes.forEach(n => {
        newLayout.nodes[n.id] = n.position;
      });
      await axios.post(`${BASE_URL}/layout?id=${archId}`, newLayout);
    } catch (err) {
      console.error('Failed to save layout:', err);
    }
  }, [archId]);

  const onResetLayout = useCallback(() => {
    const layouted = getLayoutedElements(nodes, edges, 'LR');
    setNodes([...layouted.nodes]);
    setEdges([...layouted.edges]);
    saveLayout(layouted.nodes);
  }, [nodes, edges, setNodes, setEdges, saveLayout]);

  const fetchData = useCallback(async (isWSUpdate = false) => {
    if (isUpdating.current && !isWSUpdate) return;
    try {
      const contentResp = await axios.get(`${BASE_URL}/content`);
      const { goCode: remoteGo, d2Code: remoteD2, json } = contentResp.data;
      const arch: CalmArchitecture = JSON.parse(json);
      const id = arch['unique-id'];
      
      setGoCode(remoteGo);
      setD2Code(remoteD2);
      setArchId(id);

      const layoutResp = await axios.get(`${BASE_URL}/layout?id=${id}`);
      const layout: LayoutData = layoutResp.data;

      const { nodes: initialNodes, edges: initialEdges } = transformToReactFlow(arch, layout);
      
      const hasStoredLayout = layout.nodes && Object.keys(layout.nodes).length > 0;
      if (!hasStoredLayout) {
        const layouted = getLayoutedElements(initialNodes, initialEdges, 'LR');
        setNodes(layouted.nodes);
        setEdges(initialEdges); // initialEdges are not changed by layout currently
        setTimeout(() => saveLayout(layouted.nodes), 1000);
      } else {
        setNodes(initialNodes);
        setEdges(initialEdges);
      }
    } catch (err) {
      console.error('Failed to fetch data:', err);
    } finally {
      setLoading(false);
    }
  }, [setNodes, setEdges, saveLayout]);

  useEffect(() => {
    fetchData();

    // Setup WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socket = new WebSocket(`${protocol}//${window.location.host}/ws`);
    
    socket.onmessage = (event) => {
      if (event.data === 'refresh') {
        console.log('🔄 Remote change detected, refreshing...');
        fetchData(true);
      }
    };

    ws.current = socket;
    return () => socket.close();
  }, [fetchData]);

  const onConnect = useCallback(
    (params: Connection) => setEdges((prev) => addEdge(params, prev)),
    [setEdges]
  );

  const onNodeDragStop = useCallback(async (_: any, node: Node) => {
    saveLayout(nodes.map(n => n.id === node.id ? node : n));
  }, [nodes, saveLayout]);

  const onAddNode = useCallback(async () => {
    const id = `node-${Date.now()}`;
    const newNode: Node = {
      id,
      type: 'service',
      position: { x: 100, y: 100 },
      data: { 
        label: 'New Node',
        calm: {
          'unique-id': id,
          'node-type': 'service',
          name: 'New Node',
          description: '',
        }
      },
    };
    setNodes((nds) => nds.concat(newNode));
    saveLayout(nodes.concat(newNode));

    await axios.post(`${BASE_URL}/sync-ast`, {
      action: 'add',
      nodeId: id,
      nodeType: 'Service',
      name: 'New Node',
      desc: '',
    });
    fetchData(true);
  }, [nodes, setNodes, saveLayout, fetchData]);

  const onUpdateNode = useCallback(async (id: string, updatedCalm: Partial<CalmNode>) => {
    if (updatedCalm.name) {
      await axios.post(`${BASE_URL}/sync-ast`, { action: 'update', nodeId: id, property: 'name', value: updatedCalm.name });
    }
    if (updatedCalm.owner) {
      await axios.post(`${BASE_URL}/sync-ast`, { action: 'update', nodeId: id, property: 'owner', value: updatedCalm.owner });
    }
    fetchData(true);
  }, [fetchData]);

  const onDeleteNode = useCallback(async (id: string) => {
    await axios.post(`${BASE_URL}/sync-ast`, { action: 'delete', nodeId: id });
    setSelectedNode(null);
    fetchData(true);
  }, [fetchData]);

  const handleGoCodeChange = async (val: string | undefined) => {
    if (val === undefined) return;
    setGoCode(val);
    isUpdating.current = true;
    await axios.post(`${BASE_URL}/update`, { type: 'go', content: val });
    setTimeout(() => { isUpdating.current = false; }, 1000);
  };

  const handleApplyD2 = async () => {
    try {
      const resp = await axios.post(`${BASE_URL}/d2-to-go`, { d2Code });
      if (resp.data.success) {
        alert('Applied to Go DSL!');
        fetchData(true);
      }
    } catch (err) {
      alert('Failed to apply D2 changes');
    }
  };

  const handlePreviewD2 = () => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: 'd2', content: d2Code }));
    }
  };

  const renderDiagram = () => (
    <DiagramView 
      nodes={nodes} 
      edges={edges}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      onConnect={onConnect}
      onNodeDragStop={onNodeDragStop}
      onNodeClick={(_, node) => {
        console.log('Node clicked:', node.id);
        setSelectedNode(node);
      }}
      onPaneClick={() => setSelectedNode(null)}
      onAddNode={onAddNode}
      onResetLayout={onResetLayout}
    />
  );

  if (loading) return (
    <div className="flex flex-col h-screen w-screen items-center justify-center bg-slate-950 text-slate-200 italic tracking-widest uppercase text-sm">
      <RefreshCw className="animate-spin text-blue-500 mb-4" size={32} />
      Loading Studio...
    </div>
  );

  return (
    <div className="flex flex-col h-screen w-screen bg-slate-950 text-slate-200 overflow-hidden font-sans">
      <header className="flex items-center justify-between px-4 py-2 bg-slate-900 border-b border-slate-800 flex-shrink-0">
        <div className="flex items-center gap-6">
          <h1 className="text-lg font-bold text-blue-400 flex items-center gap-2">
            <Monitor size={20} /> CALM Studio
          </h1>
          <nav className="flex bg-slate-800/50 rounded-lg p-1 border border-slate-700">
            {[
              { id: 'diagram', label: 'Diagram', icon: Monitor },
              { id: 'go', label: 'Go DSL', icon: Code2 },
              { id: 'd2', label: 'D2 Source', icon: FileJson },
              { id: 'merged', label: 'Merged', icon: Columns },
            ].map((t) => (
              <button
                key={t.id}
                onClick={() => setActiveTab(t.id as TabType)}
                className={`flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-medium transition-all ${
                  activeTab === t.id ? 'bg-blue-600 text-white shadow-lg' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700'
                }`}
              >
                <t.icon size={14} /> {t.label}
              </button>
            ))}
          </nav>
        </div>
        <div className="flex items-center gap-3">
          <button onClick={() => fetchData(true)} className="p-2 hover:bg-slate-800 rounded-full transition-colors text-slate-400" title="Sync Refresh">
            <RefreshCw size={18} />
          </button>
          <div className="h-4 w-[1px] bg-slate-700 mx-1" />
          <button className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 text-white px-4 py-1.5 rounded-md text-sm font-medium transition-all shadow-md active:scale-95">
            <CheckCircle2 size={16} /> Validate
          </button>
        </div>
      </header>

      <main className="flex-1 overflow-hidden relative">
        {activeTab === 'diagram' && renderDiagram()}
        
        {activeTab === 'go' && (
          <CodeEditor value={goCode} language="go" onChange={handleGoCodeChange} />
        )}

        {activeTab === 'd2' && (
          <div className="flex flex-col h-full">
            <div className="bg-slate-900 px-4 py-2 flex gap-3 border-b border-slate-800 shadow-sm">
              <button onClick={handlePreviewD2} className="flex items-center gap-2 px-3 py-1.5 bg-slate-800 hover:bg-slate-700 rounded text-xs font-medium border border-slate-700 transition-colors">
                <Play size={12} /> Preview Diagram
              </button>
              <button onClick={handleApplyD2} className="flex items-center gap-2 px-3 py-1.5 bg-green-700 hover:bg-green-600 rounded text-xs font-medium text-white shadow-md transition-colors">
                <Save size={12} /> Apply to Go DSL
              </button>
            </div>
            <div className="flex-1">
              <CodeEditor value={d2Code} language="yaml" onChange={(val) => setD2Code(val || '')} />
            </div>
          </div>
        )}

        {activeTab === 'merged' && (
          <Resizable.Group orientation="horizontal" className="h-full">
            <Resizable.Panel defaultSize={40} minSize={25}>
              <div className="flex flex-col h-full border-r border-slate-800">
                <div className="bg-slate-900/50 px-4 py-2 border-b border-slate-800 text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                  Go DSL Editor
                </div>
                <CodeEditor value={goCode} language="go" onChange={handleGoCodeChange} />
              </div>
            </Resizable.Panel>
            <Resizable.Separator className="w-1.5 bg-slate-900 hover:bg-blue-600 transition-colors cursor-col-resize flex items-center justify-center">
              <div className="w-0.5 h-8 bg-slate-700 rounded-full" />
            </Resizable.Separator>
            <Resizable.Panel defaultSize={60} minSize={30}>
              {renderDiagram()}
            </Resizable.Panel>
          </Resizable.Group>
        )}

        <Sidebar 
          selectedNode={selectedNode}
          onUpdate={onUpdateNode}
          onDelete={onDeleteNode}
          onClose={() => setSelectedNode(null)}
        />
      </main>
    </div>
  );
}

export default App;

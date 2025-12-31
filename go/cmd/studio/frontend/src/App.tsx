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
  Save,
  Layers,
  FileCode
} from 'lucide-react';
import * as Resizable from 'react-resizable-panels';

import { transformToReactFlow } from './utils/transformer';
import { getLayoutedElements } from './utils/layout';
import type { CalmArchitecture, LayoutData, CalmNode } from './types/calm';
import Sidebar from './components/Sidebar';
import DiagramView from './components/DiagramView';
import CodeEditor from './components/CodeEditor';

const BASE_URL = window.location.origin;

type TabType = 'merged' | 'diagram' | 'go' | 'json' | 'd2-diagram' | 'd2-dsl';

function App() {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [loading, setLoading] = useState(true);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('merged');

  // Clear selection when tab changes
  useEffect(() => {
    setSelectedNode(null);
  }, [activeTab]);

  const ws = useRef<WebSocket | null>(null);
  
  // Content states
  const [goCode, setGoCode] = useState('');
  const [d2Code, setD2Code] = useState('');
  const [jsonCode, setJsonCode] = useState('');
  const [svgCode, setSvgCode] = useState('');
  const [archId, setArchId] = useState('');
  const [showDiff, setShowDiff] = useState(false);
  const [previewCode, setPreviewCode] = useState('');

  // Flag to avoid re-fetching while we are typing
  const isUpdating = useRef(false);

  const saveLayout = useCallback(async (currentNodes: Node[]) => {
    if (!archId) return;
    try {
      const nodeMap = new Map(currentNodes.map(n => [n.id, n]));
      const absCache = new Map<string, { x: number, y: number }>();

      const resolveAbsolute = (id: string): { x: number, y: number } => {
        const cached = absCache.get(id);
        if (cached) return cached;
        const node = nodeMap.get(id);
        if (!node) return { x: 0, y: 0 };

        if (!node.parentNode) {
          absCache.set(id, node.position);
          return node.position;
        }

        const parentPos = resolveAbsolute(node.parentNode);
        const absPos = {
          x: parentPos.x + node.position.x,
          y: parentPos.y + node.position.y
        };
        absCache.set(id, absPos);
        return absPos;
      };

      const newLayout: LayoutData = { nodes: {} };
      currentNodes.forEach(n => {
        newLayout.nodes[n.id] = resolveAbsolute(n.id);
      });
      await axios.post(`${BASE_URL}/layout?id=${archId}`, newLayout);
    } catch (err) {
      console.error('Failed to save layout:', err);
    }
  }, [archId]);

  const onResetLayout = useCallback(() => {
    const layouted = getLayoutedElements(nodes, edges, 'LR');
    // Ensure we trigger a state update with fresh objects
    const refreshedNodes = layouted.nodes.map(n => ({
      ...n,
      position: { ...n.position }
    }));
    setNodes(refreshedNodes);
    setEdges([...layouted.edges]);
    saveLayout(refreshedNodes);
  }, [nodes, edges, setNodes, setEdges, saveLayout]);

  const fetchData = useCallback(async (isWSUpdate = false) => {
    if (isUpdating.current && !isWSUpdate) return;
    try {
      const contentResp = await axios.get(`${BASE_URL}/content`);
      const { goCode: remoteGo, d2Code: remoteD2, json, svg } = contentResp.data;
      
      setGoCode(remoteGo);
      setD2Code(remoteD2);
      setSvgCode(svg);

      if (!json || json === "" || json === "null" || json === "{}") {
        console.log("⚠️ JSON is empty, showing code only");
        setLoading(false);
        return;
      }

      const arch: CalmArchitecture = JSON.parse(json);
      const archUniqueID = arch['unique-id'];
      setJsonCode(JSON.stringify(arch, null, 2));
      setArchId(archUniqueID);

      const layoutResp = await axios.get(`${BASE_URL}/layout?id=${archUniqueID}`);
      const layout: LayoutData = layoutResp.data;

      const { nodes: initialNodes, edges: initialEdges } = transformToReactFlow(arch, layout);
      
      const hasStoredLayout = layout.nodes && Object.keys(layout.nodes).length > 0;
      if (!hasStoredLayout) {
        const layouted = getLayoutedElements(initialNodes, initialEdges, 'LR');
        setNodes(layouted.nodes);
        setEdges(initialEdges);
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

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socket = new WebSocket(`${protocol}//${window.location.host}/ws`);
    
    socket.onmessage = (event) => {
      if (event.data === 'refresh' || event.data === 'refresh-svg') {
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

  const handleApplyJSON = async () => {
    try {
      const resp = await axios.post(`${BASE_URL}/preview-json-sync`, { json: jsonCode });
      if (resp.data.error) {
        alert('Validation Error: ' + resp.data.error);
        return;
      }
      setPreviewCode(resp.data.newCode);
      setShowDiff(true);
    } catch (err) {
      alert('Failed to generate preview');
    }
  };

  const confirmApply = async () => {
    try {
      isUpdating.current = true;
      await axios.post(`${BASE_URL}/update`, { type: 'go', content: previewCode });
      setShowDiff(false);
      setTimeout(async () => {
        await fetchData(true);
        isUpdating.current = false;
        alert('Applied to Go DSL!');
      }, 500);
    } catch (err) {
      isUpdating.current = false;
      alert('Failed to apply changes');
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
              { id: 'merged', label: 'Merged', icon: Columns },
              { id: 'diagram', label: 'Diagram', icon: Monitor },
              { id: 'go', label: 'Go DSL', icon: Code2 },
              { id: 'json', label: 'CALM JSON', icon: FileJson },
              { id: 'd2-diagram', label: 'D2 Diagram', icon: Layers },
              { id: 'd2-dsl', label: 'D2 DSL', icon: FileCode },
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

        {activeTab === 'diagram' && renderDiagram()}
        
        {activeTab === 'go' && (
          <CodeEditor value={goCode} language="go" onChange={handleGoCodeChange} />
        )}

        {activeTab === 'json' && (
          <div className="flex flex-col h-full">
            <div className="bg-slate-900 px-4 py-2 flex gap-3 border-b border-slate-800 shadow-sm">
              <button onClick={() => alert('Validation not implemented yet')} className="flex items-center gap-2 px-3 py-1.5 bg-slate-800 hover:bg-slate-700 rounded text-xs font-medium border border-slate-700 transition-colors">
                <CheckCircle2 size={12} /> Validate JSON
              </button>
              <button onClick={handleApplyJSON} className="flex items-center gap-2 px-3 py-1.5 bg-green-700 hover:bg-green-600 rounded text-xs font-medium text-white shadow-md transition-all active:scale-95">
                <Save size={12} /> Apply to Go DSL
              </button>
            </div>
            <div className="flex-1">
              <CodeEditor value={jsonCode} language="json" onChange={(val) => setJsonCode(val || '')} />
            </div>
          </div>
        )}

        {activeTab === 'd2-diagram' && (
          <div className="flex flex-col h-full bg-slate-900 overflow-hidden">
            <div className="flex-1 overflow-auto bg-slate-800 p-8 flex items-start justify-center">
              {svgCode ? (
                <div 
                  className="bg-white rounded-xl shadow-2xl p-8"
                  style={{ minWidth: '800px' }}
                  dangerouslySetInnerHTML={{ __html: svgCode.replace('<svg ', '<svg style="width:100%;height:auto;" ') }} 
                />
              ) : (
                <div className="flex flex-col h-full items-center justify-center text-slate-500">
                  <RefreshCw className="animate-spin mb-4" size={32} />
                  <p className="text-lg">Generating High-Fidelity D2 Diagram...</p>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'd2-dsl' && (
          <div className="flex flex-col h-full">
            <div className="bg-slate-900 px-4 py-2 flex gap-3 border-b border-slate-800 shadow-sm text-xs text-slate-500 font-medium">
              Read-only D2 Source
            </div>
            <div className="flex-1">
              <CodeEditor value={d2Code} language="yaml" onChange={() => {}} />
            </div>
          </div>
        )}

        <Sidebar 
          selectedNode={selectedNode}
          onUpdate={onUpdateNode}
          onDelete={onDeleteNode}
          onClose={() => setSelectedNode(null)}
        />

        {showDiff && (
          <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm p-10">
            <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-5xl h-full max-h-[90vh] flex flex-col overflow-hidden">
              <div className="px-6 py-4 border-b border-slate-800 flex justify-between items-center">
                <h3 className="text-lg font-bold text-blue-400 flex items-center gap-2">
                  <RefreshCw size={20} /> Preview Changes (Go DSL)
                </h3>
                <button 
                  onClick={() => setShowDiff(false)}
                  className="text-slate-400 hover:text-white"
                >
                  ✕
                </button>
              </div>
              <div className="flex-1 overflow-hidden p-4">
                <CodeEditor value={previewCode} language="go" onChange={() => {}} />
              </div>
              <div className="px-6 py-4 bg-slate-950 border-t border-slate-800 flex justify-end gap-4">
                <button 
                  onClick={() => setShowDiff(false)}
                  className="px-6 py-2 rounded-md border border-slate-700 text-slate-400 hover:text-white hover:bg-slate-800 transition-all"
                >
                  Cancel
                </button>
                <button 
                  onClick={confirmApply}
                  className="px-8 py-2 bg-green-600 hover:bg-green-500 text-white rounded-md font-bold shadow-lg transition-all active:scale-95"
                >
                  Apply Changes
                </button>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
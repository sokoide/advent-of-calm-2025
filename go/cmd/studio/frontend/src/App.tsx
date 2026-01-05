import { useCallback, useEffect, useState } from 'react';
import { addEdge, type Node, type Connection } from 'reactflow';
import 'reactflow/dist/style.css';
import { Monitor, RefreshCw, CheckCircle2, Save } from 'lucide-react';
import * as Resizable from 'react-resizable-panels';

import { getLayoutedElements } from './utils/layout';
import type { StudioUseCase } from './usecase/studio';
import { createStudioUseCase } from './infra/studioApi';
import Sidebar from './components/Sidebar';
import DiagramView from './components/DiagramView';
import CodeEditor from './components/CodeEditor';
import TabNav, { type TabType } from './components/TabNav';
import { useStudioContent } from './hooks/useStudioContent';
import { useNodeOperations } from './hooks/useNodeOperations';
import { useD2Viewer } from './hooks/useD2Viewer';

function App() {
  const [studio, setStudio] = useState<StudioUseCase | null>(null);
  const [useLocalAgent, setUseLocalAgent] = useState(false);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('merged');
  const [showDiff, setShowDiff] = useState(false);
  const [previewCode, setPreviewCode] = useState('');

  const { zoom: d2Zoom, pan: d2Pan, isPanning, zoomIn, zoomOut, reset, startPanning, stopPanning, movePan } = useD2Viewer();

  const localAgentUrl = import.meta.env.VITE_LOCAL_AGENT_URL ?? 'http://localhost:8787';

  useEffect(() => {
    let cancelled = false;
    const initStudio = async () => {
      const { studio: newStudio, useLocalAgent: usesLocal } = await createStudioUseCase({
        localAgentUrl,
        fallbackBaseUrl: window.location.origin,
        host: window.location.host,
        protocol: window.location.protocol,
      });
      if (cancelled) return;
      setStudio(newStudio);
      setUseLocalAgent(usesLocal);
    };
    initStudio();
    return () => {
      cancelled = true;
    };
  }, [localAgentUrl]);

  const {
    nodes,
    edges,
    setNodes,
    setEdges,
    onNodesChange,
    onEdgesChange,
    goCode,
    d2Code,
    jsonCode,
    svgCode,
    loading,
    fetchData,
    fetchSVG,
    updateGoCode,
    setJsonCode,
    saveLayout,
  } = useStudioContent(studio);

  const { onAddNode, onUpdateNode, onDeleteNode, onNodeDragStop } = useNodeOperations({
    studio,
    nodes,
    edges,
    setNodes,
    setEdges,
    saveLayout,
    fetchData,
    setSelectedNode,
  });

  // Clear selection when tab changes
  useEffect(() => {
    setSelectedNode(null);
  }, [activeTab]);

  useEffect(() => {
    if (activeTab === 'd2-diagram') {
      fetchSVG();
    }
  }, [activeTab, fetchSVG]);

  const onResetLayout = useCallback(() => {
    const layouted = getLayoutedElements(nodes, edges, 'TB');
    const refreshedNodes = layouted.nodes.map((n) => ({
      ...n,
      position: { ...n.position },
    }));
    setNodes(refreshedNodes);
    setEdges([...layouted.edges]);
    saveLayout(refreshedNodes);
  }, [nodes, edges, setNodes, setEdges, saveLayout]);

  const onConnect = useCallback(
    (params: Connection) => setEdges((prev) => addEdge(params, prev)),
    [setEdges]
  );

  const handleGoCodeChange = async (val: string | undefined) => {
    if (val === undefined) return;
    await updateGoCode(val);
    if (useLocalAgent) {
      fetchData(true);
    }
  };

  const handleApplyJSON = async () => {
    try {
      const resp = await studio?.previewJSONSync(jsonCode);
      if (!resp) return;
      if (resp.error) {
        alert('Validation Error: ' + resp.error);
        return;
      }
      setPreviewCode(resp.newCode ?? '');
      setShowDiff(true);
    } catch {
      alert('Failed to generate preview');
    }
  };

  const confirmApply = async () => {
    try {
      await updateGoCode(previewCode);
      setShowDiff(false);
      setTimeout(async () => {
        await fetchData(true);
        alert('Applied to Go DSL!');
      }, 500);
    } catch {
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

  if (loading || !studio) {
    return (
      <div className="flex flex-col h-screen w-screen items-center justify-center bg-slate-950 text-slate-200 italic tracking-widest uppercase text-sm">
        <RefreshCw className="animate-spin text-blue-500 mb-4" size={32} />
        Loading Studio...
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen w-screen bg-slate-950 text-slate-200 overflow-hidden font-sans">
      <header className="flex items-center justify-between px-4 py-2 bg-slate-900 border-b border-slate-800 flex-shrink-0">
        <div className="flex items-center gap-6">
          <h1 className="text-lg font-bold text-blue-400 flex items-center gap-2">
            <Monitor size={20} /> CALM Studio
          </h1>
          <TabNav activeTab={activeTab} onTabChange={setActiveTab} />
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => fetchData(true)}
            className="p-2 hover:bg-slate-800 rounded-full transition-colors text-slate-400"
            title="Sync Refresh"
          >
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
              <button
                onClick={() => alert('Validation not implemented yet')}
                className="flex items-center gap-2 px-3 py-1.5 bg-slate-800 hover:bg-slate-700 rounded text-xs font-medium border border-slate-700 transition-colors"
              >
                <CheckCircle2 size={12} /> Validate JSON
              </button>
              <button
                onClick={handleApplyJSON}
                className="flex items-center gap-2 px-3 py-1.5 bg-green-700 hover:bg-green-600 rounded text-xs font-medium text-white shadow-md transition-all active:scale-95"
              >
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
            <div className="flex items-center gap-2 px-4 py-2 border-b border-slate-800 bg-slate-900/70">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wide">D2 Diagram</span>
              <div className="ml-auto flex items-center gap-2">
                <button
                  onClick={zoomOut}
                  className="px-2.5 py-1 text-xs rounded border border-slate-700 text-slate-300 hover:bg-slate-800"
                >
                  −
                </button>
                <span className="text-xs text-slate-400 w-12 text-center">{Math.round(d2Zoom * 100)}%</span>
                <button
                  onClick={zoomIn}
                  className="px-2.5 py-1 text-xs rounded border border-slate-700 text-slate-300 hover:bg-slate-800"
                >
                  +
                </button>
                <button
                  onClick={reset}
                  className="px-2.5 py-1 text-xs rounded border border-slate-700 text-slate-300 hover:bg-slate-800"
                >
                  Reset
                </button>
              </div>
            </div>
            <div className="flex-1 overflow-hidden bg-slate-800 p-6" onMouseLeave={stopPanning}>
              {svgCode ? (
                <div
                  className={`bg-white rounded-xl shadow-2xl p-6 w-full h-full select-none ${isPanning ? 'cursor-grabbing' : 'cursor-grab'
                    }`}
                  onMouseDown={(event) => {
                    if (event.button !== 0) return;
                    startPanning();
                  }}
                  onMouseUp={stopPanning}
                  onMouseMove={(event) => movePan(event.movementX, event.movementY)}
                >
                  <div
                    style={{
                      transform: `translate(${d2Pan.x}px, ${d2Pan.y}px) scale(${d2Zoom})`,
                      transformOrigin: '0 0',
                      width: '100%',
                    }}
                    dangerouslySetInnerHTML={{
                      __html: svgCode.replace('<svg ', '<svg style="width:100%;height:auto;" '),
                    }}
                  />
                </div>
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
              <CodeEditor value={d2Code} language="yaml" onChange={() => { }} readOnly />
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
                <button onClick={() => setShowDiff(false)} className="text-slate-400 hover:text-white">
                  ✕
                </button>
              </div>
              <div className="flex-1 overflow-hidden p-4">
                <CodeEditor value={previewCode} language="go" onChange={() => { }} readOnly />
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

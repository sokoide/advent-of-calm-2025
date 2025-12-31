import { useCallback, useEffect, useState } from 'react';
import ReactFlow, { 
  Background, 
  Controls, 
  useNodesState, 
  useEdgesState,
  addEdge,
  Panel,
  type Node,
  type Connection,
} from 'reactflow';
import 'reactflow/dist/style.css';
import axios from 'axios';
import { Plus, Layout } from 'lucide-react';

import { transformToReactFlow } from './utils/transformer';
import { getLayoutedElements } from './utils/layout';
import type { CalmArchitecture, LayoutData, CalmNode } from './types/calm';
import CalmNodeComponent from './components/CalmNode';
import Sidebar from './components/Sidebar';

const BASE_URL = window.location.origin;

const nodeTypes = {
  service: CalmNodeComponent,
  database: CalmNodeComponent,
  actor: CalmNodeComponent,
  system: CalmNodeComponent,
  queue: CalmNodeComponent,
};

function App() {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [loading, setLoading] = useState(true);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);

  // ... (fetchData and saveLayout remain same)

  const saveLayout = useCallback(async (currentNodes: Node[]) => {
    try {
      const contentResp = await axios.get(`${BASE_URL}/content`);
      const arch: CalmArchitecture = JSON.parse(contentResp.data.json);
      const archId = arch['unique-id'];

      const newLayout: LayoutData = { nodes: {} };
      currentNodes.forEach(n => {
        newLayout.nodes[n.id] = n.position;
      });

      await axios.post(`${BASE_URL}/layout?id=${archId}`, newLayout);
      console.log('Layout saved');
    } catch (err) {
      console.error('Failed to save layout:', err);
    }
  }, []);

  const onResetLayout = useCallback(() => {
    const layouted = getLayoutedElements(nodes, edges, 'LR');
    setNodes([...layouted.nodes]);
    setEdges([...layouted.edges]);
    saveLayout(layouted.nodes);
  }, [nodes, edges, setNodes, setEdges, saveLayout]);

  const fetchData = useCallback(async () => {
    try {
      // 1. Get CALM JSON
      const contentResp = await axios.get(`${BASE_URL}/content`);
      const arch: CalmArchitecture = JSON.parse(contentResp.data.json);
      
      // 2. Get Layout JSON
      const layoutResp = await axios.get(`${BASE_URL}/layout?id=${arch['unique-id']}`);
      const layout: LayoutData = layoutResp.data;

      // 3. Transform
      const { nodes: initialNodes, edges: initialEdges } = transformToReactFlow(arch, layout);
      
      // 4. Smart Layout
      const hasStoredLayout = layout.nodes && Object.keys(layout.nodes).length > 0;
      
      if (!hasStoredLayout) {
        console.log('No stored layout found, running auto-layout...');
        const layouted = getLayoutedElements(initialNodes, initialEdges, 'LR');
        setNodes(layouted.nodes);
        setEdges(layouted.edges);
        // Automatically save initial clean layout
        setTimeout(() => saveLayout(layouted.nodes), 1000);
      } else {
        console.log('Using stored layout data');
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
  }, [fetchData]);

  const onConnect = useCallback(
    (params: Connection) => setEdges((prev) => addEdge(params, prev)),
    [setEdges]
  );

  const onNodeDragStop = useCallback(async (_: any, node: Node) => {
    saveLayout(nodes.map(n => n.id === node.id ? node : n));
  }, [nodes, saveLayout]);

  const onNodeClick = useCallback((_: any, node: Node) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

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

    // Sync to AST
    await axios.post(`${BASE_URL}/sync-ast`, {
      action: 'add',
      nodeId: id,
      nodeType: 'Service',
      name: 'New Node',
      desc: '',
    });
  }, [nodes, setNodes, saveLayout]);

  const onUpdateNode = useCallback(async (id: string, updatedCalm: Partial<CalmNode>) => {
    setNodes((nds) =>
      nds.map((node) => {
        if (node.id === id) {
          const newCalm = { ...node.data.calm, ...updatedCalm };
          return {
            ...node,
            type: newCalm['node-type'],
            data: {
              ...node.data,
              label: newCalm.name,
              calm: newCalm,
            },
          };
        }
        return node;
      })
    );

    // Sync to AST (Property by property for now)
    if (updatedCalm.name) {
      await axios.post(`${BASE_URL}/sync-ast`, {
        action: 'update',
        nodeId: id,
        property: 'name',
        value: updatedCalm.name,
      });
    }
    if (updatedCalm.owner) {
      await axios.post(`${BASE_URL}/sync-ast`, {
        action: 'update',
        nodeId: id,
        property: 'owner',
        value: updatedCalm.owner,
      });
    }
  }, [setNodes]);

  const onDeleteNode = useCallback(async (id: string) => {
    setNodes((nds) => nds.filter((n) => n.id !== id));
    setEdges((eds) => eds.filter((e) => e.source !== id && e.target !== id));
    setSelectedNode(null);

    // Sync to AST
    await axios.post(`${BASE_URL}/sync-ast`, {
      action: 'delete',
      nodeId: id,
    });
  }, [setNodes, setEdges]);

  if (loading) return <div className="p-4">Loading architecture...</div>;

  return (
    <div style={{ width: '100vw', height: '100vh', backgroundColor: '#f8fafc' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeDragStop={onNodeDragStop}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        fitView
      >
        <Background />
        <Controls />
        <Panel position="top-left" className="flex gap-2">
          <button 
            onClick={onAddNode}
            className="flex items-center gap-2 bg-white px-3 py-2 rounded shadow border border-gray-200 hover:bg-gray-50 font-medium text-sm"
          >
            <Plus size={16} /> Add Node
          </button>
          <button 
            onClick={onResetLayout}
            className="flex items-center gap-2 bg-white px-3 py-2 rounded shadow border border-gray-200 hover:bg-gray-50 font-medium text-sm"
          >
            <Layout size={16} /> Auto Layout
          </button>
        </Panel>
      </ReactFlow>

      <Sidebar 
        selectedNode={selectedNode}
        onUpdate={onUpdateNode}
        onDelete={onDeleteNode}
        onClose={() => setSelectedNode(null)}
      />
    </div>
  );
}

export default App;
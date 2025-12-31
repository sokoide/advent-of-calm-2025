import ReactFlow, { 
  Background, 
  Controls, 
  Panel,
  type Node,
  type Edge,
  type Connection,
  type OnNodesChange,
  type OnEdgesChange,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Plus, Layout } from 'lucide-react';

import CalmNodeComponent from './CalmNode';

const nodeTypes = {
  service: CalmNodeComponent,
  database: CalmNodeComponent,
  actor: CalmNodeComponent,
  system: CalmNodeComponent,
  queue: CalmNodeComponent,
};

interface DiagramViewProps {
  nodes: Node[];
  edges: Edge[];
  onNodesChange: OnNodesChange;
  onEdgesChange: OnEdgesChange;
  onConnect: (params: Connection) => void;
  onNodeDragStop: (event: any, node: Node) => void;
  onNodeClick: (event: any, node: Node) => void;
  onPaneClick: () => void;
  onAddNode: () => void;
  onResetLayout: () => void;
}

const DiagramView = ({
  nodes,
  edges,
  onNodesChange,
  onEdgesChange,
  onConnect,
  onNodeDragStop,
  onNodeClick,
  onPaneClick,
  onAddNode,
  onResetLayout,
}: DiagramViewProps) => {
  return (
    <div className="w-full h-full bg-slate-950">
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
            className="flex items-center gap-2 bg-white px-3 py-2 rounded shadow border border-gray-200 hover:bg-gray-50 font-medium text-sm text-gray-700"
          >
            <Plus size={16} /> Add Node
          </button>
          <button 
            onClick={onResetLayout}
            className="flex items-center gap-2 bg-white px-3 py-2 rounded shadow border border-gray-200 hover:bg-gray-50 font-medium text-sm text-gray-700"
          >
            <Layout size={16} /> Auto Layout
          </button>
        </Panel>
      </ReactFlow>
    </div>
  );
};

export default DiagramView;

import { memo } from 'react';
import { Handle, Position, type NodeProps } from 'reactflow';
import { NodeResizer } from '@reactflow/node-resizer';
import '@reactflow/node-resizer/dist/style.css';
import { Database, Server, Users, Box, MessageSquare, Shield } from 'lucide-react';

const icons: Record<string, any> = {
  service: Server,
  database: Database,
  actor: Users,
  system: Box,
  queue: MessageSquare,
};

const CalmNode = ({ data, selected }: NodeProps) => {
  const type = data.calm['node-type'];
  const Icon = icons[type] || Box;
  const isContainer = data.isContainer === true;
  const hasControls = data.hasControls === true;
  const isComposedOfChild = data.isComposedOfChild === true;

  if (isContainer) {
    return (
      <>
        <NodeResizer
          color="#3b82f6"
          isVisible={selected}
          minWidth={200}
          minHeight={150}
          handleClassName="nodrag"
          lineClassName="nodrag"
          handleStyle={{
            width: 12,
            height: 12,
            backgroundColor: '#3b82f6',
            border: '2px solid white',
            borderRadius: 2,
            zIndex: 100,
          }}
          onResizeStart={() => {
            document.body.style.cursor = 'nwse-resize';
          }}
          onResizeEnd={() => {
            document.body.style.cursor = 'auto';
          }}
        />
        <div
          className={`w-full h-full rounded-xl border shadow-inner ${selected ? 'border-blue-500/70 ring-2 ring-blue-500/20' : 'border-slate-700/60'
            } bg-slate-900/40`}
        >
          <div className="flex items-center gap-3 px-4 py-3">
            <div className="rounded-full w-8 h-8 flex items-center justify-center bg-slate-800">
              <Icon size={16} className="text-blue-400" />
            </div>
            <div>
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-tight">{type}</div>
              <div className="text-sm font-semibold text-slate-200">{data.label}</div>
            </div>
          </div>
        </div>
      </>
    );
  }

  return (
    <div className={`px-4 py-2 shadow-xl rounded-lg bg-slate-900 border-2 transition-all relative
      ${selected ? 'border-blue-500 ring-2 ring-blue-500/20 scale-105 z-50' : 'border-slate-800'}
      ${isComposedOfChild ? 'ring-2 ring-purple-500/40' : ''}`}>
      <div className="flex items-center">
        <div className="rounded-full w-8 h-8 flex items-center justify-center bg-slate-800 mr-3">
          <Icon size={16} className="text-blue-400" />
        </div>
        <div>
          <div className="text-[10px] font-bold text-slate-500 uppercase tracking-tight">{type}</div>
          <div className="text-sm font-semibold text-slate-200">{data.label}</div>
        </div>
      </div>

      {hasControls && (
        <div className="absolute -top-2 -right-2 bg-amber-600 rounded-full p-1" title="Has Controls (NFRs)">
          <Shield size={10} className="text-white" />
        </div>
      )}

      <Handle type="target" position={Position.Top} className="w-2 h-2 !bg-slate-600 !border-slate-400" />
      <Handle type="source" position={Position.Bottom} className="w-2 h-2 !bg-slate-600 !border-slate-400" />
    </div>
  );
};

export default memo(CalmNode);

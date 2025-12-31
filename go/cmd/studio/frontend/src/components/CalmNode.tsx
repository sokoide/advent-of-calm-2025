import { memo } from 'react';
import { Handle, Position, type NodeProps } from 'reactflow';
import { Database, Server, Users, Box, MessageSquare } from 'lucide-react';

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

  return (
    <div className={`px-4 py-2 shadow-xl rounded-lg bg-slate-900 border-2 transition-all ${selected ? 'border-blue-500 ring-2 ring-blue-500/20 scale-105 z-50' : 'border-slate-800'}`}>
      <div className="flex items-center">
        <div className="rounded-full w-8 h-8 flex items-center justify-center bg-slate-800 mr-3">
          <Icon size={16} className="text-blue-400" />
        </div>
        <div>
          <div className="text-[10px] font-bold text-slate-500 uppercase tracking-tight">{type}</div>
          <div className="text-sm font-semibold text-slate-200">{data.label}</div>
        </div>
      </div>

      <Handle type="target" position={Position.Top} className="w-2 h-2 !bg-slate-600 !border-slate-400" />
      <Handle type="source" position={Position.Bottom} className="w-2 h-2 !bg-slate-600 !border-slate-400" />
    </div>
  );
};

export default memo(CalmNode);

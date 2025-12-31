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
    <div className={`px-4 py-2 shadow-md rounded-md bg-white border-2 transition-all ${selected ? 'border-blue-500 ring-2 ring-blue-200' : 'border-gray-200'}`}>
      <div className="flex items-center">
        <div className="rounded-full w-8 h-8 flex items-center justify-center bg-gray-100 mr-2">
          <Icon size={16} className="text-gray-600" />
        </div>
        <div>
          <div className="text-xs font-bold text-gray-500 uppercase">{type}</div>
          <div className="text-sm font-medium">{data.label}</div>
        </div>
      </div>

      <Handle type="target" position={Position.Top} className="w-2 h-2 !bg-gray-400" />
      <Handle type="source" position={Position.Bottom} className="w-2 h-2 !bg-gray-400" />
    </div>
  );
};

export default memo(CalmNode);

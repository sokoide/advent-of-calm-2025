import { useState, useEffect } from 'react';
import { type Node } from 'reactflow';
import { X, Trash2, Save } from 'lucide-react';
import type { CalmNode } from '../domain/calm';

interface SidebarProps {
  selectedNode: Node | null;
  onUpdate: (id: string, data: Partial<CalmNode>) => void;
  onDelete: (id: string) => void;
  onClose: () => void;
}

const Sidebar = ({ selectedNode, onUpdate, onDelete, onClose }: SidebarProps) => {
  const [formData, setFormData] = useState<Partial<CalmNode>>({});

  useEffect(() => {
    if (selectedNode) {
      setFormData(selectedNode.data.calm);
    }
  }, [selectedNode]);

  if (!selectedNode) return null;

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSave = () => {
    onUpdate(selectedNode.id, formData);
  };

  return (
    <aside className="absolute right-0 top-0 h-full w-80 bg-slate-900 shadow-2xl border-l border-slate-800 z-[100] flex flex-col animate-in slide-in-from-right duration-200">
      <div className="p-4 border-b border-slate-800 bg-slate-900/50 flex justify-between items-center">
        <h2 className="font-bold text-slate-200 flex items-center gap-2">
          Node Properties
        </h2>
        <button onClick={onClose} className="text-slate-500 hover:text-slate-300 transition-colors">
          <X size={20} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-5">
        <div className="space-y-1">
          <label className="block text-[10px] font-bold text-slate-500 uppercase tracking-wider">Unique ID</label>
          <input 
            type="text" 
            className="w-full p-2 bg-slate-950 border border-slate-800 rounded text-sm text-slate-400 cursor-not-allowed" 
            value={selectedNode.id} 
            readOnly 
          />
        </div>

        <div className="space-y-1">
          <label className="block text-[10px] font-bold text-slate-500 uppercase tracking-wider">Name</label>
          <input 
            name="name"
            type="text" 
            className="w-full p-2 bg-slate-800 border border-slate-700 rounded text-sm text-slate-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all" 
            value={formData.name || ''} 
            onChange={handleChange}
          />
        </div>

        <div className="space-y-1">
          <label className="block text-[10px] font-bold text-slate-500 uppercase tracking-wider">Type</label>
          <select 
            name="node-type"
            className="w-full p-2 bg-slate-800 border border-slate-700 rounded text-sm text-slate-200 outline-none focus:border-blue-500 transition-all"
            value={formData['node-type'] || ''}
            onChange={handleChange}
          >
            <option value="service">Service</option>
            <option value="database">Database</option>
            <option value="actor">Actor</option>
            <option value="system">System</option>
            <option value="queue">Queue</option>
          </select>
        </div>

        <div className="space-y-1">
          <label className="block text-[10px] font-bold text-slate-500 uppercase tracking-wider">Owner</label>
          <input 
            name="owner"
            type="text" 
            className="w-full p-2 bg-slate-800 border border-slate-700 rounded text-sm text-slate-200 focus:border-blue-500 outline-none transition-all" 
            value={formData.owner || ''} 
            onChange={handleChange}
          />
        </div>

        <div className="space-y-1">
          <label className="block text-[10px] font-bold text-slate-500 uppercase tracking-wider">Description</label>
          <textarea 
            name="description"
            className="w-full p-2 bg-slate-800 border border-slate-700 rounded text-sm text-slate-200 outline-none focus:border-blue-500 transition-all h-24 resize-none" 
            value={formData.description || ''} 
            onChange={handleChange}
          />
        </div>
      </div>

      <div className="p-4 border-t border-slate-800 bg-slate-900/50 space-y-3">
        <button 
          onClick={handleSave}
          className="w-full flex items-center justify-center gap-2 bg-blue-600 text-white py-2.5 rounded-md font-semibold hover:bg-blue-500 transition-all shadow-lg active:scale-95"
        >
          <Save size={16} /> Save Changes
        </button>
        <button 
          onClick={() => {
            if (confirm('Are you sure you want to delete this node? This will update the Go source code.')) {
              onDelete(selectedNode.id);
            }
          }}
          className="w-full flex items-center justify-center gap-2 bg-slate-800 text-red-400 border border-slate-700 py-2.5 rounded-md font-semibold hover:bg-red-950/30 hover:text-red-300 transition-all active:scale-95"
        >
          <Trash2 size={16} /> Delete Node
        </button>
      </div>
    </aside>
  );
};

export default Sidebar;

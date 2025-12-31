import { useState, useEffect } from 'react';
import { type Node } from 'reactflow';
import { X, Trash2, Save } from 'lucide-react';
import type { CalmNode } from '../types/calm';

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
    <aside className="fixed right-0 top-0 h-full w-80 bg-white shadow-xl border-l border-gray-200 z-50 flex flex-col">
      <div className="p-4 border-bottom bg-gray-50 flex justify-between items-center">
        <h2 className="font-bold text-gray-700">Node Properties</h2>
        <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
          <X size={20} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        <div>
          <label className="block text-xs font-semibold text-gray-500 uppercase mb-1">Unique ID</label>
          <input 
            type="text" 
            className="w-full p-2 bg-gray-100 border border-gray-200 rounded text-sm text-gray-600" 
            value={selectedNode.id} 
            readOnly 
          />
        </div>

        <div>
          <label className="block text-xs font-semibold text-gray-500 uppercase mb-1">Name</label>
          <input 
            name="name"
            type="text" 
            className="w-full p-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-blue-500 outline-none" 
            value={formData.name || ''} 
            onChange={handleChange}
          />
        </div>

        <div>
          <label className="block text-xs font-semibold text-gray-500 uppercase mb-1">Type</label>
          <select 
            name="node-type"
            className="w-full p-2 border border-gray-300 rounded text-sm outline-none"
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

        <div>
          <label className="block text-xs font-semibold text-gray-500 uppercase mb-1">Owner</label>
          <input 
            name="owner"
            type="text" 
            className="w-full p-2 border border-gray-300 rounded text-sm outline-none" 
            value={formData.owner || ''} 
            onChange={handleChange}
          />
        </div>

        <div>
          <label className="block text-xs font-semibold text-gray-500 uppercase mb-1">Description</label>
          <textarea 
            name="description"
            className="w-full p-2 border border-gray-300 rounded text-sm outline-none h-24" 
            value={formData.description || ''} 
            onChange={handleChange}
          />
        </div>
      </div>

      <div className="p-4 border-t bg-gray-50 space-y-2">
        <button 
          onClick={handleSave}
          className="w-full flex items-center justify-center gap-2 bg-blue-600 text-white py-2 rounded font-medium hover:bg-blue-700 transition-colors"
        >
          <Save size={16} /> Save Changes
        </button>
        <button 
          onClick={() => onDelete(selectedNode.id)}
          className="w-full flex items-center justify-center gap-2 bg-white text-red-600 border border-red-200 py-2 rounded font-medium hover:bg-red-50 transition-colors"
        >
          <Trash2 size={16} /> Delete Node
        </button>
      </div>
    </aside>
  );
};

export default Sidebar;

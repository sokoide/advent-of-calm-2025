import { useState, useEffect } from 'react';
import { X, Plus, Trash2, Shield } from 'lucide-react';
import type { CalmControl } from '../domain/calm';

interface ControlsEditorProps {
    controls: Record<string, CalmControl>;
    onAddControl: (controlId: string, desc: string) => void;
    onDeleteControl: (controlId: string) => void;
    onClose: () => void;
}

const ControlsEditor = ({ controls, onAddControl, onDeleteControl, onClose }: ControlsEditorProps) => {
    const [controlId, setControlId] = useState('');
    const [controlDesc, setControlDesc] = useState('');
    const [controlList, setControlList] = useState<Array<{ id: string; control: CalmControl }>>([]);

    useEffect(() => {
        const list = Object.entries(controls || {}).map(([id, ctrl]) => ({ id, control: ctrl }));
        setControlList(list);
    }, [controls]);

    const handleAdd = () => {
        if (!controlId || !controlDesc) {
            alert('ID and Description are required');
            return;
        }
        onAddControl(controlId, controlDesc);
        setControlId('');
        setControlDesc('');
    };

    const handleDelete = (id: string) => {
        if (confirm(`Delete control "${id}"?`)) {
            onDeleteControl(id);
        }
    };

    return (
        <div className="fixed inset-0 z-[110] flex items-center justify-center bg-black/60 backdrop-blur-sm">
            <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-xl flex flex-col max-h-[80vh]">
                <div className="px-6 py-4 border-b border-slate-800 flex justify-between items-center">
                    <h3 className="text-lg font-bold text-slate-200 flex items-center gap-2">
                        <Shield size={20} /> Controls (NFRs)
                    </h3>
                    <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
                        <X size={20} />
                    </button>
                </div>

                <div className="p-6 space-y-4 overflow-y-auto flex-1">
                    <div className="space-y-2">
                        {controlList.length === 0 ? (
                            <div className="text-center p-4 bg-slate-950 rounded border border-slate-800 border-dashed text-slate-500 text-xs">
                                No controls defined yet.
                            </div>
                        ) : (
                            controlList.map(({ id, control }) => (
                                <div key={id} className="flex items-start justify-between bg-slate-800/60 border border-slate-700/50 px-3 py-3 rounded group">
                                    <div>
                                        <div className="text-sm font-semibold text-slate-200">{id}</div>
                                        <div className="text-xs text-slate-400 mt-1">{control.description}</div>
                                    </div>
                                    <button
                                        onClick={() => handleDelete(id)}
                                        className="text-slate-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                                    >
                                        <Trash2 size={14} />
                                    </button>
                                </div>
                            ))
                        )}
                    </div>

                    <div className="border-t border-slate-800 pt-4">
                        <h4 className="text-sm font-semibold text-slate-300 mb-3">Add New Control</h4>
                        <div className="space-y-3">
                            <input
                                type="text"
                                value={controlId}
                                onChange={(e) => setControlId(e.target.value)}
                                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-blue-500"
                                placeholder="Control ID (e.g., security, performance)"
                            />
                            <textarea
                                value={controlDesc}
                                onChange={(e) => setControlDesc(e.target.value)}
                                rows={2}
                                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-blue-500 resize-none"
                                placeholder="Description of the control requirement..."
                            />
                            <button
                                onClick={handleAdd}
                                disabled={!controlId || !controlDesc}
                                className="w-full bg-blue-600/20 hover:bg-blue-600/40 text-blue-400 border border-blue-600/50 px-4 py-2 rounded text-sm font-bold transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                            >
                                <Plus size={16} /> Add Control
                            </button>
                        </div>
                    </div>
                </div>

                <div className="bg-slate-950 px-6 py-4 border-t border-slate-800 flex justify-end">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 rounded text-sm text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ControlsEditor;

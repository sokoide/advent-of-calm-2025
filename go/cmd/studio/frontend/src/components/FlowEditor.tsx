import { useState, useEffect } from 'react';
import { X, Plus, Trash2, ArrowRight, GripVertical } from 'lucide-react';
import type { CalmFlow } from '../domain/calm';
import type { Edge } from 'reactflow';

interface FlowEditorProps {
    flow?: CalmFlow | null; // If null, we are creating a new flow
    relationships: Edge[];  // Available relationships to pick from
    onSave: (flowId: string, name: string, desc: string, steps: string[]) => void;
    onClose: () => void;
}

const FlowEditor = ({ flow, relationships, onSave, onClose }: FlowEditorProps) => {
    const [id, setId] = useState('');
    const [name, setName] = useState('');
    const [desc, setDesc] = useState('');
    const [steps, setSteps] = useState<string[]>([]);
    const [selectedRel, setSelectedRel] = useState('');

    const [draggedIndex, setDraggedIndex] = useState<number | null>(null);

    useEffect(() => {
        if (flow) {
            setId(flow['unique-id']);
            setName(flow.name);
            setDesc(flow.description || '');
            const currentSteps = flow.transitions?.map(t => t['relationship-unique-id']) || [];
            setSteps(currentSteps);
        } else {
            // New flow
            setId('');
            setName('');
            setDesc('');
            setSteps([]);
        }
    }, [flow]);

    const handleAddStep = () => {
        if (selectedRel && !steps.includes(selectedRel)) {
            setSteps([...steps, selectedRel]);
            setSelectedRel('');
        }
    };

    const handleRemoveStep = (stepId: string) => {
        setSteps(steps.filter(s => s !== stepId));
    };

    const handleDragStart = (e: React.DragEvent, idx: number) => {
        setDraggedIndex(idx);
        e.dataTransfer.effectAllowed = "move";
        // Required for Firefox
        e.dataTransfer.setData("text/plain", idx.toString());
    };

    const handleDragOver = (e: React.DragEvent, _idx: number) => {
        e.preventDefault();
        e.dataTransfer.dropEffect = "move";
    };

    const handleDrop = (e: React.DragEvent, dropIndex: number) => {
        e.preventDefault();
        if (draggedIndex === null || draggedIndex === dropIndex) return;

        const newSteps = [...steps];
        const [movedItem] = newSteps.splice(draggedIndex, 1);
        newSteps.splice(dropIndex, 0, movedItem);

        setSteps(newSteps);
        setDraggedIndex(null);
    };

    const handleSave = () => {
        if (!id || !name) {
            alert("ID and Name are required");
            return;
        }
        onSave(id, name, desc, steps);
    };

    // Helper to get relationship label or ID
    const getRelLabel = (relId: string) => {
        const edge = relationships.find(e => e.id === relId);
        if (edge) {
            return `${edge.source} -> ${edge.target} (${relId})`;
        }
        return relId;
    };

    return (
        <div className="fixed inset-0 z-[110] flex items-center justify-center bg-black/60 backdrop-blur-sm">
            <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-2xl flex flex-col max-h-[90vh]">
                <div className="px-6 py-4 border-b border-slate-800 flex justify-between items-center">
                    <h3 className="text-lg font-bold text-slate-200">
                        {flow ? 'Edit Flow' : 'New Flow'}
                    </h3>
                    <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
                        <X size={20} />
                    </button>
                </div>

                <div className="p-6 space-y-4 overflow-y-auto flex-1">
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-xs font-semibold text-slate-500 mb-1">Unique ID</label>
                            <input
                                type="text"
                                value={id}
                                onChange={(e) => setId(e.target.value)}
                                disabled={!!flow} // Cannot change ID of existing flow
                                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                                placeholder="e.g. order-flow"
                            />
                        </div>
                        <div>
                            <label className="block text-xs font-semibold text-slate-500 mb-1">Flow Name</label>
                            <input
                                type="text"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-blue-500"
                                placeholder="e.g. Order Processing"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-xs font-semibold text-slate-500 mb-1">Description</label>
                        <textarea
                            value={desc}
                            onChange={(e) => setDesc(e.target.value)}
                            rows={2}
                            className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-blue-500 resize-none"
                            placeholder="Describe the flow..."
                        />
                    </div>

                    <div className="border-t border-slate-800 pt-4">
                        <h4 className="text-sm font-semibold text-slate-300 mb-3 flex items-center gap-2">
                            <ArrowRight size={16} /> Steps (Transitions)
                        </h4>

                        <div className="space-y-2 mb-4">
                            {steps.length === 0 ? (
                                <div className="text-center p-4 bg-slate-950 rounded border border-slate-800 border-dashed text-slate-500 text-xs">
                                    No steps defined. Add steps below.
                                </div>
                            ) : (
                                steps.map((stepId, idx) => (
                                    <div
                                        key={`${stepId}-${idx}`}
                                        className={`flex items-center justify-between bg-slate-800/60 border border-slate-700/50 px-3 py-2 rounded group transition-all ${draggedIndex === idx ? 'opacity-50 border-blue-500 border-dashed' : 'hover:border-slate-600'}`}
                                        draggable={true}
                                        onDragStart={(e) => handleDragStart(e, idx)}
                                        onDragOver={(e) => handleDragOver(e, idx)}
                                        onDrop={(e) => handleDrop(e, idx)}
                                        onDragEnd={() => setDraggedIndex(null)}
                                    >
                                        <div className="flex items-center gap-3">
                                            <GripVertical size={14} className="text-slate-600 cursor-grab active:cursor-grabbing" />
                                            <span className="bg-slate-900 text-slate-500 text-[10px] font-mono w-5 h-5 flex items-center justify-center rounded-full">
                                                {idx + 1}
                                            </span>
                                            <div className="text-sm text-slate-300 font-mono">{getRelLabel(stepId)}</div>
                                        </div>
                                        <button
                                            onClick={() => handleRemoveStep(stepId)}
                                            className="text-slate-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                                        >
                                            <Trash2 size={14} />
                                        </button>
                                    </div>
                                ))
                            )}
                        </div>

                        <div className="flex gap-2">
                            <select
                                value={selectedRel}
                                onChange={(e) => setSelectedRel(e.target.value)}
                                className="flex-1 bg-slate-950 border border-slate-700 rounded px-3 py-2 text-xs text-slate-300 focus:outline-none focus:border-blue-500"
                            >
                                <option value="">Select a relationship to add as step...</option>
                                {relationships.map((rel) => (
                                    <option key={rel.id} value={rel.id}>
                                        {rel.source} {'->'} {rel.target} ({rel.id})
                                    </option>
                                ))}
                            </select>
                            <button
                                onClick={handleAddStep}
                                disabled={!selectedRel}
                                className="bg-blue-600/20 hover:bg-blue-600/40 text-blue-400 border border-blue-600/50 px-3 py-2 rounded text-xs font-bold transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                            >
                                <Plus size={14} /> Add Step
                            </button>
                        </div>
                    </div>
                </div>

                <div className="bg-slate-950 px-6 py-4 border-t border-slate-800 flex justify-end gap-3">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 rounded text-sm text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        className="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded text-sm font-bold shadow-lg transition-transform active:scale-95"
                    >
                        Save Flow
                    </button>
                </div>
            </div>
        </div>
    );
};

export default FlowEditor;

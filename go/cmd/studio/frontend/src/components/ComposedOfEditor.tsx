import { useState, useMemo, useEffect } from 'react';
import { X, Plus, Trash2, Layers } from 'lucide-react';
import type { Node } from 'reactflow';
import type { CalmRelationship } from '../domain/calm';

interface ComposedOfRelation {
    id: string;
    container: string;
    nodes: string[];
}

interface ComposedOfEditorProps {
    nodes: Node[];
    relationships: CalmRelationship[];
    onSave: (containerId: string, childNodeIds: string[]) => void;
    onDelete: (composedOfId: string) => void;
    onClose: () => void;
}

const ComposedOfEditor = ({ nodes, relationships, onSave, onDelete, onClose }: ComposedOfEditorProps) => {
    const [containerId, setContainerId] = useState('');
    const [selectedChildren, setSelectedChildren] = useState<string[]>([]);
    const [selectedChild, setSelectedChild] = useState('');
    const [existingComposedOf, setExistingComposedOf] = useState<ComposedOfRelation[]>([]);

    // Extract ComposedOf relationships from relationships
    useEffect(() => {
        const composed: ComposedOfRelation[] = [];
        for (const rel of relationships) {
            const composedOfData = rel['relationship-type']?.['composed-of'];
            if (composedOfData) {
                composed.push({
                    id: rel['unique-id'],
                    container: composedOfData.container,
                    nodes: composedOfData.nodes || [],
                });
            }
        }
        setExistingComposedOf(composed);
    }, [relationships]);

    const containerNodes = useMemo(() => {
        return nodes.filter(n => n.type === 'system' || n.type === 'service' || n.type === 'group');
    }, [nodes]);

    const availableChildren = useMemo(() => {
        return nodes.filter(n => n.id !== containerId && !selectedChildren.includes(n.id));
    }, [nodes, containerId, selectedChildren]);

    const handleAddChild = () => {
        if (selectedChild && !selectedChildren.includes(selectedChild)) {
            setSelectedChildren([...selectedChildren, selectedChild]);
            setSelectedChild('');
        }
    };

    const handleRemoveChild = (nodeId: string) => {
        setSelectedChildren(selectedChildren.filter(id => id !== nodeId));
    };

    const handleSave = () => {
        if (!containerId || selectedChildren.length === 0) {
            alert('Container and at least one child node are required');
            return;
        }
        onSave(containerId, selectedChildren);
        setContainerId('');
        setSelectedChildren([]);
    };

    const handleDelete = (id: string) => {
        if (confirm(`Delete this ComposedOf relationship?`)) {
            onDelete(id);
        }
    };

    const getNodeLabel = (nodeId: string) => {
        const node = nodes.find(n => n.id === nodeId);
        return node?.data?.label || nodeId;
    };

    return (
        <div className="fixed inset-0 z-[110] flex items-center justify-center bg-black/60 backdrop-blur-sm">
            <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-xl flex flex-col max-h-[80vh]">
                <div className="px-6 py-4 border-b border-slate-800 flex justify-between items-center">
                    <h3 className="text-lg font-bold text-slate-200 flex items-center gap-2">
                        <Layers size={20} /> ComposedOf Relationships
                    </h3>
                    <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
                        <X size={20} />
                    </button>
                </div>

                <div className="p-6 space-y-6 overflow-y-auto flex-1">
                    {/* Existing ComposedOf relationships */}
                    <div>
                        <h4 className="text-sm font-semibold text-slate-300 mb-3">Existing Relationships</h4>
                        <div className="space-y-2">
                            {existingComposedOf.length === 0 ? (
                                <div className="text-center p-3 bg-slate-950 rounded border border-slate-800 border-dashed text-slate-500 text-xs">
                                    No ComposedOf relationships yet.
                                </div>
                            ) : (
                                existingComposedOf.map((rel) => (
                                    <div key={rel.id} className="bg-slate-800/60 border border-slate-700/50 px-3 py-3 rounded group">
                                        <div className="flex items-start justify-between">
                                            <div>
                                                <div className="text-sm font-semibold text-slate-200">
                                                    {getNodeLabel(rel.container)}
                                                </div>
                                                <div className="text-xs text-slate-400 mt-1">
                                                    Contains: {rel.nodes.map(n => getNodeLabel(n)).join(', ')}
                                                </div>
                                                <div className="text-[10px] text-slate-500 font-mono mt-1">{rel.id}</div>
                                            </div>
                                            <button
                                                onClick={() => handleDelete(rel.id)}
                                                className="text-slate-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                                            >
                                                <Trash2 size={14} />
                                            </button>
                                        </div>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>

                    {/* Create new ComposedOf */}
                    <div className="border-t border-slate-800 pt-4">
                        <h4 className="text-sm font-semibold text-slate-300 mb-3">Create New</h4>
                        <div className="space-y-3">
                            <div>
                                <label className="block text-xs text-slate-500 mb-1">Container Node</label>
                                <select
                                    value={containerId}
                                    onChange={(e) => setContainerId(e.target.value)}
                                    className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-blue-500"
                                >
                                    <option value="">Select container node...</option>
                                    {containerNodes.map((node) => (
                                        <option key={node.id} value={node.id}>
                                            {node.data?.label || node.id} ({node.type})
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div className="space-y-2">
                                {selectedChildren.map((nodeId, idx) => (
                                    <div key={nodeId} className="flex items-center justify-between bg-slate-800/60 border border-slate-700/50 px-3 py-2 rounded text-sm text-slate-300">
                                        <span>{idx + 1}. {getNodeLabel(nodeId)}</span>
                                        <button onClick={() => handleRemoveChild(nodeId)} className="text-slate-500 hover:text-red-400">
                                            <Trash2 size={14} />
                                        </button>
                                    </div>
                                ))}
                            </div>

                            <div className="flex gap-2">
                                <select
                                    value={selectedChild}
                                    onChange={(e) => setSelectedChild(e.target.value)}
                                    className="flex-1 bg-slate-950 border border-slate-700 rounded px-3 py-2 text-xs text-slate-300 focus:outline-none focus:border-blue-500"
                                >
                                    <option value="">Select child node...</option>
                                    {availableChildren.map((node) => (
                                        <option key={node.id} value={node.id}>
                                            {node.data?.label || node.id}
                                        </option>
                                    ))}
                                </select>
                                <button
                                    onClick={handleAddChild}
                                    disabled={!selectedChild}
                                    className="bg-blue-600/20 hover:bg-blue-600/40 text-blue-400 border border-blue-600/50 px-3 py-2 rounded text-xs font-bold transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                                >
                                    <Plus size={14} /> Add
                                </button>
                            </div>

                            <button
                                onClick={handleSave}
                                disabled={!containerId || selectedChildren.length === 0}
                                className="w-full bg-blue-600 hover:bg-blue-500 text-white rounded py-2 text-sm font-bold transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                Create Relationship
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

export default ComposedOfEditor;

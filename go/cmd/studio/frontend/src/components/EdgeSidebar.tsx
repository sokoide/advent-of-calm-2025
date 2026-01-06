import { X } from 'lucide-react';
import type { Edge } from 'reactflow';

interface EdgeSidebarProps {
    selectedEdge: Edge | null;
    onClose: () => void;
}

const EdgeSidebar = ({ selectedEdge, onClose }: EdgeSidebarProps) => {
    if (!selectedEdge) return null;

    // Extract calm relationship data from edge
    const calm = selectedEdge.data?.calm || {};
    const relationshipType = calm['relationship-type'];

    return (
        <div className="absolute right-0 top-0 h-full w-80 bg-zinc-900/95 border-l border-zinc-700 backdrop-blur-md shadow-2xl z-50 flex flex-col">
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-zinc-700/50">
                <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-purple-400"></div>
                    <h2 className="text-lg font-semibold text-white">Relationship</h2>
                </div>
                <button
                    onClick={onClose}
                    className="p-1.5 rounded-md hover:bg-zinc-800 transition-colors text-zinc-400 hover:text-white"
                    title="Close"
                >
                    <X size={18} />
                </button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {/* ID */}
                <div className="space-y-1">
                    <label className="text-xs font-medium text-zinc-400 uppercase tracking-wide">
                        ID
                    </label>
                    <div className="px-3 py-2 rounded-lg bg-zinc-800/50 border border-zinc-700/50 text-zinc-300 text-sm font-mono">
                        {selectedEdge.id}
                    </div>
                </div>

                {/* Source & Target */}
                <div className="grid grid-cols-2 gap-3">
                    <div className="space-y-1">
                        <label className="text-xs font-medium text-zinc-400 uppercase tracking-wide">
                            Source
                        </label>
                        <div className="px-3 py-2 rounded-lg bg-zinc-800/50 border border-zinc-700/50 text-zinc-300 text-sm">
                            {selectedEdge.source}
                        </div>
                    </div>
                    <div className="space-y-1">
                        <label className="text-xs font-medium text-zinc-400 uppercase tracking-wide">
                            Target
                        </label>
                        <div className="px-3 py-2 rounded-lg bg-zinc-800/50 border border-zinc-700/50 text-zinc-300 text-sm">
                            {selectedEdge.target}
                        </div>
                    </div>
                </div>

                {/* Description */}
                {calm.description && (
                    <div className="space-y-1">
                        <label className="text-xs font-medium text-zinc-400 uppercase tracking-wide">
                            Description
                        </label>
                        <div className="px-3 py-2 rounded-lg bg-zinc-800/50 border border-zinc-700/50 text-zinc-300 text-sm">
                            {calm.description}
                        </div>
                    </div>
                )}

                {/* Relationship Type */}
                {relationshipType && (
                    <div className="space-y-1">
                        <label className="text-xs font-medium text-zinc-400 uppercase tracking-wide">
                            Type
                        </label>
                        <div className="px-3 py-2 rounded-lg bg-zinc-800/50 border border-zinc-700/50 text-zinc-300 text-sm">
                            {relationshipType.connects ? 'Connects' : relationshipType.interacts ? 'Interacts' : 'ComposedOf'}
                        </div>
                    </div>
                )}

                {/* Metadata */}
                {calm.metadata && Object.keys(calm.metadata).length > 0 && (
                    <div className="space-y-2">
                        <label className="text-xs font-medium text-zinc-400 uppercase tracking-wide">
                            Metadata
                        </label>
                        <div className="space-y-1.5">
                            {Object.entries(calm.metadata).map(([key, value]) => (
                                <div
                                    key={key}
                                    className="flex items-center justify-between px-3 py-2 rounded-lg bg-zinc-800/50 border border-zinc-700/50"
                                >
                                    <span className="text-xs text-zinc-400">{key}</span>
                                    <span className="text-xs text-zinc-200 font-mono">
                                        {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                                    </span>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {/* Properties from connects/interacts */}
                {relationshipType?.connects && (
                    <div className="space-y-2">
                        <label className="text-xs font-medium text-zinc-400 uppercase tracking-wide">
                            Connection Details
                        </label>
                        <div className="space-y-1.5 text-xs">
                            {relationshipType.connects.protocol && (
                                <div className="flex justify-between px-3 py-1.5 bg-zinc-800/30 rounded">
                                    <span className="text-zinc-400">Protocol</span>
                                    <span className="text-zinc-200">{relationshipType.connects.protocol}</span>
                                </div>
                            )}
                            {relationshipType.connects['data-classification'] && (
                                <div className="flex justify-between px-3 py-1.5 bg-zinc-800/30 rounded">
                                    <span className="text-zinc-400">Classification</span>
                                    <span className="text-zinc-200">{relationshipType.connects['data-classification']}</span>
                                </div>
                            )}
                            <div className="flex justify-between px-3 py-1.5 bg-zinc-800/30 rounded">
                                <span className="text-zinc-400">Encrypted</span>
                                <span className="text-zinc-200">{relationshipType.connects['connection-encrypted'] ? 'Yes' : 'No'}</span>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default EdgeSidebar;

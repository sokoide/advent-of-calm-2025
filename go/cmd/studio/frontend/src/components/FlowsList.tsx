import { Trash2, AlertCircle } from 'lucide-react';
import type { CalmFlow } from '../domain/calm';

interface FlowsListProps {
    flows: CalmFlow[];
    onDeleteFlow: (flowId: string) => void;
    onClose: () => void;
}

const FlowsList = ({ flows, onDeleteFlow, onClose }: FlowsListProps) => {
    return (
        <div className="absolute left-4 top-16 w-80 bg-slate-900 shadow-2xl border border-slate-800 z-[100] flex flex-col rounded-lg animate-in fade-in zoom-in duration-200">
            <div className="p-3 border-b border-slate-800 bg-slate-900/50 flex justify-between items-center rounded-t-lg">
                <h2 className="font-bold text-slate-200 text-sm flex items-center gap-2">
                    Defind Flows
                    <span className="bg-slate-800 text-slate-400 text-[10px] px-1.5 py-0.5 rounded-full">{flows.length}</span>
                </h2>
                <button onClick={onClose} className="text-slate-500 hover:text-slate-300 transition-colors text-xs">
                    Close
                </button>
            </div>

            <div className="max-h-[60vh] overflow-y-auto p-2 space-y-2">
                {flows.length === 0 ? (
                    <div className="text-center p-6 text-slate-500 text-xs italic">
                        No flows defined in this architecture.
                        <br />
                        Create flows in Go code using <code className="bg-slate-800 px-1 rounded">a.DefineFlow()</code>.
                    </div>
                ) : (
                    flows.map((flow) => (
                        <div key={flow['unique-id']} className="bg-slate-800/50 border border-slate-700/50 rounded p-3 group hover:border-blue-500/30 transition-all">
                            <div className="flex justify-between items-start gap-2">
                                <div>
                                    <div className="text-sm font-semibold text-blue-300">{flow.name}</div>
                                    <div className="text-[10px] text-slate-500 font-mono mt-0.5">{flow['unique-id']}</div>
                                </div>
                                <button
                                    onClick={() => {
                                        if (confirm(`Delete flow "${flow.name}"?`)) {
                                            onDeleteFlow(flow['unique-id']);
                                        }
                                    }}
                                    className="text-slate-600 hover:text-red-400 p-1 rounded hover:bg-red-950/20 transition-colors opacity-0 group-hover:opacity-100"
                                    title="Delete Flow"
                                >
                                    <Trash2 size={14} />
                                </button>
                            </div>

                            <div className="mt-2 text-xs text-slate-400 line-clamp-2">
                                {flow.description}
                            </div>

                            <div className="mt-2 flex items-center gap-2 text-[10px] text-slate-500">
                                <span className="bg-slate-800 px-1.5 py-0.5 rounded">{flow.transitions?.length || 0} Steps</span>
                                {flow.metadata?.sla && (
                                    <span className="flex items-center gap-1 text-yellow-500/80">
                                        <AlertCircle size={10} /> {String(flow.metadata.sla).split(',')[0]}
                                    </span>
                                )}
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default FlowsList;

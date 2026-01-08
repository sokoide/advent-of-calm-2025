import { memo } from 'react';
import { BaseEdge, EdgeLabelRenderer, getBezierPath, type EdgeProps } from 'reactflow';
import { X, AlertTriangle } from 'lucide-react';

interface DeletableEdgeProps extends EdgeProps {
    data?: {
        onDelete?: (id: string) => void;
        referencedInFlow?: boolean;
        flowNames?: string[];
    };
}

const DeletableEdge = ({
    id,
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
    style = {},
    markerEnd,
    label,
    data,
}: DeletableEdgeProps) => {
    const [edgePath, labelX, labelY] = getBezierPath({
        sourceX,
        sourceY,
        sourcePosition,
        targetX,
        targetY,
        targetPosition,
    });

    const isProtected = data?.referencedInFlow;
    const flowNames = data?.flowNames || [];

    const onEdgeClick = (evt: React.MouseEvent) => {
        evt.stopPropagation();

        if (isProtected) {
            alert(`This Relationship is referenced in a Flow and cannot be deleted.\n\nReferenced Flows:\n- ${flowNames.join('\n- ')}`);
            return;
        }

        if (data?.onDelete) {
            if (confirm('Are you sure you want to delete this Relationship?')) {
                data.onDelete(id);
            }
        }
    };

    return (
        <>
            <BaseEdge path={edgePath} markerEnd={markerEnd} style={style} />
            <EdgeLabelRenderer>
                <div
                    style={{
                        position: 'absolute',
                        transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
                        pointerEvents: 'all',
                    }}
                    className="nodrag nopan group"
                >
                    {label && (
                        <div className="bg-slate-800/90 text-slate-300 text-xs px-2 py-1 rounded border border-slate-700 mb-1 flex items-center gap-1">
                            {isProtected && <AlertTriangle size={12} className="text-yellow-500" />}
                            {label}
                        </div>
                    )}
                    <button
                        onClick={onEdgeClick}
                        className={`opacity-0 group-hover:opacity-100 transition-opacity rounded-full p-1 shadow-lg ${isProtected
                                ? 'bg-gray-600 hover:bg-gray-500 text-gray-300 cursor-not-allowed'
                                : 'bg-red-600 hover:bg-red-500 text-white'
                            }`}
                        title={isProtected ? 'Cannot delete: referenced in a Flow' : 'Delete Relationship'}
                    >
                        <X size={12} />
                    </button>
                </div>
            </EdgeLabelRenderer>
        </>
    );
};

export default memo(DeletableEdge);

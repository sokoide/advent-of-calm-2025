import { memo } from 'react';
import { BaseEdge, EdgeLabelRenderer, getBezierPath, type EdgeProps } from 'reactflow';
import { X } from 'lucide-react';

interface DeletableEdgeProps extends EdgeProps {
    data?: {
        onDelete?: (id: string) => void;
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

    const onEdgeClick = (evt: React.MouseEvent) => {
        evt.stopPropagation();
        if (data?.onDelete) {
            if (confirm('Are you sure you want to delete this relationship?')) {
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
                        <div className="bg-slate-800/90 text-slate-300 text-xs px-2 py-1 rounded border border-slate-700 mb-1">
                            {label}
                        </div>
                    )}
                    <button
                        onClick={onEdgeClick}
                        className="opacity-0 group-hover:opacity-100 transition-opacity bg-red-600 hover:bg-red-500 text-white rounded-full p-1 shadow-lg"
                        title="Delete relationship"
                    >
                        <X size={12} />
                    </button>
                </div>
            </EdgeLabelRenderer>
        </>
    );
};

export default memo(DeletableEdge);

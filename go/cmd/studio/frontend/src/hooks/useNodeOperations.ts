import { useCallback } from 'react';
import type { Edge, Node } from 'reactflow';
import type { CalmNode, NodeOrigin } from '../domain/calm';
import type { StudioUseCase } from '../usecase/studio';
import type { PatchOrigin, PatchOperation } from '../domain/ports';

// Convert NodeOrigin (from CALM JSON) to PatchOrigin (for backend)
function toPatchOrigin(origin: NodeOrigin): PatchOrigin {
    return {
        file: origin.file,
        line: origin.line,
        loopVar: origin.loopVar,
        loopMax: origin.loopMax,
    };
}

interface UseNodeOperationsProps {
    studio: StudioUseCase | null;
    nodes: Node[];
    edges: Edge[];
    setNodes: (nodes: Node[] | ((nds: Node[]) => Node[])) => void;
    setEdges: (edges: Edge[] | ((eds: Edge[]) => Edge[])) => void;
    saveLayout: (nodes: Node[]) => Promise<void>;
    fetchData: (isWSUpdate?: boolean) => Promise<void>;
    setSelectedNode: (node: Node | null) => void;
}

export function useNodeOperations({
    studio,
    nodes,
    edges,
    setNodes,
    setEdges,
    saveLayout,
    fetchData,
    setSelectedNode,
}: UseNodeOperationsProps) {
    const onAddNode = useCallback(async () => {
        if (!studio) return;
        const id = `node-${Date.now()}`;
        const newNode: Node = {
            id,
            type: 'service',
            position: { x: 100, y: 100 },
            data: {
                label: 'New Node',
                calm: {
                    'unique-id': id,
                    'node-type': 'service',
                    name: 'New Node',
                    description: '',
                },
            },
        };
        setNodes((nds) => nds.concat(newNode));
        saveLayout(nodes.concat(newNode));

        await studio.syncAST({
            action: 'add',
            nodeId: id,
            nodeType: 'Service',
            name: 'New Node',
            desc: '',
        });
        fetchData(true);
    }, [nodes, setNodes, saveLayout, fetchData, studio]);

    const onUpdateNode = useCallback(
        async (id: string, updatedCalm: Partial<CalmNode>) => {
            if (!studio) return;

            // Find the node to check for origin
            const node = nodes.find(n => n.id === id);
            const origin = node?.data?.calm?._origin as NodeOrigin | undefined;

            if (origin) {
                // Use Patch API with proper origin conversion
                const patchOrigin = toPatchOrigin(origin);
                const ops: PatchOperation[] = [];
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'name')) {
                    ops.push({ type: 'update-node', origin: patchOrigin, property: 'name', value: updatedCalm.name ?? '' });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'owner')) {
                    ops.push({ type: 'update-node', origin: patchOrigin, property: 'owner', value: updatedCalm.owner ?? '' });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'description')) {
                    ops.push({ type: 'update-node', origin: patchOrigin, property: 'description', value: updatedCalm.description ?? '' });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'node-type')) {
                    ops.push({ type: 'update-node', origin: patchOrigin, property: 'node-type', value: updatedCalm['node-type'] ?? '' });
                }

                if (ops.length > 0) {
                    await studio.patchAST(ops);
                }
            } else {
                // Fallback to legacy Sync API
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'name')) {
                    await studio.syncAST({
                        action: 'update',
                        nodeId: id,
                        property: 'name',
                        value: updatedCalm.name ?? '',
                    });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'owner')) {
                    await studio.syncAST({
                        action: 'update',
                        nodeId: id,
                        property: 'owner',
                        value: updatedCalm.owner ?? '',
                    });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'description')) {
                    await studio.syncAST({
                        action: 'update',
                        nodeId: id,
                        property: 'description',
                        value: updatedCalm.description ?? '',
                    });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'node-type')) {
                    await studio.syncAST({
                        action: 'update',
                        nodeId: id,
                        property: 'node-type',
                        value: updatedCalm['node-type'] ?? '',
                    });
                }
            }
            fetchData(true);
        },
        [fetchData, studio, nodes]
    );

    const onDeleteNode = useCallback(
        async (id: string) => {
            if (!studio) return;

            const node = nodes.find(n => n.id === id);
            const origin = node?.data?.calm?._origin as NodeOrigin | undefined;

            const remainingNodes = nodes.filter((node) => node.id !== id);
            const remainingEdges = edges.filter((edge) => edge.source !== id && edge.target !== id);
            setNodes(remainingNodes);
            setEdges(remainingEdges);
            saveLayout(remainingNodes);

            if (origin) {
                const patchOrigin = toPatchOrigin(origin);
                await studio.patchAST([{ type: 'delete-node', nodeId: id, origin: patchOrigin }]);
            } else {
                await studio.syncAST({ action: 'delete', nodeId: id });
            }

            setSelectedNode(null);
            setTimeout(() => {
                fetchData(true);
            }, 500);
        },
        [edges, fetchData, nodes, saveLayout, setEdges, setNodes, setSelectedNode, studio]
    );

    const onNodeDragStop = useCallback(
        async (_: any, node: Node) => {
            saveLayout(nodes.map((n) => (n.id === node.id ? node : n)));
        },
        [nodes, saveLayout]
    );

    return {
        onAddNode,
        onUpdateNode,
        onDeleteNode,
        onNodeDragStop,
    };
}

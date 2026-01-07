import { useCallback } from 'react';
import type { Edge, Node } from 'reactflow';
import type { CalmNode, NodeOrigin, CalmRelationship } from '../domain/calm';
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
    relationships: CalmRelationship[];
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
    relationships,
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

        // Use Patch API for consistency
        await studio.patchAST([{
            type: 'add-node',
            nodeId: id,
            nodeName: 'New Node',
            nodeTypeName: 'Service',
            nodeDesc: '',
        }]);
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
                // Fallback to nodeId lookup via Patch API
                const ops: PatchOperation[] = [];
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'name')) {
                    ops.push({ type: 'update-node', nodeId: id, property: 'name', value: updatedCalm.name ?? '' });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'owner')) {
                    ops.push({ type: 'update-node', nodeId: id, property: 'owner', value: updatedCalm.owner ?? '' });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'description')) {
                    ops.push({ type: 'update-node', nodeId: id, property: 'description', value: updatedCalm.description ?? '' });
                }
                if (Object.prototype.hasOwnProperty.call(updatedCalm, 'node-type')) {
                    ops.push({ type: 'update-node', nodeId: id, property: 'node-type', value: updatedCalm['node-type'] ?? '' });
                }

                if (ops.length > 0) {
                    await studio.patchAST(ops);
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
                await studio.patchAST([{ type: 'delete-node', nodeId: id }]);
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

    const onDuplicateNode = useCallback(
        async (id: string) => {
            if (!studio) return;

            const sourceNode = nodes.find(n => n.id === id);
            if (!sourceNode) return;

            const newId = `${id}-copy-${Date.now()}`;
            const sourceCalm = sourceNode.data.calm || {};

            console.log('Duplicating node:', id);
            console.log('Source CALM data:', sourceCalm);
            console.log('Interfaces:', sourceCalm.interfaces);
            console.log('Metadata:', sourceCalm.metadata);

            // Create a new node based on the source
            const newNode: Node = {
                id: newId,
                type: sourceNode.type,
                position: {
                    x: sourceNode.position.x + 50,
                    y: sourceNode.position.y + 50,
                },
                data: {
                    label: `${sourceCalm.name || 'Node'} (Copy)`,
                    calm: {
                        'unique-id': newId,
                        'node-type': sourceCalm['node-type'] || 'service',
                        name: `${sourceCalm.name || 'Node'} (Copy)`,
                        description: sourceCalm.description || '',
                    },
                },
            };

            setNodes((nds) => nds.concat(newNode));
            saveLayout(nodes.concat(newNode));

            // Normalize node type
            const rawType = sourceCalm['node-type'] || 'Service';
            const nodeTypeName = rawType.charAt(0).toUpperCase() + rawType.slice(1);

            // Use Patch API to add node
            const ops: PatchOperation[] = [{
                type: 'add-node',
                nodeId: newId,
                nodeName: newNode.data.calm.name,
                nodeTypeName: nodeTypeName,
                nodeDesc: sourceCalm.description || '',
            }];

            // Copy owner if present
            if (sourceCalm.owner) {
                ops.push({
                    type: 'update-node',
                    nodeId: newId,
                    property: 'owner',
                    value: sourceCalm.owner,
                });
            }

            // Copy interfaces if present
            if (sourceCalm.interfaces && Array.isArray(sourceCalm.interfaces)) {
                console.log(`Copying ${sourceCalm.interfaces.length} interfaces`);
                sourceCalm.interfaces.forEach((iface: any) => {
                    // Preserve original interface ID with copy suffix
                    const originalIfaceId = iface['unique-id'] || 'interface';
                    const interfaceId = `${originalIfaceId}-copy-${Date.now()}`;
                    console.log(`Adding interface ${interfaceId} with protocol ${iface.protocol}`);
                    ops.push({
                        type: 'add-interface',
                        nodeId: newId,
                        interfaceId: interfaceId,
                        protocol: iface.protocol || 'HTTP',
                    });
                });
            } else {
                console.log('No interfaces to copy');
            }

            // Copy metadata if present
            // NOTE: Metadata copying is currently disabled due to complexity in backend AST manipulation
            // Metadata should be manually added to duplicated nodes if needed
            const hasMetadata = sourceCalm.metadata && typeof sourceCalm.metadata === 'object' && Object.keys(sourceCalm.metadata).length > 0;
            if (hasMetadata) {
                console.log('⚠️ Note: Metadata was not copied. Please manually add metadata if needed.');
                console.log('   Metadata found:', Object.keys(sourceCalm.metadata));
                console.log('   Full metadata copying will be implemented in the future.');
                // TODO: Implement metadata copying in backend
                /*
                Object.entries(sourceCalm.metadata).forEach(([key, value]) => {
                    if (key === 'owner') {
                        return; // Already handled above
                    }
                    ops.push({
                        type: 'update-node',
                        nodeId: newId,
                        property: `metadata.${key}`,
                        value: value,
                    });
                });
                */
            }

            // Check if source node is part of a ComposedOf relationship
            const parentComposedOf = relationships.find(rel => {
                const composedOf = rel['relationship-type']?.['composed-of'];
                return composedOf && composedOf.nodes.includes(id);
            });

            if (parentComposedOf) {
                console.log(`Adding duplicated node to parent ComposedOf: ${parentComposedOf['unique-id']}`);
                ops.push({
                    type: 'update-relationship',
                    composedOfId: parentComposedOf['unique-id'],
                    property: 'add-child-node',
                    value: newId,
                });
            }

            console.log('Patch operations:', ops);
            await studio.patchAST(ops);

            setTimeout(() => {
                fetchData(true);
            }, 500);
        },
        [nodes, setNodes, saveLayout, fetchData, studio]
    );

    return {
        onAddNode,
        onUpdateNode,
        onDeleteNode,
        onNodeDragStop,
        onDuplicateNode,
    };
}

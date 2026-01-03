import { useCallback, useRef, useEffect, useState } from 'react';
import { useNodesState, useEdgesState, type Node } from 'reactflow';
import type { CalmArchitecture, LayoutData } from '../domain/calm';
import { buildParentMap, parentMapEquals } from '../domain/architecture';
import { transformToReactFlow } from '../utils/transformer';
import { getLayoutedElements } from '../utils/layout';
import type { StudioUseCase } from '../usecase/studio';

interface ContentState {
    goCode: string;
    d2Code: string;
    jsonCode: string;
    svgCode: string;
    archId: string;
    loading: boolean;
}

export function useStudioContent(studio: StudioUseCase | null) {
    const [state, setState] = useState<ContentState>({
        goCode: '',
        d2Code: '',
        jsonCode: '',
        svgCode: '',
        archId: '',
        loading: true,
    });
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    const isUpdating = useRef(false);

    const saveLayout = useCallback(async (currentNodes: Node[]) => {
        if (!studio || !state.archId) return;
        try {
            const newLayout: LayoutData = { nodes: {}, parentMap: {} };
            currentNodes.forEach((n) => {
                newLayout.nodes[n.id] = n.position;
                if (n.parentNode) {
                    newLayout.parentMap![n.id] = n.parentNode;
                }
            });
            await studio.saveLayout(state.archId, newLayout);
        } catch (err) {
            console.error('Failed to save layout:', err);
        }
    }, [studio, state.archId]);

    const fetchData = useCallback(async (isWSUpdate = false) => {
        if (!studio) return;
        if (isUpdating.current && !isWSUpdate) return;
        try {
            const { goCode: remoteGo, d2Code: remoteD2, json, svg } = await studio.fetchContent();

            setState((prev) => ({
                ...prev,
                goCode: remoteGo,
                d2Code: remoteD2,
                svgCode: svg || prev.svgCode,
            }));

            if (!json || json === '' || json === 'null' || json === '{}') {
                console.log('⚠️ JSON is empty, showing code only');
                setState((prev) => ({ ...prev, loading: false }));
                return;
            }

            const arch: CalmArchitecture = JSON.parse(json);
            const archUniqueID = arch['unique-id'];
            setState((prev) => ({
                ...prev,
                jsonCode: JSON.stringify(arch, null, 2),
                archId: archUniqueID,
            }));

            const layout: LayoutData = await studio.fetchLayout(archUniqueID);
            const { nodes: initialNodes, edges: initialEdges } = transformToReactFlow(arch, layout);
            const currentParentMap = buildParentMap(arch);

            const hasStoredLayout = layout.nodes && Object.keys(layout.nodes).length > 0;
            const parentMapMatches = parentMapEquals(layout.parentMap, currentParentMap);

            if (!hasStoredLayout || !parentMapMatches) {
                const layouted = getLayoutedElements(initialNodes, initialEdges, 'LR');
                setNodes(layouted.nodes);
                setEdges(initialEdges);
                setTimeout(() => saveLayout(layouted.nodes), 1000);
            } else {
                setNodes(initialNodes);
                setEdges(initialEdges);
            }
        } catch (err) {
            console.error('Failed to fetch data:', err);
        } finally {
            setState((prev) => ({ ...prev, loading: false }));
        }
    }, [setNodes, setEdges, saveLayout, studio]);

    const fetchSVG = useCallback(async () => {
        if (!studio) return;
        try {
            const svg = await studio.fetchSVG();
            if (svg) {
                setState((prev) => ({ ...prev, svgCode: svg }));
            }
        } catch (err) {
            console.error('Failed to fetch SVG:', err);
        }
    }, [studio]);

    const updateGoCode = useCallback(async (val: string) => {
        if (!studio) return;
        setState((prev) => ({ ...prev, goCode: val }));
        isUpdating.current = true;
        try {
            await studio.updateGo(val);
        } finally {
            setTimeout(() => {
                isUpdating.current = false;
            }, 1000);
        }
    }, [studio]);

    const setJsonCode = useCallback((json: string) => {
        setState((prev) => ({ ...prev, jsonCode: json }));
    }, []);

    // Connect realtime on mount
    useEffect(() => {
        if (!studio) return;
        fetchData();
        const disconnect = studio.connectRealtime((message) => {
            if (message === 'refresh' || message === 'refresh-svg') {
                fetchData(true);
            }
        });
        return () => disconnect();
    }, [fetchData, studio]);

    return {
        nodes,
        edges,
        setNodes,
        setEdges,
        onNodesChange,
        onEdgesChange,
        ...state,
        saveLayout,
        fetchData,
        fetchSVG,
        updateGoCode,
        setJsonCode,
        isUpdating,
    };
}

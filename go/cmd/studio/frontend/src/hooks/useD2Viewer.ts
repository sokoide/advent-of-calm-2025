import { useState, useCallback } from 'react';

interface D2ViewerState {
    zoom: number;
    pan: { x: number; y: number };
    isPanning: boolean;
}

export function useD2Viewer() {
    const [state, setState] = useState<D2ViewerState>({
        zoom: 1,
        pan: { x: 0, y: 0 },
        isPanning: false,
    });

    const zoomIn = useCallback(() => {
        setState((prev) => ({
            ...prev,
            zoom: Math.min(2.5, Math.round((prev.zoom + 0.1) * 10) / 10),
        }));
    }, []);

    const zoomOut = useCallback(() => {
        setState((prev) => ({
            ...prev,
            zoom: Math.max(0.5, Math.round((prev.zoom - 0.1) * 10) / 10),
        }));
    }, []);

    const reset = useCallback(() => {
        setState({ zoom: 1, pan: { x: 0, y: 0 }, isPanning: false });
    }, []);

    const startPanning = useCallback(() => {
        setState((prev) => ({ ...prev, isPanning: true }));
    }, []);

    const stopPanning = useCallback(() => {
        setState((prev) => ({ ...prev, isPanning: false }));
    }, []);

    const movePan = useCallback((dx: number, dy: number) => {
        setState((prev) => {
            if (!prev.isPanning) return prev;
            return {
                ...prev,
                pan: { x: prev.pan.x + dx, y: prev.pan.y + dy },
            };
        });
    }, []);

    return {
        zoom: state.zoom,
        pan: state.pan,
        isPanning: state.isPanning,
        zoomIn,
        zoomOut,
        reset,
        startPanning,
        stopPanning,
        movePan,
    };
}

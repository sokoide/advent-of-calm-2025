import dagre from 'dagre';
import { type Node, type Edge, Position } from 'reactflow';

// Node size constants
const NODE_WIDTH = 200;
const NODE_HEIGHT = 80;
const CONTAINER_HEADER = 45;
const PADDING = 30;

interface LayoutNode extends Node {
  width?: number;
  height?: number;
}

/**
 * Hierarchical layout algorithm that:
 * 1. Processes containers from innermost to outermost (depth-first)
 * 2. Lays out nodes within each container group
 * 3. Then lays out containers and top-level nodes together
 * 4. Respects edge flow direction (top-to-bottom)
 */
export const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'TB') => {
  const nodeMap = new Map<string, LayoutNode>();
  const childrenMap = new Map<string, string[]>(); // parentId -> childIds
  const containerNodes: string[] = [];
  const nodeDepth = new Map<string, number>(); // node -> depth level

  // Initialize nodes and build hierarchy
  nodes.forEach((node) => {
    nodeMap.set(node.id, { ...node, width: NODE_WIDTH, height: NODE_HEIGHT });
    if (node.parentNode) {
      const siblings = childrenMap.get(node.parentNode) || [];
      childrenMap.set(node.parentNode, [...siblings, node.id]);
    }
  });

  // Calculate depth for each node (0 = top level)
  const calculateDepth = (nodeId: string): number => {
    if (nodeDepth.has(nodeId)) return nodeDepth.get(nodeId)!;
    const node = nodes.find(n => n.id === nodeId);
    if (!node?.parentNode) {
      nodeDepth.set(nodeId, 0);
      return 0;
    }
    const depth = calculateDepth(node.parentNode) + 1;
    nodeDepth.set(nodeId, depth);
    return depth;
  };

  nodes.forEach(node => calculateDepth(node.id));

  // Identify containers and sort by depth (deepest first)
  nodes.forEach((node) => {
    if (childrenMap.has(node.id)) {
      containerNodes.push(node.id);
    }
  });
  containerNodes.sort((a, b) => (nodeDepth.get(b) || 0) - (nodeDepth.get(a) || 0));

  // Get edges between specific nodes
  const getEdgesBetween = (nodeIds: string[]): Edge[] => {
    return edges.filter(e => nodeIds.includes(e.source) && nodeIds.includes(e.target));
  };

  // Layout a group of nodes
  const layoutNodes = (nodeIds: string[], relevantEdges: Edge[]): { width: number, height: number } => {
    if (nodeIds.length === 0) return { width: 0, height: 0 };

    const g = new dagre.graphlib.Graph();
    g.setGraph({
      rankdir: direction,
      nodesep: 60,
      ranksep: 80,
      align: 'DL',
      marginx: 0,
      marginy: 0
    });
    g.setDefaultEdgeLabel(() => ({}));

    // Add nodes to graph with their actual sizes
    nodeIds.forEach((id) => {
      const node = nodeMap.get(id);
      const width = node?.width || NODE_WIDTH;
      const height = node?.height || NODE_HEIGHT;
      g.setNode(id, { width, height });
    });

    // Add edges
    relevantEdges.forEach((edge) => {
      g.setEdge(edge.source, edge.target);
    });

    dagre.layout(g);

    // Calculate bounds and update positions
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;

    nodeIds.forEach((id) => {
      const n = g.node(id);
      if (!n) return;
      const left = n.x - n.width / 2;
      const top = n.y - n.height / 2;
      minX = Math.min(minX, left);
      minY = Math.min(minY, top);
      maxX = Math.max(maxX, left + n.width);
      maxY = Math.max(maxY, top + n.height);
    });

    // Normalize positions to start from (0, 0)
    nodeIds.forEach((id) => {
      const n = g.node(id);
      const node = nodeMap.get(id);
      if (node && n) {
        node.position = {
          x: (n.x - n.width / 2) - minX,
          y: (n.y - n.height / 2) - minY,
        };
        node.targetPosition = direction === 'TB' ? Position.Top : Position.Left;
        node.sourcePosition = direction === 'TB' ? Position.Bottom : Position.Right;
      }
    });

    return {
      width: maxX - minX,
      height: maxY - minY
    };
  };

  // STEP 1: Layout children inside each container (deepest containers first)
  containerNodes.forEach((containerId) => {
    const children = childrenMap.get(containerId) || [];
    if (children.length === 0) return;

    // Get edges between children of this container
    const childEdges = getEdgesBetween(children);

    // Layout children
    const size = layoutNodes(children, [...childEdges]);

    // Add padding and header to children positions
    children.forEach((childId) => {
      const child = nodeMap.get(childId);
      if (child) {
        child.position = {
          x: child.position.x + PADDING,
          y: child.position.y + PADDING + CONTAINER_HEADER,
        };
      }
    });

    // Update container size based on laid out children
    const container = nodeMap.get(containerId);
    if (container) {
      container.width = Math.max(size.width + PADDING * 2, NODE_WIDTH);
      container.height = Math.max(size.height + PADDING * 2 + CONTAINER_HEADER, NODE_HEIGHT);
      container.style = {
        ...container.style,
        width: container.width,
        height: container.height,
      };
    }
  });

  // STEP 2: Layout top-level nodes (including containers)
  const topLevelNodes: string[] = [];
  nodes.forEach((node) => {
    if (!node.parentNode) {
      topLevelNodes.push(node.id);
    }
  });

  // Create synthetic edges for layout
  // Map edges that involve container children to the container itself
  const topLevelEdges: Edge[] = [];
  const addedEdges = new Set<string>();

  // Helper to find the top-level ancestor
  const getTopLevelAncestor = (nodeId: string): string => {
    const node = nodes.find(n => n.id === nodeId);
    if (!node?.parentNode) return nodeId;
    return getTopLevelAncestor(node.parentNode);
  };

  edges.forEach((edge) => {
    const source = getTopLevelAncestor(edge.source);
    const target = getTopLevelAncestor(edge.target);

    // Only add if both are top-level and edge is new
    if (topLevelNodes.includes(source) && topLevelNodes.includes(target) && source !== target) {
      const edgeKey = `${source}->${target}`;
      if (!addedEdges.has(edgeKey)) {
        addedEdges.add(edgeKey);
        topLevelEdges.push({ ...edge, id: edgeKey, source, target });
      }
    }
  });

  layoutNodes(topLevelNodes, topLevelEdges);

  return {
    nodes: Array.from(nodeMap.values()),
    edges
  };
};

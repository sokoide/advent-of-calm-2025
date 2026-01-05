import dagre from 'dagre';
import { type Node, type Edge, Position } from 'reactflow';

// Node size constants
const NODE_WIDTH = 200;
const NODE_HEIGHT = 80;
const CONTAINER_HEADER = 50;
const PADDING = 40;

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


  // Layout a group of nodes
  const layoutNodes = (nodeIds: string[], relevantEdges: Edge[]): { width: number, height: number } => {
    if (nodeIds.length === 0) return { width: 0, height: 0 };

    const g = new dagre.graphlib.Graph();
    g.setGraph({
      rankdir: direction,

      nodesep: 100,
      ranksep: 150,
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

    // Add edges with weight and minlen to enforce hierarchy
    relevantEdges.forEach((edge) => {
      g.setEdge(edge.source, edge.target, { weight: 2, minlen: 1 });
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

  // Helper to find the direct child of a container that is an ancestor of the given node
  const getDirectChildAncestor = (nodeId: string, containerId: string | null): string | null => {
    let current = nodeId;
    while (current) {
      const node = nodes.find(n => n.id === current);
      if (!node) return null;
      if (node.parentNode === containerId) return current;
      if (!node.parentNode) return containerId === null ? current : null; // Top-level check
      current = node.parentNode;
    }
    return null;
  };

  // Helper function to layout a group of nodes (can be top-level or absolute container)
  const layoutGroup = (groupId: string | null, nodeIds: string[]): { width: number, height: number } => {
    if (nodeIds.length === 0) return { width: 0, height: 0 };

    // Find edges where both source and target trace back to nodes in this group
    const relevantEdges: Edge[] = [];
    const addedEdges = new Set<string>();

    edges.forEach((edge) => {
      const sourceAncestor = getDirectChildAncestor(edge.source, groupId);
      const targetAncestor = getDirectChildAncestor(edge.target, groupId);

      if (sourceAncestor && targetAncestor &&
        nodeIds.includes(sourceAncestor) && nodeIds.includes(targetAncestor) &&
        sourceAncestor !== targetAncestor) {

        const edgeKey = `${sourceAncestor}->${targetAncestor}`;
        if (!addedEdges.has(edgeKey)) {
          addedEdges.add(edgeKey);
          // Use a higher weight for these "structural" edges
          relevantEdges.push({ ...edge, id: edgeKey, source: sourceAncestor, target: targetAncestor });
        }
      }
    });

    return layoutNodes(nodeIds, relevantEdges);
  };

  // STEP 1: Layout children inside each container (deepest containers first)
  containerNodes.forEach((containerId) => {
    const children = childrenMap.get(containerId) || [];
    if (children.length === 0) return;

    // Layout children with recursive edge lifting
    const size = layoutGroup(containerId, children);

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

    // Update container size
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

  // STEP 2: Layout top-level nodes
  const topLevelNodes: string[] = [];
  nodes.forEach((node) => {
    if (!node.parentNode) {
      topLevelNodes.push(node.id);
    }
  });

  layoutGroup(null, topLevelNodes);

  return {
    nodes: Array.from(nodeMap.values()),
    edges
  };
};

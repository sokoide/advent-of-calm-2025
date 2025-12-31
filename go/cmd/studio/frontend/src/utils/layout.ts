import dagre from 'dagre';
import { type Node, type Edge, Position } from 'reactflow';

// Node size constants
const NODE_WIDTH = 200;
const NODE_HEIGHT = 80;
const PADDING = 40;

interface LayoutNode extends Node {
  width?: number;
  height?: number;
}

export const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
  // 1. Build Hierarchy Tree
  const nodeMap = new Map<string, LayoutNode>();
  const childrenMap = new Map<string, string[]>(); // parentId -> childIds
  const topLevelNodes: string[] = [];

  nodes.forEach((node) => {
    nodeMap.set(node.id, { ...node, width: NODE_WIDTH, height: NODE_HEIGHT });
    if (node.parentNode) {
      const siblings = childrenMap.get(node.parentNode) || [];
      childrenMap.set(node.parentNode, [...siblings, node.id]);
    } else {
      topLevelNodes.push(node.id);
    }
  });

  // 2. Recursive Layout Function
  // Returns the size (width, height) of the laid-out group
  const layoutGroup = (nodeIds: string[], rankDir: string): { width: number, height: number, offset: { x: number, y: number } } => {
    if (nodeIds.length === 0) return { width: 0, height: 0, offset: { x: 0, y: 0 } };

    const g = new dagre.graphlib.Graph();
    g.setGraph({ rankdir: rankDir, nodesep: 50, ranksep: 50 });
    g.setDefaultEdgeLabel(() => ({}));

    nodeIds.forEach((id) => {
      // Recursively layout children if this node is a container
      const children = childrenMap.get(id);
      let width = NODE_WIDTH;
      let height = NODE_HEIGHT;

      if (children && children.length > 0) {
        const size = layoutGroup(children, rankDir);
        // Resize container to fit children + padding
        width = size.width + PADDING * 2;
        height = size.height + PADDING * 2;
        // Store calculated size in the node map for the next level up
        const node = nodeMap.get(id);
        if (node) {
          node.width = width;
          node.height = height;
          
          // Apply style update for container size
          if (node.style) {
             node.style = { ...node.style, width, height };
          }
        }
      }

      g.setNode(id, { width, height });
    });

    // Add edges that are relevant to this group
    edges.forEach((edge) => {
      const sourceInGroup = nodeIds.includes(edge.source);
      const targetInGroup = nodeIds.includes(edge.target);
      if (sourceInGroup && targetInGroup) {
        g.setEdge(edge.source, edge.target);
      }
    });

    dagre.layout(g);

    // Update positions in nodeMap
    // Calculate bounding box to normalize positions (make top-left 0,0)
    let minX = Infinity, minY = Infinity;
    
    nodeIds.forEach((id) => {
      const n = g.node(id);
      minX = Math.min(minX, n.x - n.width / 2);
      minY = Math.min(minY, n.y - n.height / 2);
    });

    let maxWidth = 0;
    let maxHeight = 0;

    nodeIds.forEach((id) => {
      const n = g.node(id);
      const node = nodeMap.get(id);
      if (node) {
        // Dagre gives center coordinates, convert to top-left relative to group
        // If it's a top level layout, we don't need to offset by minX/minY strictly, 
        // but for nested groups, we want them relative to the container's 0,0 (plus padding).
        
        node.position = {
          x: (n.x - n.width / 2) - minX + PADDING, // Add padding for inner content
          y: (n.y - n.height / 2) - minY + PADDING,
        };
        
        // React Flow handles handles automatically based on position and size
        node.targetPosition = direction === 'LR' ? Position.Left : Position.Top;
        node.sourcePosition = direction === 'LR' ? Position.Right : Position.Bottom;

        maxWidth = Math.max(maxWidth, node.position.x + n.width);
        maxHeight = Math.max(maxHeight, node.position.y + n.height);
      }
    });

    return { width: maxWidth + PADDING, height: maxHeight + PADDING, offset: { x: minX, y: minY } };
  };

  // 3. Execute Layout from Top Level
  // This will recursively recurse down to leaves, then bubble up sizes
  layoutGroup(topLevelNodes, direction);

  // 4. Return updated nodes
  // We don't need to update edges as React Flow handles them based on handles
  return { 
    nodes: Array.from(nodeMap.values()), 
    edges 
  };
};

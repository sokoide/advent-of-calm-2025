import { type Node, type Edge } from 'reactflow';
import type { CalmArchitecture, LayoutData } from '../types/calm';

export const transformToReactFlow = (
  arch: CalmArchitecture,
  layout: LayoutData
): { nodes: Node[]; edges: Edge[] } => {
  // 1. Build Parent-Child Maps to determine parentNode
  const nodeToParent = new Map<string, string>();
  const allContainers = new Set<string>();

  arch.relationships.forEach((rel) => {
    const relType = rel['relationship-type'];
    if (relType['composed-of']) {
      const container = relType['composed-of'].container;
      const children = relType['composed-of'].nodes;
      allContainers.add(container);
      children.forEach((childId) => {
        // Deepest parent wins
        nodeToParent.set(childId, container);
      });
    }
  });

  // 2. Generate Nodes
  const nodes: Node[] = arch.nodes.map((n) => {
    const id = n['unique-id'];
    const parentId = nodeToParent.get(id);
    let pos = layout.nodes[id] || { x: 0, y: 0 };

    // Convert absolute to relative if it has a parent
    if (parentId && layout.nodes[parentId]) {
      const parentPos = layout.nodes[parentId];
      pos = {
        x: pos.x - parentPos.x,
        y: pos.y - parentPos.y,
      };
    }

    const isContainer = allContainers.has(id);

    return {
      id,
      type: n['node-type'],
      data: { 
        label: n.name,
        calm: n,
        isContainer
      },
      position: pos,
      parentNode: parentId,
      zIndex: isContainer ? -1 : 1,
      // Default styles for containers - will be overridden by layout engine sizes
      style: isContainer ? {
        width: 400,
        height: 250,
        backgroundColor: 'rgba(15, 23, 42, 0.1)',
        border: '2px dashed rgba(100, 116, 139, 0.5)',
        borderRadius: '12px',
      } : undefined,
    };
  });

  // 3. Generate Edges
  const edges: Edge[] = [];
  arch.relationships.forEach((rel) => {
    const relType = rel['relationship-type'];
    if (relType.connects) {
      edges.push({
        id: rel['unique-id'],
        source: relType.connects.source.node,
        target: relType.connects.destination.node,
        label: rel.description,
        animated: true,
      });
    }
    if (relType.interacts) {
      const actor = relType.interacts.actor;
      relType.interacts.nodes.forEach((targetNode, idx) => {
        edges.push({
          id: `${rel['unique-id']}-${idx}`,
          source: actor,
          target: targetNode,
          label: rel.description,
        });
      });
    }
  });

  return { nodes, edges };
};
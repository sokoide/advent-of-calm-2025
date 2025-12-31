import { type Node, type Edge } from 'reactflow';
import type { CalmArchitecture, LayoutData } from '../types/calm';

export const transformToReactFlow = (
  arch: CalmArchitecture,
  layout: LayoutData
): { nodes: Node[]; edges: Edge[] } => {
  const nodes: Node[] = arch.nodes.map((n) => {
    const id = n['unique-id'];
    const pos = layout.nodes[id] || { x: 0, y: 0 };

    return {
      id,
      type: n['node-type'],
      data: { 
        label: n.name,
        calm: n 
      },
      position: pos,
    };
  });

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

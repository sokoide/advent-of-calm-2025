import { type Node, type Edge } from 'reactflow';
import type { CalmArchitecture, LayoutData } from '../types/calm';

export const transformToReactFlow = (
  arch: CalmArchitecture,
  layout: LayoutData
): { nodes: Node[]; edges: Edge[] } => {
  // 1. Build Parent-Child Maps to determine parentNode
  const nodeToParent = new Map<string, string>();
  const parentToChildren = new Map<string, string[]>();
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
        const siblings = parentToChildren.get(container) || [];
        parentToChildren.set(container, [...siblings, childId]);
      });
    }
  });

  // 2. Generate Nodes
  const hasParentMap = layout.parentMap && Object.keys(layout.parentMap).length > 0;

  const nodes: Node[] = arch.nodes.map((n) => {
    const id = n['unique-id'];
    const parentId = nodeToParent.get(id);
    let pos = layout.nodes[id] || { x: 0, y: 0 };

    // Backward compatibility: if layout doesn't include parentMap, treat saved positions as absolute.
    if (parentId && !hasParentMap && layout.nodes[parentId]) {
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
      extent: parentId ? 'parent' : undefined,
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

  // 2.5. Resize containers to fit children based on relative positions.
  const nodeMap = new Map(nodes.map((node) => [node.id, node]));
  const sizeByType = (type: string) => {
    switch (type) {
      case 'actor':
        return { width: 150, height: 60 };
      case 'system':
        return { width: 300, height: 200 };
      default:
        return { width: 200, height: 80 };
    }
  };

  const getSize = (id: string): { width: number; height: number } => {
    const node = nodeMap.get(id);
    if (!node) return { width: 200, height: 80 };
    if (!allContainers.has(id)) {
      return sizeByType(node.type as string);
    }
    const children = parentToChildren.get(id) || [];
    if (children.length === 0) {
      return { width: 400, height: 250 };
    }
    let maxX = 0;
    let maxY = 0;
    children.forEach((childId) => {
      const child = nodeMap.get(childId);
      if (!child) return;
      const childSize = getSize(childId);
      maxX = Math.max(maxX, child.position.x + childSize.width);
      maxY = Math.max(maxY, child.position.y + childSize.height);
    });
    const padding = 40;
    return {
      width: Math.max(300, maxX + padding),
      height: Math.max(200, maxY + padding),
    };
  };

  allContainers.forEach((id) => {
    const node = nodeMap.get(id);
    if (!node || !node.style) return;
    const size = getSize(id);
    node.style = { ...node.style, width: size.width, height: size.height };
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

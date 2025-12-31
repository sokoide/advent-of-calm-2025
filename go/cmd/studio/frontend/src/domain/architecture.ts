import type { CalmArchitecture } from './calm';

export const buildParentMap = (arch: CalmArchitecture): Record<string, string> => {
  const map: Record<string, string> = {};
  arch.relationships.forEach((rel) => {
    const composed = rel['relationship-type']?.['composed-of'];
    if (!composed) return;
    const container = composed.container;
    composed.nodes.forEach((childId) => {
      map[childId] = container;
    });
  });
  return map;
};

export const parentMapEquals = (
  previous: Record<string, string> | undefined,
  current: Record<string, string>
): boolean => {
  const prev = previous || {};
  const prevKeys = Object.keys(prev);
  const curKeys = Object.keys(current);
  if (prevKeys.length !== curKeys.length) return false;
  return prevKeys.every((key) => prev[key] === current[key]);
};

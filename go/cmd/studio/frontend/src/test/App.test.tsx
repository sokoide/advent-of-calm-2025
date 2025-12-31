import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import App from '../App'

// Mock axios
vi.mock('axios', () => ({
  default: {
    get: vi.fn((url) => {
      if (url.includes('/content')) {
        return Promise.resolve({ 
          data: { 
            goCode: 'package main', 
            d2Code: 'x -> y', 
            svg: '<svg>test-svg</svg>',
            json: JSON.stringify({ 'unique-id': 'test', nodes: [], relationships: [] }) 
          } 
        });
      }
      if (url.includes('/layout')) {
        return Promise.resolve({ data: { nodes: {} } });
      }
      return Promise.resolve({ data: {} });
    }),
    post: vi.fn(() => Promise.resolve({ data: {} })),
  },
}))

// Mock ReactFlow
vi.mock('reactflow', () => ({
  useNodesState: () => [[], vi.fn(), vi.fn()],
  useEdgesState: () => [[], vi.fn(), vi.fn()],
  addEdge: vi.fn(),
  Background: () => <div data-testid="rf-background" />,
  Controls: () => <div data-testid="rf-controls" />,
  Panel: ({ children }: any) => <div>{children}</div>,
  MarkerType: { ArrowClosed: 'arrowclosed' },
  default: ({ children }: any) => <div>{children}</div>,
}))

// Mock react-resizable-panels
vi.mock('react-resizable-panels', () => ({
  Group: ({ children }: any) => <div data-testid="resizable-group">{children}</div>,
  Panel: ({ children }: any) => <div data-testid="resizable-panel">{children}</div>,
  Separator: () => <div data-testid="resizable-separator" />,
}))

describe('App Tab Navigation', () => {
  it('should have 6 tabs in the correct order and load Go DSL', async () => {
    render(<App />)
    
    // Wait for loading to finish
    await waitFor(() => {
      expect(screen.queryByText(/Loading Studio/i)).not.toBeInTheDocument()
    })
    
    const tabs = screen.getAllByRole('button').filter(b => 
      ['Merged', 'Diagram', 'Go DSL', 'CALM JSON', 'D2 Diagram', 'D2 DSL'].includes(b.textContent?.trim() || '')
    )
    
    const tabLabels = tabs.map(t => t.textContent?.trim())
    
    expect(tabLabels).toEqual([
      'Merged',
      'Diagram',
      'Go DSL',
      'CALM JSON',
      'D2 Diagram',
      'D2 DSL'
    ])

    // Verify "Merged" tab is active and shows the Go DSL editor
    expect(screen.getByText(/Go DSL Editor/i)).toBeInTheDocument()
  })
})

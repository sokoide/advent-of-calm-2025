import { Monitor, Code2, FileJson, Columns, Layers, FileCode } from 'lucide-react';

export type TabType = 'merged' | 'diagram' | 'go' | 'json' | 'd2-diagram' | 'd2-dsl';

const tabs = [
    { id: 'merged', label: 'Merged', icon: Columns },
    { id: 'diagram', label: 'Diagram', icon: Monitor },
    { id: 'go', label: 'Go DSL', icon: Code2 },
    { id: 'json', label: 'CALM JSON', icon: FileJson },
    { id: 'd2-diagram', label: 'D2 Diagram', icon: Layers },
    { id: 'd2-dsl', label: 'D2 DSL', icon: FileCode },
] as const;

interface TabNavProps {
    activeTab: TabType;
    onTabChange: (tab: TabType) => void;
}

const TabNav = ({ activeTab, onTabChange }: TabNavProps) => {
    return (
        <nav className="flex bg-slate-800/50 rounded-lg p-1 border border-slate-700">
            {tabs.map((t) => (
                <button
                    key={t.id}
                    onClick={() => onTabChange(t.id as TabType)}
                    className={`flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-medium transition-all ${activeTab === t.id
                            ? 'bg-blue-600 text-white shadow-lg'
                            : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700'
                        }`}
                >
                    <t.icon size={14} /> {t.label}
                </button>
            ))}
        </nav>
    );
};

export default TabNav;

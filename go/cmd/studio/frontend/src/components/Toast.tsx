import { useState, useCallback, useEffect } from 'react';
import { CheckCircle2, XCircle, AlertTriangle, X } from 'lucide-react';

type ToastType = 'success' | 'error' | 'warning';

interface Toast {
    id: number;
    message: string;
    type: ToastType;
}

let toastId = 0;

// Global toast function - will be set by ToastProvider
let globalShowToast: (message: string, type: ToastType) => void = () => { };

export function showToast(message: string, type: ToastType = 'success') {
    globalShowToast(message, type);
}

export function ToastProvider({ children }: { children: React.ReactNode }) {
    const [toasts, setToasts] = useState<Toast[]>([]);

    const addToast = useCallback((message: string, type: ToastType) => {
        const id = ++toastId;
        setToasts((prev) => [...prev, { id, message, type }]);
        setTimeout(() => {
            setToasts((prev) => prev.filter((t) => t.id !== id));
        }, 4000);
    }, []);

    const removeToast = useCallback((id: number) => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
    }, []);

    useEffect(() => {
        globalShowToast = addToast;
    }, [addToast]);

    const icons = {
        success: <CheckCircle2 size={18} className="text-green-400" />,
        error: <XCircle size={18} className="text-red-400" />,
        warning: <AlertTriangle size={18} className="text-yellow-400" />,
    };

    const bgColors = {
        success: 'bg-green-900/90 border-green-700',
        error: 'bg-red-900/90 border-red-700',
        warning: 'bg-yellow-900/90 border-yellow-700',
    };

    return (
        <>
            {children}
            <div className="fixed bottom-4 right-4 z-[200] flex flex-col gap-2">
                {toasts.map((toast) => (
                    <div
                        key={toast.id}
                        className={`flex items-center gap-3 px-4 py-3 rounded-lg border shadow-xl text-sm text-white animate-in slide-in-from-right duration-200 ${bgColors[toast.type]}`}
                    >
                        {icons[toast.type]}
                        <span>{toast.message}</span>
                        <button
                            onClick={() => removeToast(toast.id)}
                            className="ml-2 text-slate-400 hover:text-white"
                        >
                            <X size={14} />
                        </button>
                    </div>
                ))}
            </div>
        </>
    );
}

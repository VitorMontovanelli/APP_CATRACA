import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import "./Toast.css";

export type ToastType = "success" | "warning" | "error";

export interface ToastOptions {
  type?: ToastType;
  title: string;
  description?: string;
  duration?: number;
}

interface ToastItem {
  id: number;
  type: ToastType;
  title: string;
  description?: string;
  duration: number;
  leaving: boolean;
}

export interface NotificationItem {
  id: number;
  type: ToastType;
  title: string;
  description?: string;
  timestamp: number;
}

interface ToastContextValue {
  showToast: (options: ToastOptions) => void;
  history: NotificationItem[];
  activeCount: number;
  removeNotification: (id: number) => void;
  clearNotifications: () => void;
}

const DEFAULT_DURATION = 3000;
const EXIT_MS = 300;
const MAX_HISTORY = 30;

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

export function CheckIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <circle cx="12" cy="12" r="10" />
      <path d="m9 12 2 2 4-4" />
    </svg>
  );
}

export function WarningIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3" />
      <path d="M12 9v4" />
      <path d="M12 17h.01" />
    </svg>
  );
}

export function FingerprintIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M12 10a2 2 0 0 0-2 2c0 1.02-.1 2.51-.26 4" />
      <path d="M14 13.12c0 2.38 0 6.38-1 8.88" />
      <path d="M17.29 21.02c.12-.6.43-2.3.5-3.02" />
      <path d="M2 12a10 10 0 0 1 18-6" />
      <path d="M2 16h.01" />
      <path d="M21.8 16c.2-2 .131-5.354 0-6" />
      <path d="M5 19.5C5.5 18 6 15 6 12a6 6 0 0 1 .34-2" />
      <path d="M8.65 22c.21-.66.45-1.32.57-2" />
      <path d="M9 6.8a6 6 0 0 1 9 5.2v2" />
    </svg>
  );
}

function ToastCard({ toast, onClose }: { toast: ToastItem; onClose: (id: number) => void }) {
  const [remaining, setRemaining] = useState(toast.duration);
  const [paused, setPaused] = useState(false);

  useEffect(() => {
    if (paused) return;
    const timer = setTimeout(() => {
      setRemaining((prev) => Math.max(0, prev - 100));
    }, 100);
    return () => clearTimeout(timer);
  }, [paused, remaining]);

  useEffect(() => {
    if (remaining <= 0) onClose(toast.id);
  }, [remaining, onClose, toast.id]);

  const pct = Math.max(0, Math.min(100, (remaining / toast.duration) * 100));

  return (
    <div
      className={`toast toast--${toast.type}${toast.leaving ? " leaving" : ""}`}
      role="status"
      onClick={() => onClose(toast.id)}
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
    >
      <div className="toast-icon">
        {toast.type === "success" && <CheckIcon />}
        {toast.type === "warning" && <WarningIcon />}
        {toast.type === "error" && <FingerprintIcon />}
      </div>

      <div className="toast-content">
        <div className="toast-header">
          <span className="toast-title">{toast.title}</span>
          <button
            className="toast-close"
            onClick={() => onClose(toast.id)}
            aria-label="Fechar notificação"
          >
            ×
          </button>
        </div>
        {toast.description && (
          <div className="toast-description">{toast.description}</div>
        )}
      </div>

      <div className="toast-progress-track">
        <div className="toast-progress-fill" style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const [history, setHistory] = useState<NotificationItem[]>([]);
  const nextId = useRef(1);

  const showToast = useCallback((options: ToastOptions) => {
    const { type = "success", title, description, duration = DEFAULT_DURATION } = options;
    const id = nextId.current++;
    setToasts((prev) => [
      { id, type, title, description, duration, leaving: false },
      ...prev,
    ]);
    setHistory((prev) =>
      [{ id, type, title, description, timestamp: Date.now() }, ...prev].slice(0, MAX_HISTORY),
    );
  }, []);

  const dismiss = useCallback((id: number) => {
    setToasts((prev) =>
      prev.map((t) => (t.id === id ? { ...t, leaving: true } : t)),
    );
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, EXIT_MS);
  }, []);

  const removeNotification = useCallback((id: number) => {
    setHistory((prev) => prev.filter((n) => n.id !== id));
  }, []);

  const clearNotifications = useCallback(() => {
    setHistory([]);
    setToasts((prev) => prev.map((t) => ({ ...t, leaving: true })));
    window.setTimeout(() => setToasts([]), EXIT_MS);
  }, []);

  const activeCount = useMemo(
    () => toasts.filter((t) => !t.leaving).length,
    [toasts],
  );

  const value = useMemo(
    () => ({ showToast, history, activeCount, removeNotification, clearNotifications }),
    [showToast, history, activeCount, removeNotification, clearNotifications],
  );

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="toast-container">
        {toasts.map((toast) => (
          <ToastCard key={toast.id} toast={toast} onClose={dismiss} />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error("useToast deve ser usado dentro de <ToastProvider>");
  }
  return ctx;
}

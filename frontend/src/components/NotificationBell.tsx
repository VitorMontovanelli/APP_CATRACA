import { useEffect, useRef, useState } from "react";
import {
  useToast,
  CheckIcon,
  WarningIcon,
  FingerprintIcon,
  type ToastType,
} from "./ToastProvider";

function BellIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
      <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" />
    </svg>
  );
}

function ItemIcon({ type }: { type: ToastType }) {
  if (type === "success") return <CheckIcon />;
  if (type === "warning") return <WarningIcon />;
  return <FingerprintIcon />;
}

function formatTime(ts: number): string {
  return new Date(ts).toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" });
}

function NotificationBell() {
  const { history, activeCount, removeNotification, clearNotifications } = useToast();
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    function onMouseDown(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }

    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }

    document.addEventListener("mousedown", onMouseDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("mousedown", onMouseDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  return (
    <div className="toast-bell-wrap" ref={rootRef}>
      <button
        className="toast-bell"
        onClick={() => setOpen((o) => !o)}
        aria-label="Notificações"
        aria-expanded={open}
      >
        <BellIcon />
        {activeCount > 0 && <span className="toast-bell-badge">{activeCount}</span>}
      </button>

      {open && (
        <div className="toast-panel">
          <div className="toast-panel-header">
            <span className="toast-panel-title">Notificações</span>
            {history.length > 0 && (
              <button className="toast-panel-clear" onClick={clearNotifications}>
                Limpar tudo
              </button>
            )}
          </div>

          {history.length === 0 ? (
            <div className="toast-panel-empty">Nenhuma notificação</div>
          ) : (
            history.map((n) => (
              <div key={n.id} className={`toast-panel-item toast-panel-item--${n.type}`}>
                <div className="toast-icon">
                  <ItemIcon type={n.type} />
                </div>
                <div className="toast-content">
                  <div className="toast-panel-item-title">{n.title}</div>
                  {n.description && (
                    <div className="toast-panel-item-desc">{n.description}</div>
                  )}
                  <div className="toast-panel-item-time">{formatTime(n.timestamp)}</div>
                </div>
                <button
                  className="toast-panel-item-remove"
                  onClick={() => removeNotification(n.id)}
                  aria-label="Remover notificação"
                >
                  ×
                </button>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}

export default NotificationBell;

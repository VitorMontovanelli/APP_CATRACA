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
import {
  ListarFaturasVencendo,
  MarcarAvisoComoLido,
  MarcarTodosAvisosComoLidos,
  VerificarAvisosVencimento,
} from "../../wailsjs/go/main/App";
import { useToast } from "./ToastProvider";

export interface DueNotification {
  id: number;
  invoice_id: number;
  kind: "due" | "overdue" | string;
  student_id: number;
  student_name: string;
  amount_cents: number;
  due_date: string;
  message: string;
  read_at?: string | null;
  created_at: string;
}

export interface FaturaVencendo {
  invoice_id: number;
  student_id: number;
  student_name: string;
  amount_cents: number;
  due_date: string;
  dias_restantes: number;
}

interface DueNotificationsContextValue {
  notifications: DueNotification[];
  unreadCount: number;
  upcoming: FaturaVencendo[];
  loading: boolean;
  refresh: () => Promise<void>;
  markRead: (id: number) => Promise<void>;
  markAllRead: () => Promise<void>;
}

const DueNotificationsContext = createContext<DueNotificationsContextValue | undefined>(undefined);

const UMA_HORA_MS = 60 * 60 * 1000;

export function DueNotificationsProvider({ children }: { children: ReactNode }) {
  const { showToast } = useToast();
  const [notifications, setNotifications] = useState<DueNotification[]>([]);
  const [upcoming, setUpcoming] = useState<FaturaVencendo[]>([]);
  const [loading, setLoading] = useState(true);
  const knownIds = useRef<Set<number> | null>(null);

  const refresh = useCallback(async () => {
    try {
      const [lista, faturas] = await Promise.all([
        VerificarAvisosVencimento(),
        ListarFaturasVencendo(),
      ]);
      const ativos = (lista as unknown as DueNotification[]) || [];
      setNotifications(ativos);
      setUpcoming((faturas as unknown as FaturaVencendo[]) || []);

      const naoLidos = ativos.filter((n) => !n.read_at);
      if (knownIds.current === null) {
        knownIds.current = new Set(ativos.map((n) => n.id));
        if (naoLidos.length > 0) {
          showToast({
            type: "warning",
            title: naoLidos.length === 1 ? "Aviso de vencimento" : `${naoLidos.length} avisos de vencimento`,
            description: naoLidos[0].message,
            duration: 8000,
          });
        }
      } else {
        const novos = naoLidos.filter((n) => !knownIds.current!.has(n.id));
        for (const n of ativos) knownIds.current.add(n.id);
        if (novos.length > 0) {
          showToast({
            type: "warning",
            title: novos.length === 1 ? "Novo aviso de vencimento" : `${novos.length} novos avisos de vencimento`,
            description: novos[0].message,
            duration: 8000,
          });
        }
      }
    } catch (e) {
      console.error("Erro ao verificar avisos de vencimento", e);
    } finally {
      setLoading(false);
    }
  }, [showToast]);

  useEffect(() => {
    refresh();
    const timer = window.setInterval(refresh, UMA_HORA_MS);
    return () => window.clearInterval(timer);
  }, [refresh]);

  const markRead = useCallback(async (id: number) => {
    try {
      await MarcarAvisoComoLido(id);
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, read_at: new Date().toISOString() } : n)),
      );
    } catch (e) {
      console.error("Erro ao marcar aviso como lido", e);
    }
  }, []);

  const markAllRead = useCallback(async () => {
    try {
      await MarcarTodosAvisosComoLidos();
      setNotifications((prev) =>
        prev.map((n) => (n.read_at ? n : { ...n, read_at: new Date().toISOString() })),
      );
    } catch (e) {
      console.error("Erro ao marcar avisos como lidos", e);
    }
  }, []);

  const unreadCount = notifications.filter((n) => !n.read_at).length;

  const value = useMemo(
    () => ({ notifications, unreadCount, upcoming, loading, refresh, markRead, markAllRead }),
    [notifications, unreadCount, upcoming, loading, refresh, markRead, markAllRead],
  );

  return (
    <DueNotificationsContext.Provider value={value}>
      {children}
    </DueNotificationsContext.Provider>
  );
}

export function useDueNotifications(): DueNotificationsContextValue {
  const ctx = useContext(DueNotificationsContext);
  if (!ctx) {
    throw new Error("useDueNotifications deve ser usado dentro de <DueNotificationsProvider>");
  }
  return ctx;
}

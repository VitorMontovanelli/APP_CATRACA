import { useState } from "react";
import { SalvarPermissoesUsuario } from "../../../wailsjs/go/main/App";

interface User {
  id: number;
  name: string;
  permissoes?: string;
}

interface PermissoesModalProps {
  user: User;
  onClose: () => void;
  onSaved: () => void;
}

const AVAILABLE_PAGES = [
  { key: "home", label: "Visão Geral", icon: "📊" },
  { key: "agenda", label: "Calendários de Agendamento", icon: "📅" },
  { key: "alunos", label: "Gestão de Alunos", icon: "🎓" },
  { key: "planos", label: "Planos de Acesso", icon: "📋" },
  { key: "financeiro", label: "Visão Financeira", icon: "💰" },
  { key: "cobranca", label: "Cobranças", icon: "⚠️" },
  { key: "metodos_pagamento", label: "Pagamentos", icon: "💳" },
  { key: "usuarios", label: "Usuários Administrativos", icon: "👥" },
  { key: "logs", label: "Logs de Acesso", icon: "📋" },
];

export default function PermissoesModal({ user, onClose, onSaved }: PermissoesModalProps) {
  const initialPerms = user.permissoes ? user.permissoes.split(",") : ["home", "alunos", "planos", "financeiro", "cobranca", "metodos_pagamento", "usuarios", "logs"];
  const [selected, setSelected] = useState<string[]>(initialPerms);
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState("");

  function toggle(key: string) {
    if (selected.includes(key)) {
      setSelected(selected.filter((k) => k !== key));
    } else {
      setSelected([...selected, key]);
    }
  }

  async function handleSave() {
    setLoading(true);
    setMsg("");
    try {
      await SalvarPermissoesUsuario(user.id, selected.join(","));
      setMsg("Permissões de acesso salvas com sucesso!");
      setTimeout(() => {
        onSaved();
        onClose();
      }, 1000);
    } catch (e) {
      setMsg(`Erro: ${String(e)}`);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="vmd-modal-overlay" onClick={onClose}>
      <div className="vmd-modal-content" style={{ maxWidth: 460 }} onClick={(e) => e.stopPropagation()}>
        <div className="vmd-modal-header">
          <h3 style={{ fontSize: 16, margin: 0 }}>Acessos & Permissões: {user.name}</h3>
          <button type="button" className="vmd-modal-close" onClick={onClose}>✕</button>
        </div>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Selecione os módulos e telas que este administrador poderá acessar.
        </p>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10, margin: "16px 0" }}>
          {AVAILABLE_PAGES.map((p) => {
            const checked = selected.includes(p.key);
            return (
              <label
                key={p.key}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 8,
                  padding: "8px 10px",
                  borderRadius: 6,
                  border: `1px solid ${checked ? "var(--primary)" : "var(--border-color)"}`,
                  background: checked ? "rgba(199, 157, 51, 0.1)" : "transparent",
                  cursor: "pointer",
                  fontSize: 13,
                }}
              >
                <input
                  type="checkbox"
                  checked={checked}
                  onChange={() => toggle(p.key)}
                />
                <span>{p.icon} {p.label}</span>
              </label>
            );
          })}
        </div>

        {msg && <div style={{ fontSize: 13, color: "var(--primary)", marginBottom: 12 }}>{msg}</div>}

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <button type="button" className="vmd-btn vmd-btn-secondary" onClick={onClose}>
            Cancelar
          </button>
          <button type="button" className="vmd-btn vmd-btn-primary" onClick={handleSave} disabled={loading}>
            {loading ? "Salvando..." : "Salvar Acessos"}
          </button>
        </div>
      </div>
    </div>
  );
}

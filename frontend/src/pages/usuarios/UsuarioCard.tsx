import { useState } from "react";
import { AtualizarUsuario } from "../../../wailsjs/go/main/App";

export interface UserItem {
  id: number;
  name: string;
  email: string;
  cargo: string;
  ativo: boolean;
  permissoes?: string;
  criado_em?: string;
}

interface UsuarioCardProps {
  item: UserItem;
  actorId: number;
  actorCargo: string;
  allowedCargos: string[];
  onReload: () => void;
  onToggleAtivo: (u: UserItem) => void;
  onDelete: (u: UserItem) => void;
  onOpenPermissoes: (u: UserItem) => void;
  onOpenSenha: (u: UserItem) => void;
}

export default function UsuarioCard({
  item,
  actorId,
  actorCargo,
  allowedCargos,
  onReload,
  onToggleAtivo,
  onDelete,
  onOpenPermissoes,
  onOpenSenha,
}: UsuarioCardProps) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(item.name);
  const [email, setEmail] = useState(item.email);
  const [cargo, setCargo] = useState(item.cargo);
  const [saving, setSaving] = useState(false);

  const isSuper = actorCargo === "super_admin";
  const isSelf = item.id === actorId;
  // Administrador comum só pode alterar a própria senha. Editar, desativar,
  // excluir e promover usuários é exclusivo do super admin.
  const podeAlterarSenha = isSuper || isSelf;
  const badgeClass = item.cargo === "super_admin" ? "vmd-badge vmd-badge-danger" : "vmd-badge vmd-badge-warning";
  const cargoLabel = item.cargo === "super_admin" ? "Super Admin" : "Administrador";

  async function handleSave() {
    setSaving(true);
    try {
      await AtualizarUsuario(item.id, name, email, cargo);
      setEditing(false);
      onReload();
    } catch (e) {
      alert(String(e));
    } finally {
      setSaving(false);
    }
  }

  if (editing) {
    return (
      <div className="vmd-card" style={{ padding: "16px 20px" }}>
        <div style={{ display: "flex", gap: 12, flexWrap: "wrap", alignItems: "flex-end" }}>
          <div style={{ flex: 1, minWidth: 140 }}>
            <label className="vmd-label">Nome</label>
            <input className="vmd-input" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div style={{ flex: 1, minWidth: 160 }}>
            <label className="vmd-label">Usuário / E-mail</label>
            <input className="vmd-input" value={email} onChange={(e) => setEmail(e.target.value)} />
          </div>
          <div style={{ width: 130 }}>
            <label className="vmd-label">Cargo</label>
            <select className="vmd-select" value={cargo} onChange={(e) => setCargo(e.target.value)}>
              {allowedCargos.map((c) => <option key={c} value={c}>Admin</option>)}
            </select>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <button className="vmd-btn vmd-btn-success" onClick={handleSave} disabled={saving}>Salvar</button>
            <button className="vmd-btn vmd-btn-secondary" onClick={() => setEditing(false)}>Cancelar</button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div
      className="vmd-card"
      style={{
        padding: "16px 20px",
        opacity: item.ativo ? 1 : 0.6,
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        flexWrap: "wrap",
        gap: 12,
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        <span className={badgeClass}>{cargoLabel}</span>
        <div>
          <div style={{ fontSize: 15, fontWeight: 600, display: "flex", alignItems: "center", gap: 8 }}>
            {item.name}
            {!item.ativo && <span className="vmd-badge vmd-badge-danger">Inativo</span>}
          </div>
          <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>{item.email}</div>
        </div>
      </div>

      <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
        {isSuper && item.cargo === "admin" && (
          <button className="vmd-btn vmd-btn-secondary" style={{ padding: "6px 10px", fontSize: 12 }} onClick={() => onOpenPermissoes(item)}>
            🔑 Acessos
          </button>
        )}
        {(isSuper || item.cargo !== "super_admin") && podeAlterarSenha && (
          <button className="vmd-btn vmd-btn-secondary" style={{ padding: "6px 10px", fontSize: 12 }} onClick={() => onOpenSenha(item)}>
            🔒 Senha
          </button>
        )}
        {isSuper && (
          <>
            <button className="vmd-btn vmd-btn-secondary" style={{ padding: "6px 10px", fontSize: 12 }} onClick={() => setEditing(true)}>
              Editar
            </button>
            <button className={`vmd-btn ${item.ativo ? "vmd-btn-danger" : "vmd-btn-success"}`} style={{ padding: "6px 10px", fontSize: 12 }} onClick={() => onToggleAtivo(item)}>
              {item.ativo ? "Desativar" : "Ativar"}
            </button>
          </>
        )}
        {isSuper && item.cargo !== "super_admin" && (
          <button className="vmd-btn vmd-btn-danger" style={{ padding: "6px 10px", fontSize: 12, background: "transparent" }} onClick={() => onDelete(item)}>
            Excluir
          </button>
        )}
      </div>
    </div>
  );
}

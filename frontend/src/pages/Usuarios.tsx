import { useState, useEffect } from "react";
import { ListarUsuarios, AtivarUsuario, DeletarUsuario } from "../../wailsjs/go/main/App";
import UsuarioCard, { type UserItem } from "./usuarios/UsuarioCard";
import NovoUsuarioModal from "./usuarios/NovoUsuarioModal";
import PermissoesModal from "./usuarios/PermissoesModal";
import RedefinirSenhaModal from "./usuarios/RedefinirSenhaModal";

interface UsuariosProps {
  user: {
    id: number;
    cargo: string;
  };
}

const cargosDisponiveis: Record<string, string[]> = {
  super_admin: ["admin"],
  admin: [],
};

export default function Usuarios({ user }: UsuariosProps) {
  const [users, setUsers] = useState<UserItem[]>([]);
  const [showNovoModal, setShowNovoModal] = useState(false);
  const [permissoesUser, setPermissoesUser] = useState<UserItem | null>(null);
  const [senhaUser, setSenhaUser] = useState<UserItem | null>(null);
  const [error, setError] = useState("");

  const allowedCargos = cargosDisponiveis[user.cargo] || [];

  async function load() {
    try {
      const list = await ListarUsuarios();
      setUsers((list as unknown as UserItem[]) || []);
    } catch (e) {
      setError(String(e));
    }
  }

  useEffect(() => { load(); }, []);

  async function handleToggleAtivo(u: UserItem) {
    try {
      await AtivarUsuario(u.id, !u.ativo);
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleDelete(u: UserItem) {
    if (!confirm(`Tem certeza que deseja deletar "${u.name}"?`)) return;
    try {
      await DeletarUsuario(u.id);
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <div>
          <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>
            Usuários Administrativos
          </h2>
          <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
            Gerenciamento de contas, senhas e controle de acessos às telas do sistema.
          </p>
        </div>
        {allowedCargos.length > 0 && (
          <button onClick={() => setShowNovoModal(true)} className="vmd-btn vmd-btn-primary">
            ➕ Novo Usuário
          </button>
        )}
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 16 }}>
          <span>⚠️ {error}</span>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {users.map((u) => (
          <UsuarioCard
            key={u.id}
            item={u}
            actorCargo={user.cargo}
            allowedCargos={allowedCargos}
            onReload={load}
            onToggleAtivo={handleToggleAtivo}
            onDelete={handleDelete}
            onOpenPermissoes={(target) => setPermissoesUser(target)}
            onOpenSenha={(target) => setSenhaUser(target)}
          />
        ))}
      </div>

      {showNovoModal && (
        <NovoUsuarioModal
          allowedCargos={allowedCargos}
          onClose={() => setShowNovoModal(false)}
          onCreated={load}
        />
      )}

      {permissoesUser && (
        <PermissoesModal
          user={permissoesUser}
          onClose={() => setPermissoesUser(null)}
          onSaved={load}
        />
      )}

      {senhaUser && (
        <RedefinirSenhaModal
          user={senhaUser}
          onClose={() => setSenhaUser(null)}
          onSaved={load}
        />
      )}
    </div>
  );
}

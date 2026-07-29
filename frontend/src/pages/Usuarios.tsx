import { useState, useEffect } from "react";
import {
  CriarUsuario,
  ListarUsuarios,
  AtivarUsuario,
  DeletarUsuario,
  AtualizarUsuario,
} from "../../wailsjs/go/main/App";

interface User {
  id: number;
  name: string;
  email: string;
  cargo: string;
  ativo: boolean;
  criado_em: string;
}

const cargosDisponiveis: Record<string, string[]> = {
  super_admin: ["admin"],
  admin: [],
};

interface UsuariosProps {
  user: User;
}

function Usuarios({ user }: UsuariosProps) {
  const [users, setUsers] = useState<User[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [cargo, setCargo] = useState("");
  const [error, setError] = useState("");

  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editEmail, setEditEmail] = useState("");
  const [editCargo, setEditCargo] = useState("");

  const allowedCargos = cargosDisponiveis[user.cargo] || [];

  useEffect(() => {
    if (allowedCargos.length > 0) {
      setCargo(allowedCargos[0]);
    }
  }, [user.cargo]);

  async function load() {
    try {
      const list = await ListarUsuarios();
      setUsers(list as unknown as User[]);
    } catch (e) {
      setError(String(e));
    }
  }

  useEffect(() => { load(); }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await CriarUsuario(name, email, senha, cargo, 0);
      setName("");
      setEmail("");
      setSenha("");
      setShowModal(false);
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  async function toggleAtivo(u: User) {
    try {
      await AtivarUsuario(u.id, !u.ativo);
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  async function deletarUsuario(u: User) {
    if (!confirm(`Tem certeza que deseja deletar "${u.name}"?`)) return;
    try {
      await DeletarUsuario(u.id);
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  function iniciarEdicao(u: User) {
    setEditingId(u.id);
    setEditName(u.name);
    setEditEmail(u.email);
    setEditCargo(u.cargo);
    setError("");
  }

  async function salvarEdicao(u: User) {
    try {
      await AtualizarUsuario(u.id, editName, editEmail, editCargo);
      setEditingId(null);
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  function podeGerenciar(u: User): boolean {
    if (user.cargo === "super_admin") return true;
    if (user.cargo === "admin" && u.cargo === "aluno") return true;
    return false;
  }

  const cargoBadgeClasses: Record<string, string> = {
    super_admin: "vmd-badge vmd-badge-danger",
    admin: "vmd-badge vmd-badge-warning",
    aluno: "vmd-badge vmd-badge-success",
  };

  const cargoLabels: Record<string, string> = {
    super_admin: "Super Admin",
    admin: "Administrador",
    aluno: "Aluno/Cliente",
  };

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <div>
          <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Usuários Administrativos</h2>
          <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
            Gerenciamento de credenciais e permissões de acesso ao sistema.
          </p>
        </div>
        {allowedCargos.length > 0 && (
          <button onClick={() => setShowModal(true)} className="vmd-btn vmd-btn-primary">
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
          <div
            key={u.id}
            className="vmd-card"
            style={{
              padding: "16px 20px",
              opacity: u.ativo ? 1 : 0.6,
              background: u.ativo ? "var(--bg-card)" : "rgba(255,255,255,0.01)",
            }}
          >
            {editingId === u.id ? (
              <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap", width: "100%" }}>
                <div style={{ flex: 1, minWidth: 150 }}>
                  <label className="vmd-label">Nome</label>
                  <input
                    value={editName}
                    onChange={(e) => setEditName(e.target.value)}
                    className="vmd-input"
                  />
                </div>
                <div style={{ flex: 1, minWidth: 180 }}>
                  <label className="vmd-label">E-mail</label>
                  <input
                    value={editEmail}
                    onChange={(e) => setEditEmail(e.target.value)}
                    className="vmd-input"
                  />
                </div>
                <div style={{ width: 120 }}>
                  <label className="vmd-label">Cargo</label>
                  <select
                    value={editCargo}
                    onChange={(e) => setEditCargo(e.target.value)}
                    className="vmd-select"
                  >
                    {allowedCargos.map((c) => (
                      <option key={c} value={c}>Admin</option>
                    ))}
                  </select>
                </div>
                <div style={{ display: "flex", gap: 8, alignSelf: "flex-end", marginTop: 12 }}>
                  <button onClick={() => salvarEdicao(u)} className="vmd-btn vmd-btn-success" style={{ padding: "6px 12px", height: 32 }}>
                    Salvar
                  </button>
                  <button onClick={() => setEditingId(null)} className="vmd-btn vmd-btn-secondary" style={{ padding: "6px 12px", height: 32 }}>
                    Cancelar
                  </button>
                </div>
              </div>
            ) : (
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 12 }}>
                <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                  <span className={cargoBadgeClasses[u.cargo] || "vmd-badge"}>
                    {cargoLabels[u.cargo] || u.cargo}
                  </span>
                  <div>
                    <div style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)", display: "flex", alignItems: "center", gap: 8 }}>
                      {u.name}
                      {!u.ativo && <span className="vmd-badge vmd-badge-danger">Inativo</span>}
                    </div>
                    <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>{u.email}</div>
                  </div>
                </div>

                <div style={{ display: "flex", gap: 8 }}>
                  {podeGerenciar(u) && (
                    <>
                      <button onClick={() => iniciarEdicao(u)} className="vmd-btn vmd-btn-secondary" style={{ padding: "6px 12px", height: 32 }}>
                        Editar
                      </button>
                      <button
                        onClick={() => toggleAtivo(u)}
                        className={`vmd-btn ${u.ativo ? "vmd-btn-danger" : "vmd-btn-success"}`}
                        style={{ padding: "6px 12px", height: 32 }}
                      >
                        {u.ativo ? "Desativar" : "Ativar"}
                      </button>
                      {u.cargo !== "super_admin" && (
                        <button
                          onClick={() => deletarUsuario(u)}
                          className="vmd-btn vmd-btn-danger"
                          style={{ padding: "6px 12px", height: 32, background: "transparent" }}
                        >
                          Excluir
                        </button>
                      )}
                    </>
                  )}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {showModal && (
        <div className="vmd-modal-overlay" onClick={() => setShowModal(false)}>
          <form
            onSubmit={handleCreate}
            className="vmd-modal-content"
            style={{ maxWidth: 440 }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="vmd-modal-header">
              <h3 style={{ fontSize: 16, margin: 0 }}>Criar Novo Usuário</h3>
              <button
                type="button"
                onClick={() => setShowModal(false)}
                className="vmd-btn vmd-btn-ghost"
                style={{ padding: 4, minWidth: "auto" }}
              >
                ✕
              </button>
            </div>

            <div className="vmd-modal-body">
              <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Nome</label>
                  <input
                    placeholder="Nome completo"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    required
                    className="vmd-input"
                  />
                </div>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">E-mail</label>
                  <input
                    placeholder="exemplo@dominio.com"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    className="vmd-input"
                  />
                </div>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Senha Provisória</label>
                  <input
                    placeholder="Senha de acesso"
                    type="password"
                    value={senha}
                    onChange={(e) => setSenha(e.target.value)}
                    required
                    className="vmd-input"
                  />
                </div>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Cargo / Permissões</label>
                  <select
                    value={cargo}
                    onChange={(e) => setCargo(e.target.value)}
                    className="vmd-select"
                  >
                    {allowedCargos.map((c) => (
                      <option key={c} value={c}>Administrador (Admin)</option>
                    ))}
                  </select>
                </div>

                {error && <div style={{ color: "var(--danger)", fontSize: 12 }}>{error}</div>}
              </div>
            </div>

            <div className="vmd-modal-footer">
              <button
                type="button"
                onClick={() => setShowModal(false)}
                className="vmd-btn vmd-btn-secondary"
              >
                Cancelar
              </button>
              <button type="submit" className="vmd-btn vmd-btn-primary">
                Criar Conta
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}

export default Usuarios;

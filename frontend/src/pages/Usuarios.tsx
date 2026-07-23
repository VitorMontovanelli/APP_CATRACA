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

  const cargoColor: Record<string, string> = {
    super_admin: "#f87171",
    admin: "#60a5fa",
    aluno: "#fbbf24",
  };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 500 }}>Usuários</h2>
        {allowedCargos.length > 0 && (
          <button
            onClick={() => setShowModal(true)}
            style={{
              padding: "8px 16px",
              background: "#5eead4",
              color: "#053b32",
              border: "none",
              borderRadius: 6,
              fontSize: 13,
              fontWeight: 500,
              fontFamily: "inherit",
              cursor: "pointer",
            }}
          >
            Novo Usuário
          </button>
        )}
      </div>

      {showModal && (
        <div style={{
          position: "fixed", inset: 0, background: "rgba(0,0,0,0.7)",
          display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
        }} onClick={() => setShowModal(false)}>
          <form
            onSubmit={handleCreate}
            style={{
              background: "#121a29", border: "1px solid #232e42", borderRadius: 12,
              padding: 28, width: 440, maxWidth: "90vw",
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <h3 style={{ margin: "0 0 20px 0", fontSize: 16, fontWeight: 500 }}>Novo Usuário</h3>

            <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
              <input
                placeholder="Nome"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                style={inputStyle}
              />
              <input
                placeholder="Email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                style={inputStyle}
              />
              <input
                placeholder="Senha"
                type="password"
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
                required
                style={inputStyle}
              />
              <select
                value={cargo}
                onChange={(e) => setCargo(e.target.value)}
                style={inputStyle}
              >
                {allowedCargos.map((c) => (
                  <option key={c} value={c}>Admin</option>
                ))}
              </select>

              {error && <div style={{ color: "#f87171", fontSize: 12 }}>{error}</div>}
            </div>

            <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 24 }}>
              <button
                type="button"
                onClick={() => setShowModal(false)}
                style={{
                  padding: "8px 16px", border: "1px solid #232e42", borderRadius: 6,
                  background: "transparent", color: "#7c8798", fontSize: 12,
                  fontFamily: "inherit", cursor: "pointer",
                }}
              >
                Cancelar
              </button>
              <button
                type="submit"
                style={{
                  padding: "8px 16px", border: "none", borderRadius: 6,
                  background: "#5eead4", color: "#053b32", fontSize: 12,
                  fontWeight: 500, fontFamily: "inherit", cursor: "pointer",
                }}
              >
                Criar
              </button>
            </div>
          </form>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {users.map((u) => (
          <div
            key={u.id}
            style={{
              background: "#121a29",
              border: "1px solid #232e42",
              borderRadius: 8,
              padding: "14px 18px",
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              opacity: u.ativo ? 1 : 0.5,
              gap: 12,
            }}
          >
            {editingId === u.id ? (
              <div style={{ display: "flex", alignItems: "center", gap: 8, flex: 1 }}>
                <input
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  style={{ ...inputStyle, flex: 1 }}
                />
                <input
                  value={editEmail}
                  onChange={(e) => setEditEmail(e.target.value)}
                  style={{ ...inputStyle, flex: 1 }}
                />
                <select
                  value={editCargo}
                  onChange={(e) => setEditCargo(e.target.value)}
                  style={{ ...inputStyle, width: 120 }}
                >
                  {allowedCargos.map((c) => (
                    <option key={c} value={c}>Admin</option>
                  ))}
                </select>
                <button
                  onClick={() => salvarEdicao(u)}
                  style={{
                    padding: "6px 12px",
                    border: "none",
                    borderRadius: 6,
                    background: "#5eead4",
                    color: "#053b32",
                    fontSize: 12,
                    fontFamily: "inherit",
                    cursor: "pointer",
                  }}
                >
                  Salvar
                </button>
                <button
                  onClick={() => setEditingId(null)}
                  style={{
                    padding: "6px 12px",
                    border: "1px solid #232e42",
                    borderRadius: 6,
                    background: "transparent",
                    color: "#7c8798",
                    fontSize: 12,
                    fontFamily: "inherit",
                    cursor: "pointer",
                  }}
                >
                  Cancelar
                </button>
              </div>
            ) : (
              <>
                <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                  <span
                    style={{
                      padding: "3px 10px",
                      borderRadius: 4,
                      fontSize: 11,
                      fontWeight: 500,
                      background: cargoColor[u.cargo] + "22",
                      color: cargoColor[u.cargo],
                    }}
                  >
                    {u.cargo}
                  </span>
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 500 }}>{u.name}</div>
                    <div style={{ fontSize: 12, color: "#7c8798" }}>{u.email}</div>
                  </div>
                </div>

                <div style={{ display: "flex", gap: 8 }}>
                  {podeGerenciar(u) && (
                    <>
                      <button
                        onClick={() => iniciarEdicao(u)}
                        style={{
                          padding: "6px 12px",
                          border: "1px solid #60a5fa",
                          borderRadius: 6,
                          background: "transparent",
                          color: "#60a5fa",
                          fontSize: 12,
                          fontFamily: "inherit",
                          cursor: "pointer",
                        }}
                      >
                        Editar
                      </button>
                      <button
                        onClick={() => toggleAtivo(u)}
                        style={{
                          padding: "6px 12px",
                          border: "1px solid #232e42",
                          borderRadius: 6,
                          background: "transparent",
                          color: u.ativo ? "#f87171" : "#5eead4",
                          fontSize: 12,
                          fontFamily: "inherit",
                          cursor: "pointer",
                        }}
                      >
                        {u.ativo ? "Desativar" : "Ativar"}
                      </button>
                      <button
                        onClick={() => deletarUsuario(u)}
                        style={{
                          padding: "6px 12px",
                          border: "1px solid #f87171",
                          borderRadius: 6,
                          background: "transparent",
                          color: "#f87171",
                          fontSize: 12,
                          fontFamily: "inherit",
                          cursor: "pointer",
                        }}
                      >
                        Deletar
                      </button>
                    </>
                  )}
                </div>
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  height: 38,
  padding: "0 12px",
  background: "#0e1420",
  border: "1px solid #232e42",
  borderRadius: 6,
  color: "#e5e7eb",
  fontSize: 13,
  fontFamily: "inherit",
  outline: "none",
};

export default Usuarios;

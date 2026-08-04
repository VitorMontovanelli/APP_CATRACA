import { useState, useEffect } from "react";
import { ListarUsuarios, SalvarPermissoesUsuario } from "../../wailsjs/go/main/App";

interface User {
  id: number;
  name: string;
  email: string;
  cargo: string;
  ativo: boolean;
  permissoes: string;
}

interface PageOption {
  key: string;
  label: string;
  icon: string;
  desc: string;
}

const AVAILABLE_PAGES: PageOption[] = [
  { key: "home", label: "Visão Geral", icon: "📊", desc: "Painel financeiro, matrículas e fluxo da catraca." },
  { key: "alunos", label: "Alunos / Clientes", icon: "👥", desc: "Cadastro, foto, biometria e ficha cadastral." },
  { key: "planos", label: "Planos de Acesso", icon: "📋", desc: "Configuração de mensalidades, prazos e carência." },
  { key: "financeiro", label: "Visão Financeira", icon: "💰", desc: "Fluxo de caixa, faturas emitidas e assinaturas ativas." },
  { key: "cobranca", label: "Cobranças", icon: "⚖️", desc: "Gestão de inadimplentes e upload de comprovantes." },
  { key: "metodos_pagamento", label: "Métodos de Pagamento", icon: "💳", desc: "Gateways, chaves PIX e configurações de pagamento." },
  { key: "usuarios", label: "Usuários", icon: "👥", desc: "Cadastro de contas administrativas secundárias (Admins)." },
  { key: "logs", label: "Logs de Acesso", icon: "📋", desc: "Registros de passagens na catraca e logs de auditoria." },
];

function Acesso() {
  const [users, setUsers] = useState<User[]>([]);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successMsg, setSuccessMsg] = useState("");
  const [errorMsg, setErrorMsg] = useState("");

  useEffect(() => {
    loadUsers();
  }, []);

  async function loadUsers() {
    try {
      setLoading(true);
      const list = await ListarUsuarios();
      // Filter only "admin" cargo, super_admins have absolute permission and cannot be limited
      const admins = ((list as unknown as User[]) || []).filter((u) => u.cargo === "admin");
      setUsers(admins);
      if (admins.length > 0) {
        selectUser(admins[0]);
      } else {
        setSelectedUser(null);
      }
    } catch (e) {
      setErrorMsg(String(e));
    } finally {
      setLoading(false);
    }
  }

  function selectUser(user: User) {
    setSelectedUser(user);
    const perms = user.permissoes
      ? user.permissoes.split(",")
      : ["home", "alunos", "planos", "financeiro", "cobranca", "metodos_pagamento", "usuarios", "logs"];
    setSelectedPermissions(perms);
    setSuccessMsg("");
    setErrorMsg("");
  }

  function togglePermission(key: string) {
    if (selectedPermissions.includes(key)) {
      setSelectedPermissions(selectedPermissions.filter((p) => p !== key));
    } else {
      setSelectedPermissions([...selectedPermissions, key]);
    }
  }

  async function handleSave() {
    if (!selectedUser) return;
    setSaving(true);
    setSuccessMsg("");
    setErrorMsg("");
    try {
      const permsStr = selectedPermissions.join(",");
      await SalvarPermissoesUsuario(selectedUser.id, permsStr);
      setSuccessMsg(`Permissões de ${selectedUser.name} salvas com sucesso!`);
      // Update local users state
      setUsers(
        users.map((u) => (u.id === selectedUser.id ? { ...u, permissoes: permsStr } : u))
      );
    } catch (e) {
      setErrorMsg(String(e));
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 48, color: "var(--text-muted)", animation: "fadeIn 0.3s ease-out" }}>
        Carregando painel de permissões...
      </div>
    );
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      {/* Title */}
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Controle de Acessos</h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Defina quais telas os administradores auxiliares criados podem visualizar no sistema.
        </p>
      </div>

      {successMsg && (
        <div className="vmd-alert vmd-alert-success" style={{ marginBottom: 20 }}>
          <span>✓</span> {successMsg}
        </div>
      )}

      {errorMsg && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 20 }}>
          <span>⚠️</span> {errorMsg}
        </div>
      )}

      {users.length === 0 ? (
        <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
          Nenhum usuário administrador secundário cadastrado no sistema.
          <br />
          Crie um administrador na tela de <strong>Usuários</strong> para gerenciar seus acessos.
        </div>
      ) : (
        <div className="vmd-grid-2">
          {/* Left panel - Users List */}
          <div className="vmd-card" style={{ display: "flex", flexDirection: "column", gap: 16 }}>
            <h3 style={{ fontSize: 15, fontWeight: 600, margin: 0 }}>Selecione o Administrador</h3>
            <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 4 }}>
              {users.map((u) => {
                const isSelected = selectedUser?.id === u.id;
                return (
                  <div
                    key={u.id}
                    onClick={() => selectUser(u)}
                    className="vmd-card"
                    style={{
                      padding: "12px 16px",
                      cursor: "pointer",
                      backgroundColor: isSelected ? "var(--bg-active)" : "rgba(255,255,255,0.01)",
                      borderColor: isSelected ? "var(--primary)" : "var(--border-color)",
                      transition: "all 0.2s",
                    }}
                  >
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                      <div>
                        <div style={{ fontWeight: 600, color: "var(--text-main)", fontSize: 14 }}>
                          {u.name}
                        </div>
                        <div style={{ color: "var(--text-muted)", fontSize: 11, marginTop: 2 }}>
                          {u.email}
                        </div>
                      </div>
                      <span className="vmd-badge vmd-badge-secondary" style={{ fontSize: 10 }}>
                        {u.ativo ? "Ativo" : "Inativo"}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Right panel - Permissions Matrix */}
          {selectedUser && (
            <div className="vmd-card" style={{ display: "flex", flexDirection: "column", gap: 20 }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 12 }}>
                <div>
                  <h3 style={{ fontSize: 15, fontWeight: 600, margin: 0 }}>Permissões para {selectedUser.name}</h3>
                  <p style={{ color: "var(--text-muted)", fontSize: 11, marginTop: 2 }}>
                    Marque as telas que este usuário terá permissão de visualizar.
                  </p>
                </div>
              </div>

              {/* Checkboxes List */}
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {AVAILABLE_PAGES.map((page) => {
                  const isChecked = selectedPermissions.includes(page.key);
                  return (
                    <label
                      key={page.key}
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: 14,
                        padding: "12px 16px",
                        borderRadius: "var(--radius-md)",
                        border: "1px solid var(--border-color)",
                        backgroundColor: isChecked ? "rgba(99, 102, 241, 0.03)" : "transparent",
                        cursor: "pointer",
                        transition: "all 0.2s",
                      }}
                    >
                      <input
                        type="checkbox"
                        checked={isChecked}
                        onChange={() => togglePermission(page.key)}
                        style={{
                          width: 16,
                          height: 16,
                          accentColor: "var(--primary)",
                          cursor: "pointer",
                        }}
                      />
                      <span style={{ fontSize: 20 }}>{page.icon}</span>
                      <div style={{ flex: 1 }}>
                        <div style={{ fontWeight: 600, color: "var(--text-main)", fontSize: 13 }}>
                          {page.label}
                        </div>
                        <div style={{ color: "var(--text-muted)", fontSize: 11, marginTop: 2 }}>
                          {page.desc}
                        </div>
                      </div>
                    </label>
                  );
                })}
              </div>

              {/* Action Button */}
              <div style={{ marginTop: 8, display: "flex", justifyContent: "flex-end" }}>
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="vmd-btn vmd-btn-primary"
                  style={{ minWidth: 150 }}
                >
                  {saving ? "Salvando..." : "💾 Salvar Permissões"}
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default Acesso;

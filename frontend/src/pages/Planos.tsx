import { useState, useEffect } from "react";
import {
  ListarPlanos,
  CriarPlano,
  AtualizarPlano,
  AtivarPlano,
  DeletarPlano,
  ListarAlunosPorPlano,
} from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

interface Plan {
  id: number;
  name: string;
  description: string;
  duration_days: number;
  price_cents: number;
  grace_period_days: number;
  active: boolean;
  created_at: string;
  updated_at?: string;
}

interface PlanosProps {
  user: { cargo: string };
}

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function formatDate(d: string): string {
  if (!d) return "—";
  return d.replace("T", " ").split(" ")[0];
}

const PAGE_SIZE = 5;

function Planos({ user }: PlanosProps) {
  const [plans, setPlans] = useState<Plan[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editingPlan, setEditingPlan] = useState<Plan | null>(null);
  
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [durationDays, setDurationDays] = useState(30);
  const [priceCents, setPriceCents] = useState(0);
  const [gracePeriodDays, setGracePeriodDays] = useState(5);
  const [error, setError] = useState("");

  const [studentsMap, setStudentsMap] = useState<Record<number, main.AlunoPorPlano[]>>({});
  const [totalMap, setTotalMap] = useState<Record<number, number>>({});
  const [pageMap, setPageMap] = useState<Record<number, number>>({});
  const [loadingStudents, setLoadingStudents] = useState<Record<number, boolean>>({});
  const [expandedId, setExpandedId] = useState<number | null>(null);

  const canManage = user.cargo === "super_admin" || user.cargo === "admin";

  useEffect(() => {
    load();
  }, []);

  async function load() {
    try {
      const list = await ListarPlanos();
      setPlans(list as unknown as Plan[]);
    } catch (e) {
      setError(String(e));
    }
  }

  async function loadStudents(planId: number, page: number = 1) {
    setLoadingStudents((prev) => ({ ...prev, [planId]: true }));
    setPageMap((prev) => ({ ...prev, [planId]: page }));
    try {
      const res = await ListarAlunosPorPlano(planId, page, PAGE_SIZE) as any;
      const list = ((Array.isArray(res) ? res[0] : null) || []) as main.AlunoPorPlano[];
      const total = ((Array.isArray(res) ? res[1] : null) || 0) as number;
      setStudentsMap((prev) => ({ ...prev, [planId]: list }));
      setTotalMap((prev) => ({ ...prev, [planId]: total }));
    } catch (e) {
      setError(String(e));
    } finally {
      setLoadingStudents((prev) => ({ ...prev, [planId]: false }));
    }
  }

  function toggleExpand(planId: number) {
    if (expandedId === planId) {
      setExpandedId(null);
    } else {
      setExpandedId(planId);
      if (!studentsMap[planId]) {
        loadStudents(planId);
      }
    }
  }

  function iniciarEdicao(p: Plan) {
    setEditingPlan(p);
    setName(p.name);
    setDescription(p.description);
    setDurationDays(p.duration_days);
    setPriceCents(p.price_cents);
    setGracePeriodDays(p.grace_period_days);
    setError("");
    setShowModal(true);
  }

  function fecharModal() {
    setShowModal(false);
    setEditingPlan(null);
    setName("");
    setDescription("");
    setPriceCents(0);
    setDurationDays(30);
    setGracePeriodDays(5);
    setError("");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      if (editingPlan) {
        await AtualizarPlano(editingPlan.id, name, description, durationDays, priceCents, gracePeriodDays);
      } else {
        await CriarPlano(name, description, durationDays, priceCents, gracePeriodDays);
      }
      fecharModal();
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  async function toggleAtivo(p: Plan) {
    try {
      await AtivarPlano(p.id, !p.active);
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleDelete(id: number) {
    if (!canManage) return;
    try {
      await DeletarPlano(id);
      setExpandedId((prev) => (prev === id ? null : prev));
      await load();
    } catch (err) {
      setError(String(err));
    }
  }

  const statusLabel: Record<string, string> = {
    active: "Regular",
    overdue: "Inadimplente",
    suspended: "Suspenso",
    cancelled: "Cancelado",
    expired: "Expirado",
  };

  const statusBadgeClasses: Record<string, string> = {
    active: "vmd-badge vmd-badge-success",
    overdue: "vmd-badge vmd-badge-danger",
    suspended: "vmd-badge vmd-badge-secondary",
    cancelled: "vmd-badge vmd-badge-secondary",
    expired: "vmd-badge vmd-badge-secondary",
  };

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <div>
          <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Planos de Acesso</h2>
          <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
            Gerencie as modalidades de planos e planos vinculados aos alunos.
          </p>
        </div>
        {canManage && (
          <button onClick={() => setShowModal(true)} className="vmd-btn vmd-btn-primary">
            ➕ Novo Plano
          </button>
        )}
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 16 }}>
          <span>⚠️ {error}</span>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {plans.map((p) => {
          const students = studentsMap[p.id] || [];
          const total = totalMap[p.id] || 0;
          const page = pageMap[p.id] || 1;
          const totalPages = Math.ceil(total / PAGE_SIZE);
          const loading = loadingStudents[p.id];

          return (
            <div key={p.id} className="vmd-card" style={{ padding: 0, overflow: "hidden" }}>
              <div
                style={{
                  padding: "16px 20px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  opacity: p.active ? 1 : 0.6,
                  gap: 12,
                  background: expandedId === p.id ? "rgba(255, 255, 255, 0.01)" : "transparent",
                  borderBottom: expandedId === p.id ? "1px solid var(--border-color)" : "none",
                }}
              >
                <div
                  style={{ display: "flex", alignItems: "center", gap: 12, flex: 1, cursor: "pointer" }}
                  onClick={() => toggleExpand(p.id)}
                >
                  <span
                    style={{
                      color: "var(--text-muted)",
                      fontSize: 10,
                      transition: "transform 0.2s",
                      display: "inline-block",
                      transform: expandedId === p.id ? "rotate(90deg)" : "rotate(0deg)",
                    }}
                  >
                    ▶
                  </span>
                  <div>
                    <div style={{ fontSize: 15, fontWeight: 600, display: "flex", alignItems: "center", gap: 8 }}>
                      {p.name}
                      <span className={`vmd-badge ${p.active ? "vmd-badge-success" : "vmd-badge-danger"}`}>
                        {p.active ? "Ativo" : "Inativo"}
                      </span>
                    </div>
                    <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>{p.description || "Sem descrição."}</div>
                    <div style={{ fontSize: 11, color: "var(--text-dim)", marginTop: 4, fontWeight: 600 }}>
                      DURAÇÃO: {p.duration_days} dias · PREÇO: {formatCents(p.price_cents)} · CARÊNCIA: {p.grace_period_days}d
                    </div>
                  </div>
                </div>

                <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
                  {canManage && (
                    <>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          iniciarEdicao(p);
                        }}
                        className="vmd-btn vmd-btn-secondary"
                        style={{ padding: "6px 12px", height: 32 }}
                      >
                        Editar
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          toggleAtivo(p);
                        }}
                        className={`vmd-btn ${p.active ? "vmd-btn-danger" : "vmd-btn-success"}`}
                        style={{ padding: "6px 12px", height: 32 }}
                      >
                        {p.active ? "Desativar" : "Ativar"}
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDelete(p.id);
                        }}
                        className="vmd-btn vmd-btn-danger"
                        style={{ padding: "6px 12px", height: 32, background: "transparent" }}
                      >
                        Excluir
                      </button>
                    </>
                  )}
                </div>
              </div>

              {expandedId === p.id && (
                <div style={{ padding: "20px 24px", background: "rgba(3,7,18,0.2)" }}>
                  <div style={{ fontSize: 12, fontWeight: 600, color: "var(--text-muted)", marginBottom: 12, textTransform: "uppercase", letterSpacing: "0.5px" }}>
                    Alunos Matriculados neste Plano
                  </div>
                  {loading ? (
                    <div style={{ color: "var(--text-muted)", fontSize: 13, padding: "8px 0" }}>Carregando dados dos alunos...</div>
                  ) : students.length === 0 ? (
                    <div style={{ color: "var(--text-muted)", fontSize: 13, padding: "8px 0" }}>Nenhum aluno vinculado a este plano atualmente.</div>
                  ) : (
                    <>
                      <div className="vmd-table-container">
                        <table className="vmd-table">
                          <thead>
                            <tr>
                              <th>Aluno</th>
                              <th>CPF</th>
                              <th>Status do Aluno</th>
                              <th>Início</th>
                            </tr>
                          </thead>
                          <tbody>
                            {students.map((s) => (
                              <tr key={s.student_id}>
                                <td style={{ fontWeight: 600 }}>{s.nome}</td>
                                <td>{s.cpf}</td>
                                <td>
                                  <span className={statusBadgeClasses[s.status] || "vmd-badge"}>
                                    {statusLabel[s.status] || s.status}
                                  </span>
                                </td>
                                <td>{formatDate(s.start_date)}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>

                      {totalPages > 1 && (
                        <div style={{ display: "flex", justifyContent: "center", gap: 10, marginTop: 16 }}>
                          <button
                            disabled={page <= 1}
                            onClick={() => loadStudents(p.id, page - 1)}
                            className="vmd-btn vmd-btn-secondary"
                            style={{ padding: "4px 10px", height: 28 }}
                          >
                            Anterior
                          </button>
                          <span style={{ color: "var(--text-muted)", fontSize: 12, display: "flex", alignItems: "center" }}>
                            Página {page} de {totalPages} ({total} alunos)
                          </span>
                          <button
                            disabled={page >= totalPages}
                            onClick={() => loadStudents(p.id, page + 1)}
                            className="vmd-btn vmd-btn-secondary"
                            style={{ padding: "4px 10px", height: 28 }}
                          >
                            Próxima
                          </button>
                        </div>
                      )}
                    </>
                  )}
                </div>
              )}
            </div>
          );
        })}
        {plans.length === 0 && (
          <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
            Nenhum plano de acesso cadastrado no sistema.
          </div>
        )}
      </div>

      {showModal && (
        <div className="vmd-modal-overlay" onClick={fecharModal}>
          <form
            onSubmit={handleSubmit}
            className="vmd-modal-content"
            style={{ maxWidth: 500 }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="vmd-modal-header">
              <h3 style={{ fontSize: 16, margin: 0 }}>
                {editingPlan ? "Editar Plano de Acesso" : "Criar Novo Plano de Acesso"}
              </h3>
              <button
                type="button"
                onClick={fecharModal}
                className="vmd-btn vmd-btn-ghost"
                style={{ padding: 4, minWidth: "auto" }}
              >
                ✕
              </button>
            </div>

            <div className="vmd-modal-body">
              <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Nome do Plano</label>
                  <input
                    placeholder="Ex: Plano Anual VIP"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    required
                    className="vmd-input"
                  />
                </div>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Descrição</label>
                  <input
                    placeholder="Breve descrição dos benefícios"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    className="vmd-input"
                  />
                </div>

                <div className="vmd-grid-3">
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Duração (dias)</label>
                    <input
                      type="number"
                      min={1}
                      value={durationDays}
                      onChange={(e) => setDurationDays(Number(e.target.value))}
                      required
                      className="vmd-input"
                    />
                  </div>
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Preço</label>
                    <input
                      type="text"
                      value={formatCents(priceCents)}
                      onChange={(e) => {
                        const rawDigits = e.target.value.replace(/\D/g, "");
                        const cents = rawDigits ? parseInt(rawDigits, 10) : 0;
                        setPriceCents(cents);
                      }}
                      required
                      className="vmd-input"
                    />
                  </div>
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Carência (dias)</label>
                    <input
                      type="number"
                      min={0}
                      value={gracePeriodDays}
                      onChange={(e) => setGracePeriodDays(Number(e.target.value))}
                      className="vmd-input"
                    />
                  </div>
                </div>


                {error && <div style={{ color: "var(--danger)", fontSize: 12 }}>{error}</div>}
              </div>
            </div>

            <div className="vmd-modal-footer">
              <button
                type="button"
                onClick={fecharModal}
                className="vmd-btn vmd-btn-secondary"
              >
                Cancelar
              </button>
              <button type="submit" className="vmd-btn vmd-btn-primary">
                Salvar Plano
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}

export default Planos;

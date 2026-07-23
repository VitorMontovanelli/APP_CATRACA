import { useState, useEffect } from "react";
import {
  ListarPlanos,
  CriarPlano,
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

const inputStyle: React.CSSProperties = {
  height: 36,
  padding: "0 10px",
  background: "#0e1420",
  border: "1px solid #232e42",
  borderRadius: 6,
  color: "#e5e7eb",
  fontSize: 12,
  fontFamily: "inherit",
  outline: "none",
  width: "100%",
};

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

  useEffect(() => { load(); }, []);

  async function load() {
    try {
      const list = await ListarPlanos();
      setPlans(list as unknown as Plan[]);
    } catch (e) { setError(String(e)); }
  }

  async function loadStudents(planId: number, page: number = 1) {
    setLoadingStudents((prev) => ({ ...prev, [planId]: true }));
    setPageMap((prev) => ({ ...prev, [planId]: page }));
    try {
      const res = await ListarAlunosPorPlano(planId, page, PAGE_SIZE);
      const list = (res as unknown as main.AlunoPorPlano[][])[0] || [];
      const total = (res as unknown as number[])[1] || 0;
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

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await CriarPlano(name, description, durationDays, priceCents, gracePeriodDays);
      setName(""); setDescription(""); setPriceCents(0);
      setShowModal(false);
      await load();
    } catch (err) { setError(String(err)); }
  }

  async function toggleAtivo(p: Plan) {
    try {
      await AtivarPlano(p.id, !p.active);
      await load();
    } catch (err) { setError(String(err)); }
  }

  async function handleDelete(id: number) {
    if (!canManage) return;
    try {
      await DeletarPlano(id);
      setExpandedId((prev) => prev === id ? null : prev);
      await load();
    } catch (err) { setError(String(err)); }
  }

  const statusLabel: Record<string, string> = {
    active: "Ativo",
    overdue: "Inadimplente",
    suspended: "Suspenso",
    cancelled: "Cancelado",
    expired: "Expirado",
  };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 500 }}>Planos</h2>
        {canManage && (
          <button onClick={() => setShowModal(true)} style={{
            padding: "8px 16px", background: "#5eead4", color: "#053b32",
            border: "none", borderRadius: 6, fontSize: 13, fontWeight: 500,
            fontFamily: "inherit", cursor: "pointer",
          }}>
            Novo Plano
          </button>
        )}
      </div>

      {showModal && (
        <div style={{
          position: "fixed", inset: 0, background: "rgba(0,0,0,0.7)",
          display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
        }} onClick={() => setShowModal(false)}>
          <form onSubmit={handleCreate} style={{
            background: "#121a29", border: "1px solid #232e42", borderRadius: 12,
            padding: 28, width: 500, maxWidth: "90vw",
          }} onClick={(e) => e.stopPropagation()}>
            <h3 style={{ margin: "0 0 20px 0", fontSize: 16, fontWeight: 500 }}>Novo Plano</h3>

            <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
              <input placeholder="Nome do Plano" value={name} onChange={(e) => setName(e.target.value)} required style={inputStyle} />
              <input placeholder="Descrição" value={description} onChange={(e) => setDescription(e.target.value)} style={inputStyle} />
              <div style={{ display: "flex", gap: 8 }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 11, color: "#a7b0bf", marginBottom: 6 }}>Duração (dias)</div>
                  <input type="number" min={1} value={durationDays} onChange={(e) => setDurationDays(Number(e.target.value))} required style={inputStyle} />
                </div>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 11, color: "#a7b0bf", marginBottom: 6 }}>Preço (centavos)</div>
                  <input type="number" min={0} value={priceCents} onChange={(e) => setPriceCents(Number(e.target.value))} required style={inputStyle} />
                </div>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 11, color: "#a7b0bf", marginBottom: 6 }}>Carência (dias)</div>
                  <input type="number" min={0} value={gracePeriodDays} onChange={(e) => setGracePeriodDays(Number(e.target.value))} style={inputStyle} />
                </div>
              </div>
              {priceCents > 0 && <div style={{ color: "#7c8798", fontSize: 12 }}>Valor: {formatCents(priceCents)}</div>}
              {error && <div style={{ color: "#f87171", fontSize: 12 }}>{error}</div>}
            </div>

            <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 24 }}>
              <button type="button" onClick={() => setShowModal(false)} style={{
                padding: "8px 16px", border: "1px solid #232e42", borderRadius: 6,
                background: "transparent", color: "#7c8798", fontSize: 12,
                fontFamily: "inherit", cursor: "pointer",
              }}>Cancelar</button>
              <button type="submit" style={{
                padding: "8px 16px", border: "none", borderRadius: 6,
                background: "#5eead4", color: "#053b32", fontSize: 12,
                fontWeight: 500, fontFamily: "inherit", cursor: "pointer",
              }}>Criar Plano</button>
            </div>
          </form>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {plans.map((p) => {
          const students = studentsMap[p.id] || [];
          const total = totalMap[p.id] || 0;
          const page = pageMap[p.id] || 1;
          const totalPages = Math.ceil(total / PAGE_SIZE);
          const loading = loadingStudents[p.id];

          return (
            <div key={p.id}>
              <div style={{
                background: "#121a29", border: "1px solid #232e42", borderRadius: 8,
                padding: "14px 18px", display: "flex", alignItems: "center",
                justifyContent: "space-between", opacity: p.active ? 1 : 0.5, gap: 12,
              }}>
                <div style={{ display: "flex", alignItems: "center", gap: 12, flex: 1, cursor: "pointer" }}
                     onClick={() => toggleExpand(p.id)}>
                  <span style={{ color: "#7c8798", fontSize: 14, transition: "transform 0.2s",
                    display: "inline-block", transform: expandedId === p.id ? "rotate(90deg)" : "rotate(0deg)" }}>
                    ▶
                  </span>
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 500 }}>{p.name}</div>
                    <div style={{ fontSize: 12, color: "#7c8798" }}>{p.description || "—"}</div>
                    <div style={{ fontSize: 12, color: "#7c8798", marginTop: 2 }}>
                      {p.duration_days} dias · {formatCents(p.price_cents)} · carência: {p.grace_period_days}d
                    </div>
                  </div>
                </div>
                <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
                  <span style={{
                    padding: "3px 10px", borderRadius: 4, fontSize: 11, fontWeight: 500,
                    background: p.active ? "#5eead422" : "#f8717122",
                    color: p.active ? "#5eead4" : "#f87171",
                  }}>
                    {p.active ? "ativo" : "inativo"}
                  </span>
                  {canManage && (
                    <>
                      <button onClick={() => toggleAtivo(p)} style={{
                        padding: "6px 12px", border: "1px solid #232e42", borderRadius: 6,
                        background: "transparent", color: p.active ? "#f87171" : "#5eead4",
                        fontSize: 12, fontFamily: "inherit", cursor: "pointer",
                      }}>
                        {p.active ? "Desativar" : "Ativar"}
                      </button>
                      <button onClick={() => handleDelete(p.id)} style={{
                        padding: "6px 12px", border: "1px solid #f87171", borderRadius: 6,
                        background: "transparent", color: "#f87171", fontSize: 12,
                        fontFamily: "inherit", cursor: "pointer",
                      }}>Excluir</button>
                    </>
                  )}
                </div>
              </div>

              {expandedId === p.id && (
                <div style={{
                  background: "#0e1420", border: "1px solid #232e42", borderTop: "none",
                  borderRadius: "0 0 8px 8px", padding: "12px 18px", marginTop: -4,
                }}>
                  {loading ? (
                    <div style={{ color: "#7c8798", fontSize: 12, padding: 8 }}>Carregando...</div>
                  ) : students.length === 0 ? (
                    <div style={{ color: "#7c8798", fontSize: 12, padding: 8 }}>Nenhum aluno neste plano.</div>
                  ) : (
                    <>
                      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 12 }}>
                        <thead>
                          <tr>
                            <th style={{ padding: "6px 8px", color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}>Aluno</th>
                            <th style={{ padding: "6px 8px", color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}>CPF</th>
                            <th style={{ padding: "6px 8px", color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}>Status</th>
                            <th style={{ padding: "6px 8px", color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}>Início</th>
                          </tr>
                        </thead>
                        <tbody>
                          {students.map((s) => (
                            <tr key={s.student_id}>
                              <td style={{ padding: "6px 8px", color: "#e5e7eb", borderBottom: "1px solid #1a2440" }}>{s.nome}</td>
                              <td style={{ padding: "6px 8px", color: "#a7b0bf", borderBottom: "1px solid #1a2440" }}>{s.cpf}</td>
                              <td style={{ padding: "6px 8px", borderBottom: "1px solid #1a2440" }}>
                                <span style={{
                                  padding: "2px 6px", borderRadius: 4, fontSize: 11,
                                  background: s.status === "active" ? "#5eead422" : s.status === "overdue" ? "#f8717122" : "#7c879822",
                                  color: s.status === "active" ? "#5eead4" : s.status === "overdue" ? "#f87171" : "#7c8798",
                                }}>
                                  {statusLabel[s.status] || s.status}
                                </span>
                              </td>
                              <td style={{ padding: "6px 8px", color: "#a7b0bf", borderBottom: "1px solid #1a2440" }}>{formatDate(s.start_date)}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>

                      {totalPages > 1 && (
                        <div style={{ display: "flex", justifyContent: "center", gap: 6, marginTop: 10 }}>
                          <button disabled={page <= 1} onClick={() => loadStudents(p.id, page - 1)} style={{
                            padding: "4px 10px", border: "1px solid #232e42", borderRadius: 4,
                            background: "transparent", color: page <= 1 ? "#3a4255" : "#a7b0bf",
                            fontSize: 11, fontFamily: "inherit", cursor: page <= 1 ? "default" : "pointer",
                          }}>Anterior</button>
                          <span style={{ color: "#7c8798", fontSize: 11, padding: "4px 0" }}>
                            {page} de {totalPages} ({total} alunos)
                          </span>
                          <button disabled={page >= totalPages} onClick={() => loadStudents(p.id, page + 1)} style={{
                            padding: "4px 10px", border: "1px solid #232e42", borderRadius: 4,
                            background: "transparent", color: page >= totalPages ? "#3a4255" : "#a7b0bf",
                            fontSize: 11, fontFamily: "inherit", cursor: page >= totalPages ? "default" : "pointer",
                          }}>Próximo</button>
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
          <div style={{ color: "#7c8798", fontSize: 13, textAlign: "center", padding: 40 }}>
            Nenhum plano cadastrado.
          </div>
        )}
      </div>
    </div>
  );
}

export default Planos;

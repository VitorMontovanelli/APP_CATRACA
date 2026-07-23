import { useState, useEffect, useCallback } from "react";
import {
  ListarStudentsComPlanos,
  CriarStudentComPlano,
  AtualizarStudent,
  AtualizarPlanoAluno,
  AtivarStudent,
  ListarPaymentMethods,
  ListarPlanos,
  GerarInvoice,
} from "../../wailsjs/go/main/App";

interface StudentComPlano {
  id: number;
  nome: string;
  cpf: string;
  data_nascimento?: string;
  telefone?: string;
  email?: string;
  forma_pagamento_id?: number;
  data_entrada: string;
  observacao?: string;
  ativo: boolean;
  created_at: string;
  plano_nome?: string;
  plano_preco?: number;
  plano_status?: string;
  student_plan_id?: number;
  payment_method?: string;
  due_day?: number;
  ultima_fatura?: string;
  fatura_valor?: number;
  fatura_vencimento?: string;
}

interface Plan {
  id: number;
  name: string;
  price_cents: number;
  duration_days: number;
  active: boolean;
}

interface PaymentMethod {
  id: number;
  name: string;
  type: string;
  enabled: boolean;
}

interface AlunosProps {
  user: { cargo: string };
}

const inputStyle: React.CSSProperties = {
  height: 34, padding: "0 10px", background: "#121a29",
  border: "1px solid #232e42", borderRadius: 6, color: "#e5e7eb",
  fontSize: 12, fontFamily: "inherit", outline: "none", width: "100%", boxSizing: "border-box",
};

const inputDateStyle: React.CSSProperties = {
  ...inputStyle, colorScheme: "dark",
};

const labelStyle: React.CSSProperties = {
  fontSize: 11, color: "#a7b0bf", marginBottom: 6,
};

function maskCPF(v: string): string {
  const d = v.replace(/\D/g, "").slice(0, 11);
  return d.replace(/^(\d{3})(\d{3})(\d{3})(\d{2})$/, "$1.$2.$3-$4")
          .replace(/^(\d{3})(\d{3})(\d{1,3})$/, "$1.$2.$3")
          .replace(/^(\d{3})(\d{1,3})$/, "$1.$2");
}

function maskTel(v: string): string {
  const d = v.replace(/\D/g, "").slice(0, 11);
  return d.replace(/^(\d{2})(\d{5})(\d{4})$/, "($1) $2-$3")
          .replace(/^(\d{2})(\d{1,5})$/, "($1) $2")
          .replace(/^(\d{1,2})$/, "($1");
}

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

const statusColors: Record<string, string> = {
  pending: "#fbbf24",
  paid: "#5eead4",
  overdue: "#f87171",
  cancelled: "#7c8798",
};

const statusLabels: Record<string, string> = {
  pending: "Pendente",
  paid: "Pago",
  overdue: "Vencido",
  cancelled: "Cancelado",
};

function Alunos({ user }: AlunosProps) {
  const [students, setStudents] = useState<StudentComPlano[]>([]);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editId, setEditId] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; msg: string } | null>(null);
  const [loadingInvoice, setLoadingInvoice] = useState<number | null>(null);
  const [searchTerm, setSearchTerm] = useState("");

  const filteredStudents = searchTerm
    ? students.filter((s) => s.nome.toLowerCase().includes(searchTerm.toLowerCase()))
    : students;

  const [nome, setNome] = useState("");
  const [cpf, setCpf] = useState("");
  const [dataNasc, setDataNasc] = useState("");
  const [telefone, setTelefone] = useState("");
  const [email, setEmail] = useState("");
  const [formaPagamentoId, setFormaPagamentoId] = useState(0);
  const [planId, setPlanId] = useState(0);
  const [dueDay, setDueDay] = useState(5);

  const canManage = user.cargo === "super_admin" || user.cargo === "admin";

  const showFeedback = useCallback((type: "success" | "error", msg: string) => {
    setFeedback({ type, msg });
    setTimeout(() => setFeedback(null), 4000);
  }, []);

  useEffect(() => { load(); }, []);

  async function handleRegistrarPagamento(studentPlanId: number) {
    setLoadingInvoice(studentPlanId);
    try {
      await GerarInvoice(studentPlanId);
      showFeedback("success", "Fatura gerada com sucesso!");
      await load();
    } catch (e) {
      showFeedback("error", String(e));
    } finally {
      setLoadingInvoice(null);
    }
  }

  async function load() {
    try {
      const [s, p, m] = await Promise.all([
        ListarStudentsComPlanos(),
        ListarPlanos(),
        ListarPaymentMethods(),
      ]);
      setStudents(s as unknown as StudentComPlano[]);
      setPlans((p as unknown as Plan[]).filter((pl) => pl.active));
      setMethods((m as unknown as PaymentMethod[]).filter((pm) => pm.enabled));
    } catch (e) { setError(String(e)); }
  }

  function openCreate() {
    setEditId(null);
    setNome(""); setCpf(""); setDataNasc(""); setTelefone(""); setEmail("");
    setFormaPagamentoId(0); setPlanId(0); setDueDay(5);
    setError(""); setShowModal(true);
  }

  function openEdit(s: StudentComPlano) {
    setEditId(s.id);
    setNome(s.nome); setCpf(s.cpf);
    setDataNasc(s.data_nascimento || "");
    setTelefone(s.telefone || ""); setEmail(s.email || "");
    setFormaPagamentoId(s.forma_pagamento_id || 0);
    setPlanId(s.student_plan_id ? (plans.find(p => p.name === s.plano_nome)?.id || 0) : 0);
    setDueDay(s.due_day || 5);
    setError(""); setShowModal(true);
  }

  async function handleSave() {
    setError("");
    try {
      if (editId) {
        await AtualizarStudent(editId, nome, cpf, dataNasc, telefone, email, formaPagamentoId);
        await AtualizarPlanoAluno(editId, planId, formaPagamentoId, dueDay);
      } else {
        await CriarStudentComPlano(nome, cpf, dataNasc, telefone, email, formaPagamentoId, planId, dueDay);
      }
      setShowModal(false);
      await load();
    } catch (e) { setError(String(e)); }
  }

  async function toggleAtivo(s: StudentComPlano) {
    try {
      await AtivarStudent(s.id, !s.ativo);
      await load();
    } catch (e) { setError(String(e)); }
  }

  function formatDate(d: string): string {
    if (!d) return "—";
    return d.split("T")[0] || d;
  }

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 500 }}>Alunos</h2>
        {canManage && (
          <button onClick={openCreate} style={{
            padding: "8px 16px", background: "#5eead4", color: "#053b32",
            border: "none", borderRadius: 6, fontSize: 13, fontWeight: 500,
            fontFamily: "inherit", cursor: "pointer",
          }}>Novo Aluno</button>
        )}
      </div>

      {error && <div style={{ color: "#f87171", fontSize: 12, marginBottom: 12 }}>{error}</div>}
      {feedback && (
        <div style={{
          padding: "10px 16px", borderRadius: 6, marginBottom: 12, fontSize: 13,
          background: feedback.type === "success" ? "#5eead422" : "#f8717122",
          color: feedback.type === "success" ? "#5eead4" : "#f87171",
          border: `1px solid ${feedback.type === "success" ? "#5eead4" : "#f87171"}`,
        }}>{feedback.msg}</div>
      )}

      <div style={{ marginBottom: 16 }}>
        <input
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Buscar aluno por nome..."
          style={{
            height: 38, padding: "0 14px", background: "#121a29",
            border: "1px solid #232e42", borderRadius: 8, color: "#e5e7eb",
            fontSize: 13, fontFamily: "inherit", outline: "none", width: 320, maxWidth: "100%",
            boxSizing: "border-box",
          }}
        />
        {searchTerm && (
          <span style={{ fontSize: 12, color: "#7c8798", marginLeft: 12 }}>
            {filteredStudents.length} de {students.length} aluno(s)
          </span>
        )}
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {filteredStudents.map((s) => (
          <div key={s.id} style={{
            background: "#121a29", border: "1px solid #232e42", borderRadius: 8,
            padding: "14px 18px", opacity: s.ativo ? 1 : 0.5,
          }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, fontWeight: 500 }}>{s.nome}</div>
                <div style={{ fontSize: 12, color: "#7c8798", marginTop: 2 }}>
                  CPF: {s.cpf} {s.data_nascimento ? `· Nasc: ${formatDate(s.data_nascimento)}` : ""}
                </div>
              </div>
              <div style={{ display: "flex", gap: 8, alignItems: "center", flexShrink: 0 }}>
                <span style={{
                  padding: "3px 10px", borderRadius: 4, fontSize: 11, fontWeight: 500,
                  background: s.ativo ? "#5eead422" : "#f8717122",
                  color: s.ativo ? "#5eead4" : "#f87171",
                }}>{s.ativo ? "ativo" : "inativo"}</span>
                {canManage && (
                  <>
                    <button onClick={() => openEdit(s)} style={{
                      padding: "6px 12px", border: "1px solid #60a5fa", borderRadius: 6,
                      background: "transparent", color: "#60a5fa", fontSize: 12,
                      fontFamily: "inherit", cursor: "pointer",
                    }}>Editar</button>
                    <button onClick={() => toggleAtivo(s)} style={{
                      padding: "6px 12px", border: "1px solid #232e42", borderRadius: 6,
                      background: "transparent", color: s.ativo ? "#f87171" : "#5eead4",
                      fontSize: 12, fontFamily: "inherit", cursor: "pointer",
                    }}>{s.ativo ? "Desativar" : "Ativar"}</button>
                  </>
                )}
              </div>
            </div>

            {s.plano_nome && (
              <div style={{
                marginTop: 10, display: "flex", gap: 20, flexWrap: "wrap",
                padding: "8px 12px", background: "#0e1420", borderRadius: 6,
              }}>
                <div>
                  <div style={{ fontSize: 10, color: "#7c8798", textTransform: "uppercase" }}>Plano</div>
                  <div style={{ fontSize: 13, fontWeight: 500, marginTop: 2 }}>
                    {s.plano_nome}
                    {s.plano_preco ? ` (${formatCents(s.plano_preco)})` : ""}
                  </div>
                </div>
                <div>
                  <div style={{ fontSize: 10, color: "#7c8798", textTransform: "uppercase" }}>Forma de Pagamento</div>
                  <div style={{ fontSize: 13, marginTop: 2 }}>{s.payment_method || "—"}</div>
                </div>
                <div>
                  <div style={{ fontSize: 10, color: "#7c8798", textTransform: "uppercase" }}>Vencimento</div>
                  <div style={{ fontSize: 13, marginTop: 2 }}>Dia {s.due_day || "—"}</div>
                </div>
                <div>
                  <div style={{ fontSize: 10, color: "#7c8798", textTransform: "uppercase" }}>Status do Plano</div>
                  <div style={{
                    fontSize: 13, marginTop: 2,
                    color: s.plano_status === "active" ? "#5eead4" :
                           s.plano_status === "overdue" ? "#f87171" : "#7c8798",
                  }}>
                    {s.plano_status === "active" ? "Ativo" :
                     s.plano_status === "overdue" ? "Inadimplente" :
                     s.plano_status === "suspended" ? "Suspenso" : s.plano_status || "—"}
                  </div>
                </div>
                <div>
                  <div style={{ fontSize: 10, color: "#7c8798", textTransform: "uppercase" }}>Última Fatura</div>
                  <div style={{ fontSize: 13, marginTop: 2 }}>
                    {s.ultima_fatura ? (
                      <span style={{
                        color: statusColors[s.ultima_fatura] || "#7c8798",
                      }}>
                        {statusLabels[s.ultima_fatura] || s.ultima_fatura}
                        {s.fatura_valor ? ` · ${formatCents(s.fatura_valor)}` : ""}
                        {s.fatura_vencimento ? ` · ${s.fatura_vencimento}` : ""}
                      </span>
                    ) : "Nenhuma"}
                  </div>
                </div>
              </div>
            )}

            {canManage && s.student_plan_id && loadingInvoice !== s.student_plan_id && (
              <div style={{ marginTop: 10, textAlign: "right" }}>
                <button onClick={() => handleRegistrarPagamento(s.student_plan_id!)} style={{
                  padding: "6px 14px", border: "1px solid #5eead4", borderRadius: 6,
                  background: "transparent", color: "#5eead4", fontSize: 12,
                  fontFamily: "inherit", cursor: "pointer",
                }}>Registrar Pagamento</button>
              </div>
            )}
            {canManage && s.student_plan_id && loadingInvoice === s.student_plan_id && (
              <div style={{ marginTop: 10, textAlign: "right", color: "#7c8798", fontSize: 12 }}>
                Gerando fatura...
              </div>
            )}
          </div>
        ))}
        {filteredStudents.length === 0 && (
          <div style={{ color: "#7c8798", fontSize: 13, textAlign: "center", padding: 40 }}>
            {searchTerm ? "Nenhum aluno encontrado para esta busca." : "Nenhum aluno cadastrado."}
          </div>
        )}
      </div>

      {showModal && (
        <div style={{
          position: "fixed", inset: 0, background: "rgba(0,0,0,0.7)",
          display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
        }} onClick={() => setShowModal(false)}>
          <div style={{
            background: "#121a29", border: "1px solid #232e42", borderRadius: 12,
            padding: 28, width: 520, maxWidth: "90vw", maxHeight: "90vh", overflow: "auto",
          }} onClick={(e) => e.stopPropagation()}>
            <h3 style={{ margin: "0 0 20px 0", fontSize: 16, fontWeight: 500 }}>
              {editId ? "Editar Aluno" : "Novo Aluno"}
            </h3>

            <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
              <div>
                <div style={labelStyle}>Nome completo</div>
                <input value={nome} onChange={(e) => setNome(e.target.value)}
                  placeholder="Nome do aluno" style={inputStyle} />
              </div>

              <div style={{ display: "flex", gap: 12 }}>
                <div style={{ flex: 1 }}>
                  <div style={labelStyle}>CPF</div>
                  <input value={cpf} onChange={(e) => setCpf(maskCPF(e.target.value))}
                    placeholder="000.000.000-00" maxLength={14} style={inputStyle} />
                </div>
                <div style={{ flex: 1 }}>
                  <div style={labelStyle}>Data de nascimento</div>
                  <input type="date" value={dataNasc} onChange={(e) => setDataNasc(e.target.value)}
                    style={inputDateStyle} />
                </div>
              </div>

              <div style={{ display: "flex", gap: 12 }}>
                <div style={{ flex: 1 }}>
                  <div style={labelStyle}>Telefone</div>
                  <input value={telefone} onChange={(e) => setTelefone(maskTel(e.target.value))}
                    placeholder="(00) 00000-0000" maxLength={15} style={inputStyle} />
                </div>
                <div style={{ flex: 1 }}>
                  <div style={labelStyle}>E-mail</div>
                  <input type="email" value={email} onChange={(e) => setEmail(e.target.value)}
                    placeholder="aluno@email.com" style={inputStyle} />
                </div>
              </div>

              <div style={{ borderTop: "1px solid #232e42", margin: "8px 0" }} />

              <div style={{ display: "flex", gap: 12 }}>
                <div style={{ flex: 1 }}>
                  <div style={labelStyle}>Plano</div>
                  <select value={planId} onChange={(e) => setPlanId(Number(e.target.value))}
                    style={{ ...inputStyle, cursor: "pointer" }}>
                    <option value={0}>Sem plano</option>
                    {plans.map((pl) => (
                      <option key={pl.id} value={pl.id}>
                        {pl.name} — {formatCents(pl.price_cents)}
                      </option>
                    ))}
                  </select>
                </div>
                <div style={{ flex: 1 }}>
                  <div style={labelStyle}>Forma de Pagamento</div>
                  <select value={formaPagamentoId} onChange={(e) => setFormaPagamentoId(Number(e.target.value))}
                    style={{ ...inputStyle, cursor: "pointer" }}>
                    <option value={0}>Selecione...</option>
                    {methods.map((m) => (
                      <option key={m.id} value={m.id}>{m.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div style={{ width: 120 }}>
                <div style={labelStyle}>Dia de Vencimento</div>
                <input type="number" min={1} max={28} value={dueDay}
                  onChange={(e) => setDueDay(Number(e.target.value))}
                  style={inputStyle} />
              </div>
            </div>

            {error && <div style={{ color: "#f87171", fontSize: 12, marginTop: 12 }}>{error}</div>}

            <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 24 }}>
              <button onClick={() => setShowModal(false)} style={{
                padding: "8px 16px", border: "1px solid #232e42", borderRadius: 6,
                background: "transparent", color: "#7c8798", fontSize: 12, fontFamily: "inherit", cursor: "pointer",
              }}>Cancelar</button>
              <button onClick={handleSave} style={{
                padding: "8px 16px", border: "none", borderRadius: 6,
                background: "#5eead4", color: "#053b32", fontSize: 12, fontWeight: 500,
                fontFamily: "inherit", cursor: "pointer",
              }}>{editId ? "Salvar" : "Criar Aluno"}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Alunos;

import { useState, useEffect } from "react";
import {
  ListarInvoicesPendentes,
  ListarHistoricoPagamentos,
  ListarStudentPlans,
  ListarPaymentMethods,
  ConfirmarPagamento,
  CancelarInvoice,
  DeletarInvoice,
  EstornarPagamento,
} from "../../wailsjs/go/main/App";
import { useToast } from "../components/ToastProvider";

interface Invoice {
  id: number;
  student_plan_id: number;
  student_id: number;
  plan_id: number;
  amount_cents: number;
  status: string;
  payment_method_id?: number;
  due_date: string;
  paid_at?: string;
  paid_amount_cents?: number;
  confirmed_by?: string;
  refunded_by?: string;
  refunded_at?: string;
  notes?: string;
  created_at: string;
  student_name: string;
  plan_name: string;
  payment_method_name: string;
}

interface StudentPlan {
  id: number;
  student_id: number;
  plan_id: number;
  status: string;
  student_name: string;
  plan_name: string;
  plan_price_cents: number;
}

interface PaymentMethod {
  id: number;
  name: string;
  type: string;
  enabled: boolean;
}

interface FinanceiroProps {
  user: { cargo: string };
}

interface PagamentoForm {
  invoice: Invoice;
  valor: string;
  data: string;
  metodoId: number;
}

const formasPreferidas = ["Dinheiro", "Pix", "Cartão de Crédito", "Cartão de Débito", "Transferência"];

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function formatDate(d?: string): string {
  if (!d) return "—";
  const iso = d.length > 10 ? d.slice(0, 10) : d;
  const [ano, mes, dia] = iso.split("-");
  if (ano && mes && dia) return `${dia}/${mes}/${ano}`;
  return d;
}

function hojeISO(): string {
  const agora = new Date();
  const local = new Date(agora.getTime() - agora.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 10);
}

function normalizar(texto: string): string {
  return (texto || "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase();
}

const statusLabels: Record<string, string> = {
  pending: "Pendente",
  paid: "Pago",
  overdue: "Vencida",
  cancelled: "Cancelado",
  refunded: "Reembolsado",
};

const statusBadgeClasses: Record<string, string> = {
  pending: "vmd-badge vmd-badge-warning",
  paid: "vmd-badge vmd-badge-success",
  overdue: "vmd-badge vmd-badge-danger",
  cancelled: "vmd-badge vmd-badge-secondary",
  refunded: "vmd-badge vmd-badge-secondary",
};

function statusEfetivo(inv: Invoice): string {
  if ((inv.status === "pending" || inv.status === "overdue") && inv.due_date && inv.due_date < hojeISO()) {
    return "overdue";
  }
  return inv.status;
}

function Financeiro({ user }: FinanceiroProps) {
  const [pendentes, setPendentes] = useState<Invoice[]>([]);
  const [historico, setHistorico] = useState<Invoice[]>([]);
  const [studentPlans, setStudentPlans] = useState<StudentPlan[]>([]);
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [busca, setBusca] = useState("");
  const [error, setError] = useState("");
  const [tab, setTab] = useState<"invoices" | "history" | "plans">("invoices");
  const [pagamento, setPagamento] = useState<PagamentoForm | null>(null);
  const [salvando, setSalvando] = useState(false);
  const [erroPagamento, setErroPagamento] = useState("");
  const { showToast } = useToast();

  const canManage = user.cargo === "super_admin" || user.cargo === "admin";

  useEffect(() => { load(); }, []);

  async function load() {
    try {
      const [pend, hist, sp, pm] = await Promise.all([
        ListarInvoicesPendentes(),
        ListarHistoricoPagamentos(),
        ListarStudentPlans(),
        ListarPaymentMethods(),
      ]);
      setPendentes((pend as unknown as Invoice[]) || []);
      setHistorico((hist as unknown as Invoice[]) || []);
      setStudentPlans((sp as unknown as StudentPlan[]) || []);
      setMethods((pm as unknown as PaymentMethod[]) || []);
      setError("");
    } catch (e) { setError(String(e)); }
  }

  const formasDisponiveis = (() => {
    const habilitadas = methods.filter((m) => m.enabled);
    const preferidas = habilitadas
      .filter((m) => formasPreferidas.includes(m.name))
      .sort((a, b) => formasPreferidas.indexOf(a.name) - formasPreferidas.indexOf(b.name));
    return preferidas.length > 0 ? preferidas : habilitadas;
  })();

  function abrirPagamento(inv: Invoice) {
    if (!canManage) return;
    const metodoPadrao = formasDisponiveis[0]?.id || 0;
    setErroPagamento("");
    setPagamento({
      invoice: inv,
      valor: (inv.amount_cents / 100).toFixed(2),
      data: hojeISO(),
      metodoId: metodoPadrao,
    });
  }

  async function submitPagamento(e: React.FormEvent) {
    e.preventDefault();
    if (!pagamento) return;
    setErroPagamento("");

    const valorNumerico = Number(pagamento.valor.replace(",", "."));
    if (!Number.isFinite(valorNumerico) || valorNumerico < 0) {
      setErroPagamento("Informe um valor pago válido.");
      return;
    }
    if (!pagamento.metodoId) {
      setErroPagamento("Selecione a forma de pagamento.");
      return;
    }
    if (!pagamento.data) {
      setErroPagamento("Informe a data do pagamento.");
      return;
    }

    setSalvando(true);
    try {
      const cents = Math.round(valorNumerico * 100);
      await ConfirmarPagamento(pagamento.invoice.id, pagamento.metodoId, cents, pagamento.data);
      const aluno = pagamento.invoice.student_name;
      setPagamento(null);
      await load();
      showToast({
        type: "success",
        title: "Pagamento registrado",
        description: `Fatura de ${aluno} confirmada como paga.`,
      });
    } catch (err) {
      setErroPagamento(String(err));
    } finally {
      setSalvando(false);
    }
  }

  async function cancelInvoice(inv: Invoice) {
    if (!canManage) return;
    try {
      await CancelarInvoice(inv.id);
      await load();
      showToast({
        type: "warning",
        title: "Fatura cancelada",
        description: `A fatura de ${inv.student_name} foi cancelada.`,
      });
    } catch (e) {
      showToast({
        type: "error",
        title: "Erro ao cancelar fatura",
        description: String(e),
      });
    }
  }

  async function deleteInvoice(inv: Invoice) {
    if (!canManage) return;
    if (!confirm(`Tem certeza que deseja excluir a fatura de "${inv.student_name}"?`)) return;
    try {
      await DeletarInvoice(inv.id);
      await load();
      showToast({
        type: "success",
        title: "Fatura excluída",
        description: `A fatura de ${inv.student_name} foi removida.`,
      });
    } catch (e) {
      showToast({
        type: "error",
        title: "Erro ao excluir fatura",
        description: String(e),
      });
    }
  }

  async function estornar(inv: Invoice) {
    if (!canManage) return;
    if (!confirm(`Estornar o pagamento de "${inv.student_name}"? A fatura voltará para pendente.`)) return;
    try {
      await EstornarPagamento(inv.id);
      await load();
      showToast({
        type: "warning",
        title: "Pagamento estornado",
        description: `A fatura de ${inv.student_name} voltou para Faturas Emitidas.`,
      });
    } catch (e) {
      showToast({
        type: "error",
        title: "Erro ao estornar pagamento",
        description: String(e),
      });
    }
  }

  const activePlans = studentPlans.filter((sp) => sp.status === "active");
  const termo = normalizar(busca.trim());
  const historicoFiltrado = termo
    ? historico.filter((h) => normalizar(h.student_name).includes(termo))
    : historico;

  const valorFatura = pagamento ? pagamento.invoice.amount_cents : 0;
  const valorInformado = pagamento
    ? Math.round((Number(pagamento.valor.replace(",", ".")) || 0) * 100)
    : 0;
  const desconto = Math.max(valorFatura - valorInformado, 0);

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Gestão Financeira</h2>
      </div>

      <div className="vmd-tabs">
        <button
          onClick={() => setTab("invoices")}
          className={`vmd-tab-btn ${tab === "invoices" ? "active" : ""}`}
        >
          Faturas Emitidas
        </button>
        <button
          onClick={() => setTab("history")}
          className={`vmd-tab-btn ${tab === "history" ? "active" : ""}`}
        >
          Histórico de Pagamentos
        </button>
        {canManage && (
          <button
            onClick={() => setTab("plans")}
            className={`vmd-tab-btn ${tab === "plans" ? "active" : ""}`}
          >
            Assinaturas Ativas
          </button>
        )}
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 20 }}>
          <span>⚠️ {error}</span>
        </div>
      )}

      {tab === "invoices" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {pendentes.map((inv) => {
            const status = statusEfetivo(inv);
            return (
              <div
                key={inv.id}
                className="vmd-card"
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  flexWrap: "wrap",
                  gap: 16,
                  padding: "16px 20px",
                }}
              >
                <div style={{ flex: 1, minWidth: 220 }}>
                  <div style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)" }}>
                    {inv.student_name}
                  </div>
                  <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
                    Plano: <strong style={{ color: "var(--text-main)" }}>{inv.plan_name}</strong> · Valor: {formatCents(inv.amount_cents)} · Vencimento: {formatDate(inv.due_date)}
                  </div>
                  {inv.refunded_by && (
                    <div style={{ fontSize: 12, color: "var(--warning, #eab308)", marginTop: 4 }}>
                      ↩ Estornada por {inv.refunded_by}
                      {inv.refunded_at ? ` em ${formatDate(inv.refunded_at)}` : ""}
                    </div>
                  )}
                </div>

                <div style={{ display: "flex", alignItems: "center", gap: 8, flexShrink: 0 }}>
                  <span className={statusBadgeClasses[status] || "vmd-badge"}>
                    {statusLabels[status] || status}
                  </span>

                  {canManage && (
                    <>
                      <button
                        onClick={() => abrirPagamento(inv)}
                        className="vmd-btn vmd-btn-success"
                        style={{ padding: "6px 12px", height: 32, fontSize: 11 }}
                      >
                        Confirmar Pago
                      </button>
                      <button
                        onClick={() => cancelInvoice(inv)}
                        className="vmd-btn vmd-btn-secondary"
                        style={{ padding: "6px 12px", height: 32, fontSize: 11 }}
                      >
                        Cancelar
                      </button>
                      <button
                        onClick={() => deleteInvoice(inv)}
                        className="vmd-btn vmd-btn-danger"
                        style={{ padding: "6px 12px", height: 32, fontSize: 11, background: "transparent" }}
                      >
                        Excluir
                      </button>
                    </>
                  )}
                </div>
              </div>
            );
          })}
          {pendentes.length === 0 && (
            <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
              Nenhuma fatura pendente ou vencida.
            </div>
          )}
        </div>
      )}

      {tab === "history" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          <div className="vmd-card" style={{ padding: "12px 16px" }}>
            <input
              className="vmd-input"
              placeholder="Buscar pelo nome do aluno..."
              value={busca}
              onChange={(e) => setBusca(e.target.value)}
            />
          </div>

          {historicoFiltrado.map((inv) => (
            <div
              key={inv.id}
              className="vmd-card"
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                flexWrap: "wrap",
                gap: 16,
                padding: "16px 20px",
              }}
            >
              <div style={{ flex: 1, minWidth: 220 }}>
                <div style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)" }}>
                  {inv.student_name}
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
                  Plano: <strong style={{ color: "var(--text-main)" }}>{inv.plan_name}</strong> · Pago: <strong style={{ color: "var(--success)" }}>{formatCents(inv.paid_amount_cents ?? inv.amount_cents)}</strong> · Data: {formatDate(inv.paid_at)}
                  {inv.payment_method_name && ` · Forma: ${inv.payment_method_name}`}
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
                  Confirmado por: {inv.confirmed_by || "—"}
                </div>
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: 8, flexShrink: 0 }}>
                <span className={statusBadgeClasses.paid}>
                  {statusLabels.paid}
                </span>
                {canManage && (
                  <button
                    onClick={() => estornar(inv)}
                    className="vmd-btn vmd-btn-danger"
                    style={{ padding: "6px 12px", height: 32, fontSize: 11, background: "transparent" }}
                  >
                    Estornar
                  </button>
                )}
              </div>
            </div>
          ))}
          {historicoFiltrado.length === 0 && (
            <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
              {historico.length === 0 ? "Nenhum pagamento registrado." : "Nenhum pagamento encontrado para esta busca."}
            </div>
          )}
        </div>
      )}

      {tab === "plans" && canManage && (
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {activePlans.map((sp) => (
            <div
              key={sp.id}
              className="vmd-card"
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                padding: "16px 20px",
              }}
            >
              <div>
                <div style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)" }}>
                  {sp.student_name}
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
                  Plano: <strong style={{ color: "var(--text-main)" }}>{sp.plan_name}</strong> · Mensalidade: {formatCents(sp.plan_price_cents)}
                </div>
              </div>
              <div>
                <span className="vmd-badge vmd-badge-success">
                  {sp.status === "active" ? "Regular" : sp.status}
                </span>
              </div>
            </div>
          ))}
          {activePlans.length === 0 && (
            <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
              Nenhum plano ativo de aluno registrado.
            </div>
          )}
        </div>
      )}

      {pagamento && (
        <div className="vmd-modal-overlay" onClick={() => !salvando && setPagamento(null)}>
          <form
            onSubmit={submitPagamento}
            className="vmd-modal-content"
            style={{ maxWidth: 440 }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="vmd-modal-header">
              <h3 style={{ fontSize: 16, margin: 0 }}>Confirmar Pagamento</h3>
              <button type="button" className="vmd-modal-close" onClick={() => setPagamento(null)} disabled={salvando}>✕</button>
            </div>

            <div style={{ fontSize: 13, color: "var(--text-muted)", marginTop: 12 }}>
              <div><strong style={{ color: "var(--text-main)" }}>{pagamento.invoice.student_name}</strong></div>
              <div style={{ marginTop: 2 }}>Plano: {pagamento.invoice.plan_name}</div>
              <div style={{ marginTop: 2 }}>Valor da fatura: {formatCents(pagamento.invoice.amount_cents)}</div>
            </div>

            <div className="vmd-form-group" style={{ marginTop: 16 }}>
              <label className="vmd-label">Valor pago (R$)</label>
              <input
                className="vmd-input"
                type="number"
                min="0"
                step="0.01"
                value={pagamento.valor}
                onChange={(e) => setPagamento({ ...pagamento, valor: e.target.value })}
                required
              />
              {desconto > 0 && (
                <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
                  Desconto de {formatCents(desconto)} aplicado sobre o valor da fatura.
                </div>
              )}
            </div>

            <div className="vmd-form-group">
              <label className="vmd-label">Data do pagamento</label>
              <input
                className="vmd-input"
                type="date"
                value={pagamento.data}
                onChange={(e) => setPagamento({ ...pagamento, data: e.target.value })}
                required
              />
            </div>

            <div className="vmd-form-group">
              <label className="vmd-label">Forma de pagamento</label>
              <select
                className="vmd-select"
                value={pagamento.metodoId}
                onChange={(e) => setPagamento({ ...pagamento, metodoId: Number(e.target.value) })}
                required
              >
                <option value={0} disabled>Selecione...</option>
                {formasDisponiveis.map((m) => (
                  <option key={m.id} value={m.id}>{m.name}</option>
                ))}
              </select>
            </div>

            {erroPagamento && <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 12 }}>⚠️ {erroPagamento}</div>}

            <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
              <button type="button" className="vmd-btn vmd-btn-secondary" onClick={() => setPagamento(null)} disabled={salvando}>
                Cancelar
              </button>
              <button type="submit" className="vmd-btn vmd-btn-success" disabled={salvando}>
                {salvando ? "Confirmando..." : "Confirmar Pagamento"}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}

export default Financeiro;

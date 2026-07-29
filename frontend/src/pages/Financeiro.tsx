import { useState, useEffect } from "react";
import {
  ListarInvoices,
  ListarStudentPlans,
  ConfirmarPagamento,
  CancelarInvoice,
  DeletarInvoice,
} from "../../wailsjs/go/main/App";

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
  gateway_transaction_id?: string;
  pix_qr_code?: string;
  pix_br_code?: string;
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

interface FinanceiroProps {
  user: { cargo: string };
}

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function formatDate(d: string): string {
  if (!d) return "—";
  return d.replace("T", " ").split(" ")[0];
}

const statusLabels: Record<string, string> = {
  pending: "Pendente",
  paid: "Pago",
  overdue: "Vencido",
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

function Financeiro({ user }: FinanceiroProps) {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [studentPlans, setStudentPlans] = useState<StudentPlan[]>([]);
  const [error, setError] = useState("");
  const [tab, setTab] = useState<"invoices" | "plans">("invoices");

  const canManage = user.cargo === "super_admin" || user.cargo === "admin";

  useEffect(() => { load(); }, []);

  async function load() {
    try {
      const [inv, sp] = await Promise.all([
        ListarInvoices(),
        ListarStudentPlans(),
      ]);
      setInvoices(inv as unknown as Invoice[]);
      setStudentPlans(sp as unknown as StudentPlan[]);
    } catch (e) { setError(String(e)); }
  }

  async function confirmPayment(inv: Invoice) {
    if (!canManage) return;
    try {
      await ConfirmarPagamento(inv.id, 1, inv.amount_cents);
      await load();
    } catch (e) { setError(String(e)); }
  }

  async function cancelInvoice(inv: Invoice) {
    if (!canManage) return;
    try {
      await CancelarInvoice(inv.id);
      await load();
    } catch (e) { setError(String(e)); }
  }

  async function deleteInvoice(inv: Invoice) {
    if (!canManage) return;
    try {
      await DeletarInvoice(inv.id);
      await load();
    } catch (e) { setError(String(e)); }
  }

  const activePlans = studentPlans.filter((sp) => sp.status === "active");

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Gestão Financeira</h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Gerenciamento de cobranças, faturas e planos dos alunos matriculados.
        </p>
      </div>

      <div className="vmd-tabs">
        <button
          onClick={() => setTab("invoices")}
          className={`vmd-tab-btn ${tab === "invoices" ? "active" : ""}`}
        >
          Faturas Emitidas
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
          {invoices.map((inv) => (
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
                  Plano: <strong style={{ color: "var(--text-main)" }}>{inv.plan_name}</strong> · Valor: {formatCents(inv.amount_cents)} · Vencimento: {inv.due_date}
                </div>
                {inv.paid_at && (
                  <div style={{ fontSize: 12, color: "var(--success)", fontWeight: 500, marginTop: 4 }}>
                    ✓ Pago em {formatDate(inv.paid_at)} {inv.payment_method_name && `via ${inv.payment_method_name}`}
                  </div>
                )}
              </div>
              
              <div style={{ display: "flex", alignItems: "center", gap: 8, flexShrink: 0 }}>
                <span className={statusBadgeClasses[inv.status] || "vmd-badge"}>
                  {statusLabels[inv.status] || inv.status}
                </span>

                {canManage && inv.status === "pending" && (
                  <>
                    <button
                      onClick={() => confirmPayment(inv)}
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

                {canManage && inv.status !== "pending" && (
                  <button
                    onClick={() => deleteInvoice(inv)}
                    className="vmd-btn vmd-btn-danger"
                    style={{ padding: "6px 12px", height: 32, fontSize: 11, background: "transparent" }}
                  >
                    Excluir
                  </button>
                )}
              </div>
            </div>
          ))}
          {invoices.length === 0 && (
            <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
              Nenhuma fatura encontrada.
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
    </div>
  );
}

export default Financeiro;

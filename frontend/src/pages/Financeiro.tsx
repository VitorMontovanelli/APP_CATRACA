import { useState, useEffect } from "react";
import {
  ListarInvoices,
  ListarStudentPlans,
  ConfirmarPagamento,
  CancelarInvoice,
  DeletarInvoice,
  GerarInvoice,
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

const statusColors: Record<string, string> = {
  pending: "#fbbf24",
  paid: "#5eead4",
  overdue: "#f87171",
  cancelled: "#7c8798",
  refunded: "#60a5fa",
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

  async function generateInvoice(spId: number) {
    if (!canManage) return;
    try {
      await GerarInvoice(spId);
      await load();
    } catch (e) { setError(String(e)); }
  }

  const activePlans = studentPlans.filter((sp) => sp.status === "active");

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <h2 style={{ margin: "0 0 16px 0", fontSize: 20, fontWeight: 500 }}>Financeiro</h2>
        <div style={{ display: "flex", gap: 8 }}>
          <button onClick={() => setTab("invoices")} style={{
            padding: "7px 14px", border: "none", borderRadius: 6,
            background: tab === "invoices" ? "#5eead4" : "#121a29",
            color: tab === "invoices" ? "#053b32" : "#a7b0bf",
            fontSize: 12, fontFamily: "inherit", cursor: "pointer",
          }}>Faturas</button>
          {canManage && (
            <button onClick={() => setTab("plans")} style={{
              padding: "7px 14px", border: "none", borderRadius: 6,
              background: tab === "plans" ? "#5eead4" : "#121a29",
              color: tab === "plans" ? "#053b32" : "#a7b0bf",
              fontSize: 12, fontFamily: "inherit", cursor: "pointer",
            }}>Planos dos Alunos</button>
          )}
        </div>
      </div>

      {error && <div style={{ color: "#f87171", fontSize: 12, marginBottom: 12 }}>{error}</div>}

      {tab === "invoices" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {invoices.map((inv) => (
            <div key={inv.id} style={{
              background: "#121a29", border: "1px solid #232e42", borderRadius: 8,
              padding: "14px 18px", display: "flex", alignItems: "center",
              justifyContent: "space-between", gap: 12,
            }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, fontWeight: 500 }}>{inv.student_name}</div>
                <div style={{ fontSize: 12, color: "#7c8798" }}>
                  {inv.plan_name} · {formatCents(inv.amount_cents)} · Venc: {inv.due_date}
                </div>
                {inv.paid_at && (
                  <div style={{ fontSize: 12, color: "#5eead4" }}>
                    Pago em {formatDate(inv.paid_at)} {inv.payment_method_name && `via ${inv.payment_method_name}`}
                  </div>
                )}
              </div>
              <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                <span style={{
                  padding: "3px 10px", borderRadius: 4, fontSize: 11, fontWeight: 500,
                  background: (statusColors[inv.status] || "#7c8798") + "22",
                  color: statusColors[inv.status] || "#7c8798",
                }}>
                  {statusLabels[inv.status] || inv.status}
                </span>
                {canManage && inv.status === "pending" && (
                  <>
                    <button onClick={() => confirmPayment(inv)} style={{
                      padding: "6px 12px", border: "none", borderRadius: 6,
                      background: "#5eead4", color: "#053b32", fontSize: 12,
                      fontFamily: "inherit", cursor: "pointer",
                    }}>Confirmar</button>
                    <button onClick={() => cancelInvoice(inv)} style={{
                      padding: "6px 12px", border: "1px solid #f87171", borderRadius: 6,
                      background: "transparent", color: "#f87171", fontSize: 12,
                      fontFamily: "inherit", cursor: "pointer",
                    }}>Cancelar</button>
                    <button onClick={() => deleteInvoice(inv)} style={{
                      padding: "6px 12px", border: "1px solid #f87171", borderRadius: 6,
                      background: "#f8717122", color: "#f87171", fontSize: 12,
                      fontFamily: "inherit", cursor: "pointer",
                    }}>Excluir</button>
                  </>
                )}
                {canManage && inv.status !== "pending" && (
                  <button onClick={() => deleteInvoice(inv)} style={{
                    padding: "6px 12px", border: "1px solid #f87171", borderRadius: 6,
                    background: "#f8717122", color: "#f87171", fontSize: 12,
                    fontFamily: "inherit", cursor: "pointer",
                  }}>Excluir</button>
                )}
              </div>
            </div>
          ))}
          {invoices.length === 0 && (
            <div style={{ color: "#7c8798", fontSize: 13, textAlign: "center", padding: 40 }}>
              Nenhuma fatura encontrada.
            </div>
          )}
        </div>
      )}

      {tab === "plans" && canManage && (
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {activePlans.map((sp) => (
            <div key={sp.id} style={{
              background: "#121a29", border: "1px solid #232e42", borderRadius: 8,
              padding: "14px 18px", display: "flex", alignItems: "center",
              justifyContent: "space-between", gap: 12,
            }}>
              <div>
                <div style={{ fontSize: 14, fontWeight: 500 }}>{sp.student_name}</div>
                <div style={{ fontSize: 12, color: "#7c8798" }}>
                  {sp.plan_name} · {formatCents(sp.plan_price_cents)} · {sp.status}
                </div>
              </div>
              <button onClick={() => generateInvoice(sp.id)} style={{
                padding: "6px 12px", border: "1px solid #60a5fa", borderRadius: 6,
                background: "transparent", color: "#60a5fa", fontSize: 12,
                fontFamily: "inherit", cursor: "pointer",
              }}>Gerar Fatura</button>
            </div>
          ))}
          {activePlans.length === 0 && (
            <div style={{ color: "#7c8798", fontSize: 13, textAlign: "center", padding: 40 }}>
              Nenhum plano de aluno ativo.
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default Financeiro;

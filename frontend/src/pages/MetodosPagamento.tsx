import { useState, useEffect } from "react";
import {
  ListarPaymentMethods,
  CriarPaymentMethod,
  AtualizarPaymentMethod,
  AtivarPaymentMethod,
  DeletarPaymentMethod,
} from "../../wailsjs/go/main/App";

interface PaymentMethod {
  id: number;
  name: string;
  type: string;
  enabled: boolean;
  fee_percent: number;
  fee_fixed_cents: number;
  created_at: string;
}

interface MetodosPagamentoProps {
  user: { cargo: string };
}

type FeeMode = "none" | "percent" | "fixed";

const typeOptions = [
  { value: "cash", label: "Dinheiro" },
  { value: "pix", label: "Pix" },
  { value: "credit_card", label: "Cartão de Crédito" },
  { value: "debit_card", label: "Cartão de Débito" },
  { value: "boleto", label: "Boleto" },
  { value: "transfer", label: "Transferência" },
  { value: "wallet", label: "Carteira Digital" },
  { value: "other", label: "Outro" },
];

const typeLabels: Record<string, string> = Object.fromEntries(
  typeOptions.map((t) => [t.value, t.label])
);

const methodIcons: Record<string, string> = {
  cash: "💵",
  pix: "⚡",
  credit_card: "💳",
  debit_card: "💳",
  boleto: "📄",
  transfer: "🏦",
  wallet: "📱",
  other: "🧾",
};

function formatFee(m: PaymentMethod): string {
  if (m.fee_percent > 0) {
    return `${m.fee_percent.toLocaleString("pt-BR", { maximumFractionDigits: 2 })}%`;
  }
  if (m.fee_fixed_cents > 0) {
    return `R$ ${(m.fee_fixed_cents / 100).toLocaleString("pt-BR", { minimumFractionDigits: 2 })}`;
  }
  return "";
}

function MetodosPagamento({ user }: MetodosPagamentoProps) {
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [error, setError] = useState("");
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<PaymentMethod | null>(null);
  const [name, setName] = useState("");
  const [mtype, setMtype] = useState("pix");
  const [feeMode, setFeeMode] = useState<FeeMode>("none");
  const [feeValue, setFeeValue] = useState("");
  const [saving, setSaving] = useState(false);

  const canManage = user.cargo === "admin" || user.cargo === "super_admin";

  useEffect(() => { load(); }, []);

  async function load() {
    try {
      const list = await ListarPaymentMethods();
      setMethods((list as unknown as PaymentMethod[]) || []);
      setError("");
    } catch (e) { setError(String(e)); }
  }

  function abrirNovo() {
    setEditing(null);
    setName("");
    setMtype("pix");
    setFeeMode("none");
    setFeeValue("");
    setError("");
    setShowModal(true);
  }

  function abrirEdicao(m: PaymentMethod) {
    setEditing(m);
    setName(m.name);
    setMtype(m.type || "other");
    if (m.fee_percent > 0) {
      setFeeMode("percent");
      setFeeValue(String(m.fee_percent).replace(".", ","));
    } else if (m.fee_fixed_cents > 0) {
      setFeeMode("fixed");
      setFeeValue((m.fee_fixed_cents / 100).toFixed(2).replace(".", ","));
    } else {
      setFeeMode("none");
      setFeeValue("");
    }
    setError("");
    setShowModal(true);
  }

  function fecharModal() {
    if (saving) return;
    setShowModal(false);
    setEditing(null);
    setError("");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    const nome = name.trim();
    if (!nome) {
      setError("Informe o nome do método de pagamento.");
      return;
    }

    let feePercent = 0;
    let feeFixedCents = 0;
    if (feeMode !== "none") {
      const numero = Number(feeValue.replace(",", "."));
      if (!Number.isFinite(numero) || numero < 0) {
        setError("Informe uma taxa válida.");
        return;
      }
      if (feeMode === "percent") {
        feePercent = numero;
      } else {
        feeFixedCents = Math.round(numero * 100);
      }
    }

    setSaving(true);
    try {
      if (editing) {
        await AtualizarPaymentMethod(editing.id, nome, mtype, feePercent, feeFixedCents);
      } else {
        await CriarPaymentMethod(nome, mtype, feePercent, feeFixedCents);
      }
      setShowModal(false);
      setEditing(null);
      await load();
    } catch (err) {
      setError(String(err));
    } finally {
      setSaving(false);
    }
  }

  async function toggleMethod(m: PaymentMethod) {
    try {
      await AtivarPaymentMethod(m.id, !m.enabled);
      await load();
    } catch (err) { setError(String(err)); }
  }

  async function deleteMethod(m: PaymentMethod) {
    if (!confirm(`Excluir o método "${m.name}"? Métodos já usados não podem ser excluídos.`)) return;
    try {
      await DeletarPaymentMethod(m.id);
      await load();
    } catch (err) { setError(String(err)); }
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <div>
          <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Métodos de Pagamento</h2>
        </div>
        {canManage && (
          <button onClick={abrirNovo} className="vmd-btn vmd-btn-primary">
            ➕ Adicionar método
          </button>
        )}
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 16 }}>
          <span>⚠️ {error}</span>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {methods.map((m) => {
          const fee = formatFee(m);
          return (
            <div
              key={m.id}
              className="vmd-card"
              style={{
                padding: "16px 20px",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                flexWrap: "wrap",
                gap: 12,
                opacity: m.enabled ? 1 : 0.6,
                background: m.enabled ? "var(--bg-card)" : "rgba(255,255,255,0.01)",
              }}
            >
              <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
                <span style={{ fontSize: 24, width: 32, height: 32, display: "flex", alignItems: "center", justifyContent: "center", background: "rgba(255,255,255,0.03)", borderRadius: "50%", border: "1px solid var(--border-color)" }}>
                  {methodIcons[m.type] || "💳"}
                </span>
                <div>
                  <div style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)", display: "flex", alignItems: "center", gap: 8 }}>
                    {m.name}
                    <span className={`vmd-badge ${m.enabled ? "vmd-badge-success" : "vmd-badge-secondary"}`}>
                      {m.enabled ? "Ativo" : "Inativo"}
                    </span>
                  </div>
                  <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>
                    Tipo: {typeLabels[m.type] || m.type}
                    {fee ? ` · taxa: ${fee}` : " · sem taxa"}
                  </div>
                </div>
              </div>

              {canManage && (
                <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
                  <button
                    onClick={() => abrirEdicao(m)}
                    className="vmd-btn vmd-btn-secondary"
                    style={{ padding: "6px 12px", height: 32 }}
                  >
                    Editar
                  </button>
                  <button
                    onClick={() => toggleMethod(m)}
                    className={`vmd-btn ${m.enabled ? "vmd-btn-danger" : "vmd-btn-success"}`}
                    style={{ padding: "6px 12px", height: 32 }}
                  >
                    {m.enabled ? "Desativar" : "Ativar"}
                  </button>
                  <button
                    onClick={() => deleteMethod(m)}
                    className="vmd-btn vmd-btn-danger"
                    style={{ padding: "6px 12px", height: 32, background: "transparent" }}
                  >
                    Excluir
                  </button>
                </div>
              )}
            </div>
          );
        })}
        {methods.length === 0 && (
          <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
            Nenhum método de pagamento cadastrado.
          </div>
        )}
      </div>

      {showModal && (
        <div className="vmd-modal-overlay" onClick={fecharModal}>
          <form
            onSubmit={handleSubmit}
            className="vmd-modal-content"
            style={{ maxWidth: 460 }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="vmd-modal-header">
              <h3 style={{ fontSize: 16, margin: 0 }}>
                {editing ? "Editar Método de Pagamento" : "Adicionar Método de Pagamento"}
              </h3>
              <button
                type="button"
                onClick={fecharModal}
                className="vmd-btn vmd-btn-ghost"
                style={{ padding: 4, minWidth: "auto" }}
                disabled={saving}
              >
                ✕
              </button>
            </div>

            <div className="vmd-modal-body">
              <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Nome *</label>
                  <input
                    placeholder="Ex: Pix, Cartão da loja..."
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    required
                    className="vmd-input"
                  />
                </div>

                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Tipo *</label>
                  <select
                    value={mtype}
                    onChange={(e) => setMtype(e.target.value)}
                    required
                    className="vmd-select"
                  >
                    {typeOptions.map((t) => (
                      <option key={t.value} value={t.value}>{t.label}</option>
                    ))}
                  </select>
                </div>

                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Taxa (opcional)</label>
                  <select
                    value={feeMode}
                    onChange={(e) => {
                      setFeeMode(e.target.value as FeeMode);
                      setFeeValue("");
                    }}
                    className="vmd-select"
                  >
                    <option value="none">Sem taxa</option>
                    <option value="percent">Percentual (%)</option>
                    <option value="fixed">Valor fixo (R$)</option>
                  </select>
                </div>

                {feeMode !== "none" && (
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">
                      {feeMode === "percent" ? "Percentual da taxa" : "Valor da taxa (R$)"}
                    </label>
                    <input
                      type="text"
                      inputMode="decimal"
                      placeholder={feeMode === "percent" ? "Ex: 3,5" : "Ex: 3,49"}
                      value={feeValue}
                      onChange={(e) => setFeeValue(e.target.value)}
                      className="vmd-input"
                      required
                    />
                  </div>
                )}

                {error && <div style={{ color: "var(--danger)", fontSize: 12 }}>{error}</div>}
              </div>
            </div>

            <div className="vmd-modal-footer">
              <button
                type="button"
                onClick={fecharModal}
                className="vmd-btn vmd-btn-secondary"
                disabled={saving}
              >
                Cancelar
              </button>
              <button type="submit" className="vmd-btn vmd-btn-primary" disabled={saving}>
                {saving ? "Salvando..." : "Salvar"}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}

export default MetodosPagamento;

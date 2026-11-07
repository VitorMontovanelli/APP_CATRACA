import { useEffect, useState } from "react";
import { ConfirmarPagamento, ListarPaymentMethods } from "../../wailsjs/go/main/App";
import { useToast } from "./ToastProvider";

export interface FaturaParaPagamento {
  id: number;
  student_name: string;
  plan_name: string;
  amount_cents: number;
}

interface PaymentMethod {
  id: number;
  name: string;
  type: string;
  enabled: boolean;
}

interface PagamentoModalProps {
  invoice: FaturaParaPagamento;
  onClose: () => void;
  onConfirmed: () => void;
}

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function hojeISO(): string {
  const agora = new Date();
  const local = new Date(agora.getTime() - agora.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 10);
}

/**
 * Fluxo único de confirmação de pagamento, compartilhado entre a Gestão
 * Financeira e o Painel de Cobranças.
 */
export default function PagamentoModal({ invoice, onClose, onConfirmed }: PagamentoModalProps) {
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [valor, setValor] = useState((invoice.amount_cents / 100).toFixed(2));
  const [data, setData] = useState(hojeISO());
  const [metodoId, setMetodoId] = useState(0);
  const [salvando, setSalvando] = useState(false);
  const [erro, setErro] = useState("");
  const { showToast } = useToast();

  useEffect(() => {
    let ativo = true;
    ListarPaymentMethods()
      .then((lista) => {
        if (!ativo) return;
        const disponiveis = ((lista as unknown as PaymentMethod[]) || []).filter((m) => m.enabled);
        setMethods(disponiveis);
        setMetodoId(disponiveis[0]?.id || 0);
      })
      .catch((e) => {
        if (ativo) setErro(String(e));
      });
    return () => {
      ativo = false;
    };
  }, []);

  const valorInformado = Math.round((Number(valor.replace(",", ".")) || 0) * 100);
  const desconto = Math.max(invoice.amount_cents - valorInformado, 0);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErro("");

    const valorNumerico = Number(valor.replace(",", "."));
    if (!Number.isFinite(valorNumerico) || valorNumerico < 0) {
      setErro("Informe um valor pago válido.");
      return;
    }
    if (!metodoId) {
      setErro("Selecione a forma de pagamento.");
      return;
    }
    if (!data) {
      setErro("Informe a data do pagamento.");
      return;
    }

    setSalvando(true);
    try {
      await ConfirmarPagamento(invoice.id, metodoId, Math.round(valorNumerico * 100), data);
      showToast({
        type: "success",
        title: "Pagamento registrado",
        description: `Fatura de ${invoice.student_name} confirmada como paga.`,
      });
      onConfirmed();
      onClose();
    } catch (err) {
      setErro(String(err));
    } finally {
      setSalvando(false);
    }
  }

  return (
    <div className="vmd-modal-overlay" onClick={() => !salvando && onClose()}>
      <form
        onSubmit={submit}
        className="vmd-modal-content"
        style={{ maxWidth: 440 }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="vmd-modal-header">
          <h3 style={{ fontSize: 16, margin: 0 }}>Confirmar Pagamento</h3>
          <button type="button" className="vmd-modal-close" onClick={onClose} disabled={salvando}>✕</button>
        </div>

        <div style={{ fontSize: 13, color: "var(--text-muted)", marginTop: 12 }}>
          <div><strong style={{ color: "var(--text-main)" }}>{invoice.student_name}</strong></div>
          <div style={{ marginTop: 2 }}>Plano: {invoice.plan_name}</div>
          <div style={{ marginTop: 2 }}>Valor da fatura: {formatCents(invoice.amount_cents)}</div>
        </div>

        <div className="vmd-form-group" style={{ marginTop: 16 }}>
          <label className="vmd-label">Valor pago (R$)</label>
          <input
            className="vmd-input"
            type="number"
            min="0"
            step="0.01"
            value={valor}
            onChange={(e) => setValor(e.target.value)}
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
            value={data}
            onChange={(e) => setData(e.target.value)}
            required
          />
        </div>

        <div className="vmd-form-group">
          <label className="vmd-label">Forma de pagamento</label>
          <select
            className="vmd-select"
            value={metodoId}
            onChange={(e) => setMetodoId(Number(e.target.value))}
            required
          >
            <option value={0} disabled>Selecione...</option>
            {methods.map((m) => (
              <option key={m.id} value={m.id}>{m.name}</option>
            ))}
          </select>
        </div>

        {methods.length === 0 && (
          <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 12 }}>
            Nenhuma forma de pagamento ativa. Ative ou cadastre uma em "Pagamentos".
          </div>
        )}

        {erro && <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 12 }}>⚠️ {erro}</div>}

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <button type="button" className="vmd-btn vmd-btn-secondary" onClick={onClose} disabled={salvando}>
            Cancelar
          </button>
          <button type="submit" className="vmd-btn vmd-btn-success" disabled={salvando || methods.length === 0}>
            {salvando ? "Confirmando..." : "Confirmar Pagamento"}
          </button>
        </div>
      </form>
    </div>
  );
}

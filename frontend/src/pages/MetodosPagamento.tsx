import { useState, useEffect } from "react";
import {
  ListarPaymentMethods,
  AtivarPaymentMethod,
  ListarPaymentGateways,
  AtivarPaymentGateway,
  AtualizarConfigGateway,
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

interface PaymentGateway {
  id: number;
  name: string;
  type: string;
  enabled: boolean;
  config?: string;
  created_at: string;
}

interface MetodosPagamentoProps {
  user: { cargo: string };
}

const typeLabels: Record<string, string> = {
  pix: "Pix",
  credit_card: "Cartão de Crédito",
  boleto: "Boleto bancário",
  debit_card: "Cartão de Débito",
  wallet: "Carteira Digital",
  cash: "Dinheiro espécie",
};

const methodIcons: Record<string, string> = {
  pix: "⚡",
  credit_card: "💳",
  boleto: "📄",
  debit_card: "💳",
  wallet: "📱",
  cash: "💵",
};

function MetodosPagamento({ user }: MetodosPagamentoProps) {
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [gateways, setGateways] = useState<PaymentGateway[]>([]);
  const [error, setError] = useState("");
  const [editingGateway, setEditingGateway] = useState<number | null>(null);
  const [gatewayConfig, setGatewayConfig] = useState("");

  const isSuper = user.cargo === "super_admin";

  useEffect(() => { load(); }, []);

  async function load() {
    try {
      const [m, g] = await Promise.all([
        ListarPaymentMethods(),
        ListarPaymentGateways(),
      ]);
      setMethods((m as unknown as PaymentMethod[]) || []);
      setGateways((g as unknown as PaymentGateway[]) || []);
    } catch (e) { setError(String(e)); }
  }

  async function toggleMethod(m: PaymentMethod) {
    try {
      await AtivarPaymentMethod(m.id, !m.enabled);
      await load();
    } catch (err) { setError(String(err)); }
  }

  async function toggleGateway(g: PaymentGateway) {
    try {
      await AtivarPaymentGateway(g.id, !g.enabled);
      await load();
    } catch (err) { setError(String(err)); }
  }

  async function saveGatewayConfig(g: PaymentGateway) {
    try {
      await AtualizarConfigGateway(g.id, gatewayConfig);
      setEditingGateway(null);
      await load();
    } catch (err) { setError(String(err)); }
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Métodos de Pagamento</h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Habilite e configure as formas de recebimento e integradores de faturas da academia.
        </p>
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 16 }}>
          <span>⚠️ {error}</span>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 12, marginBottom: 32 }}>
        {methods.map((m) => (
          <div
            key={m.id}
            className="vmd-card"
            style={{
              padding: "16px 20px",
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              opacity: m.enabled ? 1 : 0.6,
              background: m.enabled ? "var(--bg-card)" : "rgba(255,255,255,0.01)",
            }}
          >
            <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
              <span style={{ fontSize: 24, width: 32, height: 32, display: "flex", alignItems: "center", justifyContent: "center", background: "rgba(255,255,255,0.03)", borderRadius: "50%", border: "1px solid var(--border-color)" }}>
                {methodIcons[m.type] || "💳"}
              </span>
              <div>
                <div style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)" }}>
                  {m.name}
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>
                  Tipo: {typeLabels[m.type] || m.type}
                  {m.fee_percent > 0 && ` · taxa: ${m.fee_percent}%`}
                  {m.fee_fixed_cents > 0 && ` + R$ ${(m.fee_fixed_cents / 100).toFixed(2)}`}
                </div>
              </div>
            </div>
            {isSuper && (
              <button
                onClick={() => toggleMethod(m)}
                className={`vmd-btn ${m.enabled ? "vmd-btn-danger" : "vmd-btn-success"}`}
                style={{ padding: "6px 12px", height: 32 }}
              >
                {m.enabled ? "Desativar" : "Ativar"}
              </button>
            )}
          </div>
        ))}
      </div>

      <div style={{ marginBottom: 16 }}>
        <h3 style={{ fontSize: 18, fontWeight: 600, margin: 0 }}>Gateways de Integração</h3>
        <p style={{ color: "var(--text-muted)", fontSize: 12, marginTop: 4 }}>
          Provedores de serviços de pagamento e liquidação automatizada.
        </p>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {gateways.map((g) => (
          <div
            key={g.id}
            className="vmd-card"
            style={{
              padding: "16px 20px",
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              opacity: g.enabled ? 1 : 0.6,
              background: g.enabled ? "var(--bg-card)" : "rgba(255,255,255,0.01)",
            }}
          >
            {editingGateway === g.id ? (
              <div style={{ display: "flex", alignItems: "center", gap: 12, flex: 1, flexWrap: "wrap", width: "100%" }}>
                <div style={{ flex: 1, minWidth: 200 }}>
                  <label className="vmd-label">Parâmetros de Configuração (JSON)</label>
                  <input
                    value={gatewayConfig}
                    onChange={(e) => setGatewayConfig(e.target.value)}
                    placeholder='{"api_key":"...", "sandbox":true}'
                    className="vmd-input"
                    style={{ fontFamily: "monospace" }}
                  />
                </div>
                <div style={{ display: "flex", gap: 8, alignSelf: "flex-end", marginTop: 12 }}>
                  <button onClick={() => saveGatewayConfig(g)} className="vmd-btn vmd-btn-success" style={{ padding: "6px 12px", height: 32 }}>
                    Salvar
                  </button>
                  <button onClick={() => setEditingGateway(null)} className="vmd-btn vmd-btn-secondary" style={{ padding: "6px 12px", height: 32 }}>
                    Cancelar
                  </button>
                </div>
              </div>
            ) : (
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", width: "100%", flexWrap: "wrap", gap: 12 }}>
                <div>
                  <div style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)", display: "flex", alignItems: "center", gap: 8 }}>
                    {g.name}
                    <span className={`vmd-badge ${g.enabled ? "vmd-badge-success" : "vmd-badge-secondary"}`}>
                      {g.enabled ? "Ativo" : "Inativo"}
                    </span>
                  </div>
                  <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
                    Provedor: {g.type} {g.config ? "· ✓ Parâmetros configurados" : "· ⚠ Não configurado"}
                  </div>
                </div>
                
                <div style={{ display: "flex", gap: 8 }}>
                  {isSuper && (
                    <>
                      <button
                        onClick={() => {
                          setEditingGateway(g.id);
                          setGatewayConfig(g.config || "");
                        }}
                        className="vmd-btn vmd-btn-secondary"
                        style={{ padding: "6px 12px", height: 32 }}
                      >
                        Configurar
                      </button>
                      <button
                        onClick={() => toggleGateway(g)}
                        className={`vmd-btn ${g.enabled ? "vmd-btn-danger" : "vmd-btn-success"}`}
                        style={{ padding: "6px 12px", height: 32 }}
                      >
                        {g.enabled ? "Desativar" : "Ativar"}
                      </button>
                    </>
                  )}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

export default MetodosPagamento;

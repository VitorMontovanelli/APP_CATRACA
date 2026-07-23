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

const inputStyle: React.CSSProperties = {
  height: 34,
  padding: "0 10px",
  background: "#0e1420",
  border: "1px solid #232e42",
  borderRadius: 6,
  color: "#e5e7eb",
  fontSize: 12,
  fontFamily: "inherit",
  outline: "none",
};

const typeLabels: Record<string, string> = {
  pix: "Pix",
  credit_card: "Cartão de Crédito",
  boleto: "Boleto",
  debit_card: "Cartão de Débito",
  wallet: "Carteira Digital",
  cash: "Dinheiro",
};

const methodIcons: Record<string, string> = {
  pix: "💳",
  credit_card: "💳",
  boleto: "📄",
  debit_card: "💳",
  wallet: "📱",
  cash: "💰",
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
      setMethods(m as unknown as PaymentMethod[]);
      setGateways(g as unknown as PaymentGateway[]);
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
    <div>
      <h2 style={{ margin: "0 0 24px 0", fontSize: 20, fontWeight: 500 }}>Métodos de Pagamento</h2>

      {error && <div style={{ color: "#f87171", fontSize: 12, marginBottom: 12 }}>{error}</div>}

      <div style={{ display: "flex", flexDirection: "column", gap: 8, marginBottom: 32 }}>
        {methods.map((m) => (
          <div key={m.id} style={{
            background: "#121a29", border: "1px solid #232e42", borderRadius: 8,
            padding: "14px 18px", display: "flex", alignItems: "center",
            justifyContent: "space-between", opacity: m.enabled ? 1 : 0.5,
          }}>
            <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
              <span style={{ fontSize: 18 }}>{methodIcons[m.type] || "💳"}</span>
              <div>
                <div style={{ fontSize: 14, fontWeight: 500 }}>{m.name}</div>
                <div style={{ fontSize: 12, color: "#7c8798" }}>
                  {typeLabels[m.type] || m.type}
                  {m.fee_percent > 0 && ` · taxa: ${m.fee_percent}%`}
                  {m.fee_fixed_cents > 0 && ` + R$ ${(m.fee_fixed_cents / 100).toFixed(2)}`}
                </div>
              </div>
            </div>
            {isSuper && (
              <button onClick={() => toggleMethod(m)} style={{
                padding: "6px 12px", border: "1px solid #232e42", borderRadius: 6,
                background: "transparent", color: m.enabled ? "#f87171" : "#5eead4",
                fontSize: 12, fontFamily: "inherit", cursor: "pointer",
              }}>
                {m.enabled ? "Desativar" : "Ativar"}
              </button>
            )}
          </div>
        ))}
      </div>

      <h3 style={{ margin: "0 0 16px 0", fontSize: 16, fontWeight: 500, color: "#a7b0bf" }}>
        Gateways de Pagamento
      </h3>

      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {gateways.map((g) => (
          <div key={g.id} style={{
            background: "#121a29", border: "1px solid #232e42", borderRadius: 8,
            padding: "14px 18px", display: "flex", alignItems: "center",
            justifyContent: "space-between", opacity: g.enabled ? 1 : 0.5,
          }}>
            {editingGateway === g.id ? (
              <div style={{ display: "flex", alignItems: "center", gap: 8, flex: 1 }}>
                <input value={gatewayConfig} onChange={(e) => setGatewayConfig(e.target.value)}
                  placeholder='{"api_key":"..."}' style={{ ...inputStyle, flex: 1, fontFamily: "monospace" }} />
                <button onClick={() => saveGatewayConfig(g)} style={{
                  padding: "6px 12px", border: "none", borderRadius: 6,
                  background: "#5eead4", color: "#053b32", fontSize: 12,
                  fontFamily: "inherit", cursor: "pointer",
                }}>Salvar</button>
                <button onClick={() => setEditingGateway(null)} style={{
                  padding: "6px 12px", border: "1px solid #232e42", borderRadius: 6,
                  background: "transparent", color: "#7c8798", fontSize: 12,
                  fontFamily: "inherit", cursor: "pointer",
                }}>Cancelar</button>
              </div>
            ) : (
              <>
                <div>
                  <div style={{ fontSize: 14, fontWeight: 500 }}>{g.name}</div>
                  <div style={{ fontSize: 12, color: "#7c8798" }}>
                    {g.type} · {g.enabled ? "ativo" : "inativo"}
                    {g.config && " · configurado"}
                  </div>
                </div>
                <div style={{ display: "flex", gap: 8 }}>
                  {isSuper && (
                    <>
                      <button onClick={() => {
                        setEditingGateway(g.id);
                        setGatewayConfig(g.config || "");
                      }} style={{
                        padding: "6px 12px", border: "1px solid #60a5fa", borderRadius: 6,
                        background: "transparent", color: "#60a5fa", fontSize: 12,
                        fontFamily: "inherit", cursor: "pointer",
                      }}>Configurar</button>
                      <button onClick={() => toggleGateway(g)} style={{
                        padding: "6px 12px", border: "1px solid #232e42", borderRadius: 6,
                        background: "transparent", color: g.enabled ? "#f87171" : "#5eead4",
                        fontSize: 12, fontFamily: "inherit", cursor: "pointer",
                      }}>{g.enabled ? "Desativar" : "Ativar"}</button>
                    </>
                  )}
                </div>
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

export default MetodosPagamento;

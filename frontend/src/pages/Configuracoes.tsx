import { useEffect, useState } from "react";
import {
  ObterConfigAvisoVencimento,
  SalvarConfigAvisoVencimento,
} from "../../wailsjs/go/main/App";
import { useDueNotifications } from "../components/DueNotificationsProvider";

function Configuracoes() {
  const { refresh } = useDueNotifications();
  const [dias, setDias] = useState(1);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [erro, setErro] = useState("");
  const [sucesso, setSucesso] = useState("");

  useEffect(() => {
    async function carregar() {
      try {
        const cfg = (await ObterConfigAvisoVencimento()) as unknown as {
          dias_antecedencia: number;
        };
        setDias(cfg?.dias_antecedencia ?? 1);
      } catch (e) {
        setErro("Não foi possível carregar a configuração. " + String(e));
      } finally {
        setLoading(false);
      }
    }
    carregar();
  }, []);

  async function salvar() {
    setErro("");
    setSucesso("");
    if (!Number.isInteger(dias) || dias < 0 || dias > 30) {
      setErro("Informe uma antecedência entre 0 e 30 dias.");
      return;
    }
    setSaving(true);
    try {
      await SalvarConfigAvisoVencimento(dias);
      setSucesso("Configuração salva com sucesso.");
      await refresh();
    } catch (e) {
      setErro("Não foi possível salvar. " + String(e));
    } finally {
      setSaving(false);
    }
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 24 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>
          Configurações
        </h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 6 }}>
          Ajuste o comportamento dos avisos de vencimento exibidos no sino.
        </p>
      </div>

      <div className="vmd-card" style={{ maxWidth: 620, display: "flex", flexDirection: "column", gap: 16 }}>
        <div>
          <h3 style={{ fontSize: 16, fontWeight: 600, margin: 0 }}>Avisos de vencimento</h3>
          <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 6, marginBottom: 0 }}>
            Quantos dias antes do vencimento o administrador deve ser avisado. O padrão é 1 dia.
          </p>
        </div>

        <div className="vmd-form-group" style={{ marginBottom: 0, maxWidth: 220 }}>
          <label className="vmd-label">Antecedência (dias)</label>
          <input
            type="number"
            min={0}
            max={30}
            step={1}
            value={dias}
            disabled={loading || saving}
            onChange={(e) => setDias(Number(e.target.value))}
            className="vmd-input"
          />
          <div style={{ fontSize: 11, color: "var(--text-dim)", marginTop: 6 }}>
            0 = avisar somente no dia do vencimento. Máximo de 30 dias.
          </div>
        </div>

        <div
          style={{
            fontSize: 12,
            color: "var(--text-muted)",
            background: "rgba(255,255,255,0.03)",
            border: "1px solid var(--border-color)",
            borderRadius: 8,
            padding: "10px 12px",
            lineHeight: 1.5,
          }}
        >
          Se o aviso cair em sábado ou domingo, ele é antecipado para a sexta-feira anterior.
          Faturas vencidas e ainda pendentes geram um único aviso de atraso por fatura.
        </div>

        {erro && <div style={{ fontSize: 13, color: "var(--danger)" }}>{erro}</div>}
        {sucesso && <div style={{ fontSize: 13, color: "var(--success)" }}>{sucesso}</div>}

        <div>
          <button
            className="vmd-btn vmd-btn-primary"
            onClick={salvar}
            disabled={loading || saving}
          >
            {saving ? "Salvando..." : "Salvar configuração"}
          </button>
        </div>
      </div>
    </div>
  );
}

export default Configuracoes;

import { useState } from "react";
import {
  VerificarAcessoAluno,
  CadastrarAluno,
} from "../../wailsjs/go/main/App";

interface Resultado {
  nome: string;
  liberado: boolean;
  mensagem: string;
}

function Catraca() {
  const [biometria, setBiometria] = useState("");
  const [resultado, setResultado] = useState<Resultado | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [showCadastro, setShowCadastro] = useState(false);
  const [cadNome, setCadNome] = useState("");
  const [cadCpf, setCadCpf] = useState("");
  const [cadDias, setCadDias] = useState(30);
  const [cadBio, setCadBio] = useState("");

  async function verificar() {
    setError("");
    setResultado(null);
    const id = Number(biometria);
    if (!id) {
      setError("Digite um ID de biometria válido");
      return;
    }
    setLoading(true);
    try {
      const r = await VerificarAcessoAluno(id);
      setResultado(r);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  async function handleCadastro() {
    setError("");
    if (!cadNome || !cadCpf || !cadBio) {
      setError("Preencha nome, CPF e ID da biometria");
      return;
    }
    try {
      await CadastrarAluno(cadNome, cadCpf, cadDias, Number(cadBio));
      setShowCadastro(false);
      setCadNome("");
      setCadCpf("");
      setCadDias(30);
      setCadBio("");
      setError("Aluno cadastrado na catraca com sucesso!");
      setTimeout(() => setError(""), 4000);
    } catch (e) {
      setError(String(e));
    }
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Catraca Virtual</h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Simulador de controle de acesso biométrico físico.
        </p>
      </div>

      {error && (
        <div
          className={`vmd-alert ${
            error.includes("sucesso") ? "vmd-alert-success" : "vmd-alert-danger"
          }`}
          style={{ maxWidth: 480 }}
        >
          <span>{error.includes("sucesso") ? "✨" : "⚠️"} {error}</span>
        </div>
      )}

      <div className="vmd-card" style={{ maxWidth: 480, marginBottom: 24 }}>
        <div style={{ fontSize: 14, fontWeight: 600, color: "var(--text-main)", marginBottom: 12 }}>
          Simular Leitura da Digital
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <input
            value={biometria}
            onChange={(e) => setBiometria(e.target.value)}
            placeholder="ID da Biometria (ex: 1001)"
            className="vmd-input"
            onKeyDown={(e) => e.key === "Enter" && verificar()}
          />
          <button
            onClick={verificar}
            disabled={loading}
            className="vmd-btn vmd-btn-primary"
          >
            {loading ? "Verificando..." : "Verificar"}
          </button>
        </div>
      </div>

      {resultado && (
        <div
          className="vmd-card"
          style={{
            maxWidth: 480,
            textAlign: "center",
            padding: "36px 24px",
            border: `2px solid ${resultado.liberado ? "var(--success)" : "var(--danger)"}`,
            background: resultado.liberado
              ? "radial-gradient(circle, rgba(16, 185, 129, 0.1) 0%, transparent 100%)"
              : "radial-gradient(circle, rgba(239, 68, 68, 0.1) 0%, transparent 100%)",
          }}
        >
          <div style={{ fontSize: 54, marginBottom: 16 }}>
            {resultado.liberado ? "🔓" : "🔒"}
          </div>
          <div
            style={{
              fontSize: 22,
              fontWeight: 700,
              letterSpacing: "0.5px",
              marginBottom: 8,
              color: resultado.liberado ? "var(--success)" : "var(--danger)",
            }}
          >
            {resultado.liberado ? "ACESSO PERMITIDO" : "ACESSO BLOQUEADO"}
          </div>
          {resultado.nome && (
            <div style={{ fontSize: 16, fontWeight: 600, color: "var(--text-main)", marginBottom: 6 }}>
              {resultado.nome}
            </div>
          )}
          <div style={{ fontSize: 13, color: "var(--text-muted)" }}>
            {resultado.mensagem}
          </div>
        </div>
      )}

      <div style={{ marginTop: 24, display: "flex", flexDirection: "column", gap: 16 }}>
        <div>
          <button
            onClick={() => setShowCadastro(true)}
            className="vmd-btn vmd-btn-secondary"
          >
            ➕ Cadastrar Aluno Direto na Catraca
          </button>
        </div>

        <div className="vmd-card" style={{ maxWidth: 480, padding: 16, background: "rgba(255,255,255,0.01)" }}>
          <div style={{ fontSize: 11, fontWeight: 600, color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.5px", marginBottom: 6 }}>
            IDs de Teste Rápidos
          </div>
          <div style={{ fontSize: 12, color: "var(--text-muted)", lineHeight: 1.6 }}>
            • <strong style={{ color: "var(--success)" }}>1001</strong>: Carlos (Acesso Liberado)<br />
            • <strong style={{ color: "var(--success)" }}>1002</strong>: Maria (Acesso Liberado)<br />
            • <strong style={{ color: "var(--danger)" }}>1003</strong>: João (Acesso Negado / Vencido)<br />
            • <strong style={{ color: "var(--text-dim)" }}>9999</strong>: Biometria não encontrada
          </div>
        </div>
      </div>

      {showCadastro && (
        <div className="vmd-modal-overlay" onClick={() => setShowCadastro(false)}>
          <div className="vmd-modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="vmd-modal-header">
              <h3 style={{ fontSize: 16, margin: 0 }}>Cadastrar Aluno na Catraca</h3>
              <button
                onClick={() => setShowCadastro(false)}
                className="vmd-btn vmd-btn-ghost"
                style={{ padding: 4, minWidth: "auto" }}
              >
                ✕
              </button>
            </div>

            <div className="vmd-modal-body">
              <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Nome Completo</label>
                  <input
                    value={cadNome}
                    onChange={(e) => setCadNome(e.target.value)}
                    placeholder="Ex: Pedro Silva"
                    className="vmd-input"
                  />
                </div>
                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">CPF</label>
                  <input
                    value={cadCpf}
                    onChange={(e) => setCadCpf(e.target.value)}
                    placeholder="000.000.000-00"
                    className="vmd-input"
                  />
                </div>
                <div className="vmd-grid-2">
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Validade (dias)</label>
                    <input
                      type="number"
                      min={1}
                      value={cadDias}
                      onChange={(e) => setCadDias(Number(e.target.value))}
                      className="vmd-input"
                    />
                  </div>
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">ID Biometria</label>
                    <input
                      type="number"
                      value={cadBio}
                      onChange={(e) => setCadBio(e.target.value)}
                      className="vmd-input"
                    />
                  </div>
                </div>
              </div>
            </div>

            <div className="vmd-modal-footer">
              <button
                onClick={() => setShowCadastro(false)}
                className="vmd-btn vmd-btn-secondary"
              >
                Cancelar
              </button>
              <button
                onClick={handleCadastro}
                className="vmd-btn vmd-btn-primary"
              >
                Cadastrar Aluno
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Catraca;

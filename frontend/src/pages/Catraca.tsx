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

const inputStyle: React.CSSProperties = {
  height: 40, padding: "0 14px", background: "#121a29",
  border: "1px solid #232e42", borderRadius: 8, color: "#e5e7eb",
  fontSize: 14, fontFamily: "inherit", outline: "none", width: "100%",
  boxSizing: "border-box",
};

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
    if (!id) { setError("Digite um ID de biometria válido"); return; }
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
      setCadNome(""); setCadCpf(""); setCadDias(30); setCadBio("");
      setError("Aluno cadastrado com sucesso!");
      setTimeout(() => setError(""), 3000);
    } catch (e) {
      setError(String(e));
    }
  }

  return (
    <div>
      <h2 style={{ margin: "0 0 20px 0", fontSize: 20, fontWeight: 500 }}>
        Catraca Virtual
      </h2>

      {error && (
        <div style={{
          padding: "10px 16px", borderRadius: 6, marginBottom: 16, fontSize: 13,
          background: error.includes("sucesso") ? "#5eead422" : "#f8717122",
          color: error.includes("sucesso") ? "#5eead4" : "#f87171",
          border: `1px solid ${error.includes("sucesso") ? "#5eead4" : "#f87171"}`,
        }}>{error}</div>
      )}

      <div style={{
        background: "#121a29", border: "1px solid #232e42", borderRadius: 12,
        padding: 24, marginBottom: 16, maxWidth: 480,
      }}>
        <div style={{ fontSize: 13, color: "#a7b0bf", marginBottom: 8 }}>
          Simular leitura da digital
        </div>
        <div style={{ display: "flex", gap: 8 }}>
          <input
            value={biometria}
            onChange={(e) => setBiometria(e.target.value)}
            placeholder="ID da Biometria (ex: 1001)"
            style={inputStyle}
            onKeyDown={(e) => e.key === "Enter" && verificar()}
          />
          <button onClick={verificar} disabled={loading} style={{
            padding: "10px 20px", border: "none", borderRadius: 8,
            background: "#5eead4", color: "#053b32", fontSize: 13,
            fontWeight: 600, fontFamily: "inherit", cursor: "pointer",
            whiteSpace: "nowrap", opacity: loading ? 0.6 : 1,
          }}>
            {loading ? "Verificando..." : "Verificar"}
          </button>
        </div>
      </div>

      {resultado && (
        <div style={{
          background: resultado.liberado ? "#0a2e2a" : "#2e0a0a",
          border: `2px solid ${resultado.liberado ? "#5eead4" : "#f87171"}`,
          borderRadius: 12, padding: 28, maxWidth: 480,
          textAlign: "center",
        }}>
          <div style={{
            fontSize: 48, marginBottom: 8,
          }}>{resultado.liberado ? "✅" : "⛔"}</div>
          <div style={{
            fontSize: 22, fontWeight: 600, marginBottom: 4,
            color: resultado.liberado ? "#5eead4" : "#f87171",
          }}>
            {resultado.liberado ? "ACESSO LIBERADO" : "ACESSO NEGADO"}
          </div>
          {resultado.nome && (
            <div style={{ fontSize: 16, color: "#e5e7eb", marginBottom: 4 }}>
              {resultado.nome}
            </div>
          )}
          <div style={{
            fontSize: 14, color: "#a7b0bf", marginTop: 4,
          }}>{resultado.mensagem}</div>
        </div>
      )}

      <div style={{ marginTop: 24 }}>
        <button onClick={() => setShowCadastro(true)} style={{
          padding: "8px 16px", border: "1px solid #60a5fa", borderRadius: 8,
          background: "transparent", color: "#60a5fa", fontSize: 13,
          fontFamily: "inherit", cursor: "pointer",
        }}>+ Cadastrar Aluno na Catraca</button>
      </div>

      <div style={{ marginTop: 24, fontSize: 12, color: "#7c8798" }}>
        <strong>Alunos de teste:</strong> 1001 (Carlos - ativo), 1002 (Maria - ativo), 1003 (João - vencido), 9999 (inexistente)
      </div>

      {showCadastro && (
        <div style={{
          position: "fixed", inset: 0, background: "rgba(0,0,0,0.7)",
          display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
        }} onClick={() => setShowCadastro(false)}>
          <div style={{
            background: "#121a29", border: "1px solid #232e42", borderRadius: 12,
            padding: 28, width: 400, maxWidth: "90vw",
          }} onClick={(e) => e.stopPropagation()}>
            <h3 style={{ margin: "0 0 20px 0", fontSize: 16, fontWeight: 500 }}>
              Cadastrar Aluno na Catraca
            </h3>

            <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
              <div>
                <div style={{ fontSize: 11, color: "#a7b0bf", marginBottom: 6 }}>Nome</div>
                <input value={cadNome} onChange={(e) => setCadNome(e.target.value)}
                  placeholder="Nome do aluno" style={inputStyle} />
              </div>
              <div>
                <div style={{ fontSize: 11, color: "#a7b0bf", marginBottom: 6 }}>CPF</div>
                <input value={cadCpf} onChange={(e) => setCadCpf(e.target.value)}
                  placeholder="000.000.000-00" style={inputStyle} />
              </div>
              <div style={{ display: "flex", gap: 12 }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 11, color: "#a7b0bf", marginBottom: 6 }}>Dias de validade</div>
                  <input type="number" min={1} value={cadDias}
                    onChange={(e) => setCadDias(Number(e.target.value))} style={inputStyle} />
                </div>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 11, color: "#a7b0bf", marginBottom: 6 }}>ID Biometria</div>
                  <input type="number" value={cadBio}
                    onChange={(e) => setCadBio(e.target.value)} style={inputStyle} />
                </div>
              </div>
            </div>

            <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 24 }}>
              <button onClick={() => setShowCadastro(false)} style={{
                padding: "8px 16px", border: "1px solid #232e42", borderRadius: 6,
                background: "transparent", color: "#7c8798", fontSize: 12,
                fontFamily: "inherit", cursor: "pointer",
              }}>Cancelar</button>
              <button onClick={handleCadastro} style={{
                padding: "8px 16px", border: "none", borderRadius: 6,
                background: "#5eead4", color: "#053b32", fontSize: 12,
                fontWeight: 500, fontFamily: "inherit", cursor: "pointer",
              }}>Cadastrar</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Catraca;

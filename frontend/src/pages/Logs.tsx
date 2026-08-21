import { useState, useEffect } from "react";
import { ListarAccessLogs, ListarAuditLogs } from "../../wailsjs/go/main/App";

interface AccessLog {
  id: number;
  user_id?: number;
  timestamp: string;
  resultado: string;
  motivo: string;
}

interface AuditLog {
  id: number;
  actor_id: number;
  action: string;
  target_id?: number;
  details: string;
  timestamp: string;
}

const actionLabels: Record<string, string> = {
  login: "Login no sistema",
  cadastrou_aluno_catraca: "Cadastro de aluno (catraca)",
  criou_aluno: "Criou novo aluno",
  editou_aluno: "Editou dados do aluno",
  ativou_aluno: "Ativou aluno",
  desativou_aluno: "Desativou aluno",
  criou_usuario: "Criou usuário adm",
  editou_usuario: "Editou usuário adm",
  ativou_usuario: "Ativou usuário adm",
  desativou_usuario: "Desativou usuário adm",
  deletou_usuario: "Excluiu usuário adm",
  alterou_senha: "Alterou senha",
  alterou_foto: "Alterou foto de perfil",
  criou_plano: "Criou novo plano",
  editou_plano: "Editou plano",
  ativou_plano: "Ativou plano",
  desativou_plano: "Desativou plano",
  criou_metodo_pagamento: "Criou método pagamento",
  ativou_metodo_pagamento: "Ativou método pagamento",
  desativou_metodo_pagamento: "Desativou método pagamento",
  criou_plano_aluno: "Vinculou plano ao aluno",
  cancelou_plano_aluno: "Cancelou plano de aluno",
  gerou_fatura: "Emitiu fatura",
  confirmou_pagamento: "Confirmou pagamento de fatura",
  cancelou_fatura: "Cancelou fatura",
  ativou_gateway: "Ativou gateway de pagamento",
  desativou_gateway: "Desativou gateway de pagamento",
  configurou_gateway: "Configurou parâmetros do gateway",
  agendou_aluno: "Agendou aluno na agenda",
  removeu_agendamento: "Removeu agendamento",
};

function Logs() {
  const [accessLogs, setAccessLogs] = useState<AccessLog[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [tab, setTab] = useState<"acesso" | "auditoria">("acesso");
  const [error, setError] = useState("");

  useEffect(() => {
    ListarAccessLogs()
      .then((list) => setAccessLogs((list as unknown as AccessLog[]) || []))
      .catch((e) => setError(String(e)));
    ListarAuditLogs()
      .then((list) => setAuditLogs((list as unknown as AuditLog[]) || []))
      .catch((e) => setError(String(e)));
  }, []);

  const resultBadgeClasses: Record<string, string> = {
    liberado: "vmd-badge vmd-badge-success",
    negado: "vmd-badge vmd-badge-danger",
  };

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Histórico e Auditoria</h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Histórico de acessos à catraca e logs de auditoria das ações administrativas.
        </p>
      </div>

      <div className="vmd-tabs">
        <button
          onClick={() => setTab("acesso")}
          className={`vmd-tab-btn ${tab === "acesso" ? "active" : ""}`}
        >
          Acessos à Catraca
        </button>
        <button
          onClick={() => setTab("auditoria")}
          className={`vmd-tab-btn ${tab === "auditoria" ? "active" : ""}`}
        >
          Ações de Auditoria
        </button>
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 16 }}>
          <span>⚠️ {error}</span>
        </div>
      )}

      {tab === "acesso" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {accessLogs.map((l) => (
            <div
              key={l.id}
              className="vmd-card"
              style={{
                padding: "12px 18px",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                fontSize: 13,
              }}
            >
              <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
                <span style={{ color: "var(--text-dim)", fontSize: 12, fontFamily: "monospace" }}>
                  {l.timestamp}
                </span>
                {l.motivo && <span style={{ color: "var(--text-main)", fontWeight: 500 }}>{l.motivo}</span>}
              </div>
              <span className={resultBadgeClasses[l.resultado] || "vmd-badge"}>
                {l.resultado === "liberado" ? "Liberado" : "Bloqueado"}
              </span>
            </div>
          ))}
          {accessLogs.length === 0 && (
            <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
              Nenhum log de acesso à catraca registrado.
            </div>
          )}
        </div>
      )}

      {tab === "auditoria" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {auditLogs.map((l) => (
            <div
              key={l.id}
              className="vmd-card"
              style={{
                padding: "14px 18px",
                fontSize: 13,
              }}
            >
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 8 }}>
                <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                  <span style={{ color: "var(--text-dim)", fontSize: 12, fontFamily: "monospace" }}>
                    {l.timestamp}
                  </span>
                  <span className="vmd-badge vmd-badge-secondary" style={{ fontFamily: "monospace" }}>
                    ID Actor #{l.actor_id}
                  </span>
                  <span style={{ color: "var(--text-main)", fontWeight: 600 }}>
                    {actionLabels[l.action] || l.action}
                  </span>
                </div>
                {l.target_id && (
                  <span style={{ color: "var(--text-muted)", fontSize: 11 }}>
                    Alvo ID: #{l.target_id}
                  </span>
                )}
              </div>
              {l.details && (
                <div
                  style={{
                    color: "var(--text-muted)",
                    fontSize: 11,
                    marginTop: 6,
                    fontFamily: "monospace",
                    background: "rgba(0,0,0,0.2)",
                    padding: "8px 12px",
                    borderRadius: "var(--radius-sm)",
                    border: "1px solid var(--border-color)",
                    whiteSpace: "pre-wrap",
                  }}
                >
                  {l.details}
                </div>
              )}
            </div>
          ))}
          {auditLogs.length === 0 && (
            <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
              Nenhum log de auditoria administrativa registrado.
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default Logs;
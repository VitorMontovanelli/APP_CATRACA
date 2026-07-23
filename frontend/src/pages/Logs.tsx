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
  login: "Login",
  cadastrou_aluno_catraca: "Cadastrou aluno (catraca)",
  criou_aluno: "Criou aluno",
  editou_aluno: "Editou aluno",
  ativou_aluno: "Ativou aluno",
  desativou_aluno: "Desativou aluno",
  criou_usuario: "Criou usuário",
  editou_usuario: "Editou usuário",
  ativou_usuario: "Ativou usuário",
  desativou_usuario: "Desativou usuário",
  deletou_usuario: "Deletou usuário",
  alterou_senha: "Alterou senha",
  alterou_foto: "Alterou foto",
  criou_plano: "Criou plano",
  editou_plano: "Editou plano",
  ativou_plano: "Ativou plano",
  desativou_plano: "Desativou plano",
  criou_metodo_pagamento: "Criou método de pagamento",
  ativou_metodo_pagamento: "Ativou método de pagamento",
  desativou_metodo_pagamento: "Desativou método de pagamento",
  criou_plano_aluno: "Criou plano de aluno",
  cancelou_plano_aluno: "Cancelou plano de aluno",
  gerou_fatura: "Gerou fatura",
  confirmou_pagamento: "Confirmou pagamento",
  cancelou_fatura: "Cancelou fatura",
  ativou_gateway: "Ativou gateway",
  desativou_gateway: "Desativou gateway",
  configurou_gateway: "Configurou gateway",
};

function Logs() {
  const [accessLogs, setAccessLogs] = useState<AccessLog[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [tab, setTab] = useState<"acesso" | "auditoria">("acesso");
  const [error, setError] = useState("");

  useEffect(() => {
    ListarAccessLogs()
      .then((list) => setAccessLogs(list as unknown as AccessLog[]))
      .catch((e) => setError(String(e)));
    ListarAuditLogs()
      .then((list) => setAuditLogs(list as unknown as AuditLog[]))
      .catch((e) => setError(String(e)));
  }, []);

  const resultColor: Record<string, string> = {
    liberado: "#5eead4",
    negado: "#f87171",
  };

  return (
    <div>
      <h2 style={{ margin: "0 0 20px 0", fontSize: 20, fontWeight: 500 }}>Logs</h2>

      <div style={{ display: "flex", gap: 8, marginBottom: 20 }}>
        <button onClick={() => setTab("acesso")} style={{
          padding: "7px 14px", border: "none", borderRadius: 6,
          background: tab === "acesso" ? "#5eead4" : "#121a29",
          color: tab === "acesso" ? "#053b32" : "#a7b0bf",
          fontSize: 12, fontFamily: "inherit", cursor: "pointer",
        }}>Acesso à Catraca</button>
        <button onClick={() => setTab("auditoria")} style={{
          padding: "7px 14px", border: "none", borderRadius: 6,
          background: tab === "auditoria" ? "#5eead4" : "#121a29",
          color: tab === "auditoria" ? "#053b32" : "#a7b0bf",
          fontSize: 12, fontFamily: "inherit", cursor: "pointer",
        }}>Auditoria</button>
      </div>

      {error && <div style={{ color: "#f87171", fontSize: 12, marginBottom: 12 }}>{error}</div>}

      {tab === "acesso" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          {accessLogs.map((l) => (
            <div key={l.id} style={{
              background: "#121a29", border: "1px solid #232e42", borderRadius: 8,
              padding: "12px 16px", display: "flex", alignItems: "center",
              justifyContent: "space-between", fontSize: 13,
            }}>
              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <span style={{ color: "#7c8798", fontSize: 12 }}>{l.timestamp}</span>
                {l.motivo && <span style={{ color: "#a7b0bf" }}>{l.motivo}</span>}
              </div>
              <span style={{
                padding: "3px 10px", borderRadius: 4, fontSize: 11, fontWeight: 500,
                background: (resultColor[l.resultado] || "#7c8798") + "22",
                color: resultColor[l.resultado] || "#7c8798",
              }}>{l.resultado}</span>
            </div>
          ))}
          {accessLogs.length === 0 && (
            <div style={{ color: "#7c8798", fontSize: 13, textAlign: "center", padding: 40 }}>
              Nenhum log de acesso registrado.
            </div>
          )}
        </div>
      )}

      {tab === "auditoria" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          {auditLogs.map((l) => (
            <div key={l.id} style={{
              background: "#121a29", border: "1px solid #232e42", borderRadius: 8,
              padding: "12px 16px", fontSize: 13,
            }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  <span style={{ color: "#7c8798", fontSize: 12 }}>{l.timestamp}</span>
                  <span style={{
                    padding: "2px 8px", borderRadius: 4, fontSize: 11,
                    background: "#60a5fa22", color: "#60a5fa",
                  }}>#{l.actor_id}</span>
                  <span style={{ color: "#e5e7eb" }}>{actionLabels[l.action] || l.action}</span>
                </div>
                {l.target_id && (
                  <span style={{ color: "#7c8798", fontSize: 11 }}>target: {l.target_id}</span>
                )}
              </div>
              {l.details && (
                <div style={{ color: "#7c8798", fontSize: 11, marginTop: 4, fontFamily: "monospace" }}>
                  {l.details}
                </div>
              )}
            </div>
          ))}
          {auditLogs.length === 0 && (
            <div style={{ color: "#7c8798", fontSize: 13, textAlign: "center", padding: 40 }}>
              Nenhum log de auditoria registrado.
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default Logs;
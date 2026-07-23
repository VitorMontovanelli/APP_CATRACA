import { useState, useEffect, useRef } from "react";
import { CriarInvoiceTeste, DeletarDadosTeste, ListarCobranca, SalvarComprovante } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

const tdStyle: React.CSSProperties = {
  padding: "10px 12px", fontSize: 12, borderBottom: "1px solid #232e42",
  verticalAlign: "middle",
};

const thStyle: React.CSSProperties = {
  padding: "10px 12px", fontSize: 11, fontWeight: 600, color: "#7c8798",
  textTransform: "uppercase", textAlign: "left", borderBottom: "1px solid #232e42",
  whiteSpace: "nowrap",
};

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function formatDate(d: string): string {
  if (!d) return "—";
  return d.replace("T", " ").split(" ")[0];
}

const statusLabel: Record<string, string> = {
  pending: "Pendente", paid: "Pago", overdue: "Vencido",
  cancelled: "Cancelado", refunded: "Reembolsado",
};

const statusColor: Record<string, string> = {
  pending: "#fbbf24", paid: "#5eead4", overdue: "#f87171",
  cancelled: "#7c8798", refunded: "#60a5fa",
};

function Cobranca() {
  const [alunos, setAlunos] = useState<main.CobrancaAluno[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [uploadingId, setUploadingId] = useState<number | null>(null);
  const [pendingUploadId, setPendingUploadId] = useState<number | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => { load(); }, []);

  async function load() {
    setLoading(true);
    try {
      const data = await ListarCobranca();
      setAlunos(data as unknown as main.CobrancaAluno[]);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  function handleUpload(invoiceId: number) {
    setPendingUploadId(invoiceId);
    fileInputRef.current?.click();
  }

  function handleFileSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file || !pendingUploadId) return;

    setUploadingId(pendingUploadId);
    setError("");

    const reader = new FileReader();
    reader.onload = async () => {
      try {
        const base64 = reader.result as string;
        await SalvarComprovante(pendingUploadId, base64, file.name);
        await load();
      } catch (err) {
        setError(String(err));
      } finally {
        setUploadingId(null);
        setPendingUploadId(null);
        if (fileInputRef.current) fileInputRef.current.value = "";
      }
    };
    reader.onerror = () => {
      setError("Erro ao ler o arquivo");
      setUploadingId(null);
      setPendingUploadId(null);
    };
    reader.readAsDataURL(file);
  }

  const filtered = searchTerm
    ? alunos.filter((a) => a.nome.toLowerCase().includes(searchTerm.toLowerCase()))
    : alunos;

  const inputSearchStyle: React.CSSProperties = {
    height: 38, padding: "0 14px", background: "#121a29",
    border: "1px solid #232e42", borderRadius: 8, color: "#e5e7eb",
    fontSize: 13, fontFamily: "inherit", outline: "none", width: 320, maxWidth: "100%",
    boxSizing: "border-box",
  };

  const planStatusLabel: Record<string, string> = {
    active: "Ativo", overdue: "Inadimplente", expired: "Expirado",
  };

  return (
    <div>
      <input
        ref={fileInputRef}
        type="file"
        accept=".pdf"
        onChange={handleFileSelected}
        style={{ display: "none" }}
      />

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
        <div>
          <h2 style={{ margin: 0, fontSize: 20, fontWeight: 500 }}>Cobrança</h2>
          {!loading && (
            <div style={{ fontSize: 13, color: "#7c8798", marginTop: 4 }}>
              {alunos.length} aluno(s)
            </div>
          )}
        </div>
        <input
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Buscar por nome..."
          style={inputSearchStyle}
        />
      </div>

      {error && <div style={{ color: "#f87171", fontSize: 12, marginBottom: 12 }}>{error}</div>}

      {loading ? (
        <div style={{ color: "#7c8798", fontSize: 13, textAlign: "center", padding: 40 }}>
          Carregando...
        </div>
      ) : filtered.length === 0 ? (
        <div style={{ color: "#7c8798", fontSize: 13, textAlign: "center", padding: 40 }}>
          {searchTerm ? "Nenhum aluno encontrado." : "Nenhum aluno cadastrado."}
        </div>
      ) : (
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse", minWidth: 800 }}>
            <thead>
              <tr>
                <th style={{ ...thStyle, width: 28 }}></th>
                <th style={thStyle}>Aluno</th>
                <th style={thStyle}>Plano</th>
                <th style={thStyle}>Status</th>
                <th style={thStyle}>Vencimento</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((al) => {
                const isExpanded = expandedId === al.student_id;
                const ultFat = al.faturas && al.faturas.length > 0 ? al.faturas[0] : null;
                return (
                  <tr key={al.student_id}>
                    <td style={{ ...tdStyle, cursor: "pointer", textAlign: "center" }}
                        onClick={() => setExpandedId(isExpanded ? null : al.student_id)}>
                      <span style={{ color: "#7c8798", fontSize: 12, display: "inline-block",
                        transform: isExpanded ? "rotate(90deg)" : "rotate(0deg)" }}>▶</span>
                    </td>
                    <td style={tdStyle}>
                      <div style={{ fontWeight: 500, fontSize: 13 }}>{al.nome}</div>
                      <div style={{ color: "#7c8798", fontSize: 11 }}>{al.cpf}</div>
                    </td>
                    <td style={tdStyle}>
                      <div>{al.plano_nome}</div>
                      <div style={{ color: "#7c8798" }}>{formatCents(al.plano_preco)}/mês</div>
                    </td>
                    <td style={tdStyle}>
                      <span style={{
                        padding: "2px 8px", borderRadius: 4, fontSize: 11, fontWeight: 500,
                        background: (al.status_plano === "active" ? "#5eead422" : al.status_plano === "overdue" ? "#f8717122" : "#7c879822"),
                        color: al.status_plano === "active" ? "#5eead4" : al.status_plano === "overdue" ? "#f87171" : "#7c8798",
                      }}>
                        {planStatusLabel[al.status_plano] || al.status_plano}
                      </span>
                    </td>
                    <td style={tdStyle}>
                      {ultFat ? (
                        <div>
                          <div>{formatDate(ultFat.due_date)}</div>
                          <div style={{ color: "#7c8798", fontSize: 11 }}>{formatCents(ultFat.amount_cents)}</div>
                        </div>
                      ) : (
                        <span style={{ color: "#7c8798" }}>—</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>

          {expandedId && <FaturaDetalhe aluno={alunos.find((a) => a.student_id === expandedId)} onUpload={handleUpload} uploadingId={uploadingId} />}
        </div>
      )}

      {!loading && (
        <div style={{ marginTop: 20, display: "flex", gap: 8 }}>
          <button onClick={load} style={{
            padding: "8px 16px", border: "1px solid #232e42", borderRadius: 6,
            background: "transparent", color: "#a7b0bf", fontSize: 12,
            fontFamily: "inherit", cursor: "pointer",
          }}>Atualizar</button>
          <button onClick={async () => {
            setError("");
            try {
              await CriarInvoiceTeste();
              await load();
            } catch (e) { setError(String(e)); }
          }} style={{
            padding: "8px 16px", border: "1px solid #60a5fa", borderRadius: 6,
            background: "transparent", color: "#60a5fa", fontSize: 12,
            fontFamily: "inherit", cursor: "pointer",
          }}>Criar Teste</button>
          <button onClick={async () => {
            setError("");
            try {
              await DeletarDadosTeste();
              await load();
            } catch (e) { setError(String(e)); }
          }} style={{
            padding: "8px 16px", border: "1px solid #f87171", borderRadius: 6,
            background: "#f8717122", color: "#f87171", fontSize: 12,
            fontFamily: "inherit", cursor: "pointer",
          }}>Limpar Testes</button>
        </div>
      )}
    </div>
  );
}

function FaturaDetalhe({ aluno, onUpload, uploadingId }: {
  aluno?: main.CobrancaAluno;
  onUpload: (id: number) => void;
  uploadingId: number | null;
}) {
  if (!aluno) return null;
  return (
    <div style={{
      background: "#0e1420", border: "1px solid #232e42", borderRadius: 8,
      padding: "16px 18px", marginTop: 12,
    }}>
      <h4 style={{ margin: "0 0 12px 0", fontSize: 13, fontWeight: 500, color: "#e5e7eb" }}>
        Faturas de {aluno.nome}
      </h4>
      {aluno.faturas && aluno.faturas.length > 0 ? (
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr>
              <th style={{ padding: "6px 8px", fontSize: 11, color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}>Vencimento</th>
              <th style={{ padding: "6px 8px", fontSize: 11, color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}>Valor</th>
              <th style={{ padding: "6px 8px", fontSize: 11, color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}>Status</th>
              <th style={{ padding: "6px 8px", fontSize: 11, color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}>Pagamento</th>
              <th style={{ padding: "6px 8px", fontSize: 11, color: "#7c8798", textAlign: "left", borderBottom: "1px solid #232e42" }}></th>
            </tr>
          </thead>
          <tbody>
            {aluno.faturas.map((f) => (
              <tr key={f.id}>
                <td style={{ padding: "6px 8px", fontSize: 12, color: "#e5e7eb", borderBottom: "1px solid #1a2440" }}>
                  {formatDate(f.due_date)}
                  {f.status !== "paid" && f.dias_vencido > 0 && (
                    <span style={{ color: "#f87171", marginLeft: 6, fontSize: 11 }}>{f.dias_vencido}d</span>
                  )}
                </td>
                <td style={{ padding: "6px 8px", fontSize: 12, color: "#a7b0bf", borderBottom: "1px solid #1a2440" }}>
                  {formatCents(f.amount_cents)}
                </td>
                <td style={{ padding: "6px 8px", fontSize: 12, borderBottom: "1px solid #1a2440" }}>
                  <span style={{
                    padding: "2px 6px", borderRadius: 4, fontSize: 11, fontWeight: 500,
                    background: (statusColor[f.status] || "#7c8798") + "22",
                    color: statusColor[f.status] || "#7c8798",
                  }}>
                    {statusLabel[f.status] || f.status}
                  </span>
                </td>
                <td style={{ padding: "6px 8px", fontSize: 12, color: "#a7b0bf", borderBottom: "1px solid #1a2440" }}>
                  {f.paid_at ? formatDate(f.paid_at) : "—"}
                </td>
                <td style={{ padding: "6px 8px", fontSize: 12, borderBottom: "1px solid #1a2440" }}>
                  {f.status === "pending" || f.status === "overdue" ? (
                    <button
                      onClick={() => onUpload(f.id)}
                      disabled={uploadingId === f.id}
                      style={{
                        padding: "4px 10px", border: "1px solid #5eead4", borderRadius: 6,
                        background: "transparent", color: "#5eead4", fontSize: 11,
                        fontFamily: "inherit", cursor: "pointer",
                        opacity: uploadingId === f.id ? 0.5 : 1,
                      }}
                    >
                      {uploadingId === f.id ? "Enviando..." : "Upload PDF"}
                    </button>
                  ) : (
                    <span style={{ color: "#3a4255", fontSize: 11 }}>
                      {f.comprovante ? "OK" : ""}
                    </span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <div style={{ color: "#7c8798", fontSize: 12, padding: 8 }}>Nenhuma fatura encontrada.</div>
      )}
    </div>
  );
}

export default Cobranca;

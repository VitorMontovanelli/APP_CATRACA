import { useState, useEffect, useRef } from "react";
import { ListarInadimplentes, SalvarComprovante } from "../../wailsjs/go/main/App";
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

function Cobranca() {
  const [invoices, setInvoices] = useState<main.InadimplenteReport[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [uploadingId, setUploadingId] = useState<number | null>(null);
  const [pendingUploadId, setPendingUploadId] = useState<number | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => { load(); }, []);

  async function load() {
    setLoading(true);
    try {
      const data = await ListarInadimplentes();
      setInvoices(data as unknown as main.InadimplenteReport[]);
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
    ? invoices.filter((a) => a.nome.toLowerCase().includes(searchTerm.toLowerCase()))
    : invoices;

  const inputSearchStyle: React.CSSProperties = {
    height: 38, padding: "0 14px", background: "#121a29",
    border: "1px solid #232e42", borderRadius: 8, color: "#e5e7eb",
    fontSize: 13, fontFamily: "inherit", outline: "none", width: 320, maxWidth: "100%",
    boxSizing: "border-box",
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
              {invoices.length} fatura(s) vencida(s)
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
          {searchTerm ? "Nenhuma fatura encontrada." : "Nenhuma fatura vencida."}
        </div>
      ) : (
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse", minWidth: 800 }}>
            <thead>
              <tr>
                <th style={thStyle}>Aluno</th>
                <th style={thStyle}>Plano</th>
                <th style={thStyle}>Valor</th>
                <th style={thStyle}>Vencimento</th>
                <th style={thStyle}>Dias</th>
                <th style={thStyle}>Total Aberto</th>
                <th style={thStyle}></th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((f) => (
                <tr key={f.invoice_id}>
                  <td style={tdStyle}>
                    <div style={{ fontWeight: 500, fontSize: 13 }}>{f.nome}</div>
                    <div style={{ color: "#7c8798", fontSize: 11 }}>{f.cpf}</div>
                  </td>
                  <td style={tdStyle}>{f.plano_nome}</td>
                  <td style={tdStyle}>
                    <div>{formatCents(f.valor_devido)}</div>
                    <div style={{ color: "#7c8798", fontSize: 11 }}>{formatCents(f.plano_preco)}/mês</div>
                  </td>
                  <td style={tdStyle}>{formatDate(f.data_vencimento)}</td>
                  <td style={tdStyle}>
                    <span style={{
                      padding: "2px 8px", borderRadius: 4, fontSize: 11, fontWeight: 500,
                      color: f.dias_vencido > 5 ? "#f87171" : f.dias_vencido > 0 ? "#fbbf24" : "#5eead4",
                      background: f.dias_vencido > 5 ? "#f8717122" : f.dias_vencido > 0 ? "#fbbf2422" : "#5eead422",
                    }}>
                      {f.dias_vencido}d
                    </span>
                  </td>
                  <td style={tdStyle}>{formatCents(f.total_em_aberto)}</td>
                  <td style={tdStyle}>
                    <button
                      onClick={() => handleUpload(f.invoice_id)}
                      disabled={uploadingId === f.invoice_id}
                      style={{
                        padding: "4px 10px", border: "1px solid #5eead4", borderRadius: 6,
                        background: "transparent", color: "#5eead4", fontSize: 11,
                        fontFamily: "inherit", cursor: "pointer",
                        opacity: uploadingId === f.invoice_id ? 0.5 : 1,
                      }}
                    >
                      {uploadingId === f.invoice_id ? "Enviando..." : "Upload PDF"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!loading && (
        <div style={{ marginTop: 20, display: "flex", gap: 8 }}>
          <button onClick={load} style={{
            padding: "8px 16px", border: "1px solid #232e42", borderRadius: 6,
            background: "transparent", color: "#a7b0bf", fontSize: 12,
            fontFamily: "inherit", cursor: "pointer",
          }}>Atualizar</button>
        </div>
      )}
    </div>
  );
}

export default Cobranca;

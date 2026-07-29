import { useState, useEffect, useRef } from "react";
import { ListarInadimplentes, SalvarComprovante } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

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

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <input
        ref={fileInputRef}
        type="file"
        accept=".pdf"
        onChange={handleFileSelected}
        style={{ display: "none" }}
      />

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24, flexWrap: "wrap", gap: 16 }}>
        <div>
          <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Painel de Cobranças</h2>
          <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
            {!loading ? `${invoices.length} fatura(s) vencida(s) no sistema.` : "Carregando faturas..."}
          </p>
        </div>
        <input
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Buscar por aluno..."
          className="vmd-input"
          style={{ width: 320, maxWidth: "100%" }}
        />
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 16 }}>
          <span>⚠️ {error}</span>
        </div>
      )}

      {loading ? (
        <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
          Carregando lista de inadimplentes...
        </div>
      ) : filtered.length === 0 ? (
        <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
          {searchTerm ? "Nenhuma fatura encontrada para esta busca." : "Parabéns! Não existem faturas em atraso."}
        </div>
      ) : (
        <div className="vmd-table-container">
          <table className="vmd-table">
            <thead>
              <tr>
                <th>Aluno</th>
                <th>Plano</th>
                <th>Valor Fatura</th>
                <th>Vencimento</th>
                <th>Atraso</th>
                <th>Total em Aberto</th>
                <th style={{ textAlign: "right" }}>Ações</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((f) => (
                <tr key={f.invoice_id}>
                  <td>
                    <div style={{ fontWeight: 600 }}>{f.nome}</div>
                    <div style={{ color: "var(--text-muted)", fontSize: 11 }}>CPF: {f.cpf}</div>
                  </td>
                  <td>{f.plano_nome}</td>
                  <td>
                    <div>{formatCents(f.valor_devido)}</div>
                    <div style={{ color: "var(--text-dim)", fontSize: 11 }}>{formatCents(f.plano_preco)}/mês</div>
                  </td>
                  <td>{formatDate(f.data_vencimento)}</td>
                  <td>
                    <span
                      className={`vmd-badge ${
                        f.dias_vencido > 15
                          ? "vmd-badge-danger"
                          : f.dias_vencido > 0
                          ? "vmd-badge-warning"
                          : "vmd-badge-secondary"
                      }`}
                    >
                      {f.dias_vencido} dias
                    </span>
                  </td>
                  <td style={{ fontWeight: 700, color: "var(--danger)" }}>{formatCents(f.total_em_aberto)}</td>
                  <td style={{ textAlign: "right" }}>
                    <button
                      onClick={() => handleUpload(f.invoice_id)}
                      disabled={uploadingId === f.invoice_id}
                      className="vmd-btn vmd-btn-success"
                      style={{ padding: "5px 10px", height: 28, fontSize: 11 }}
                    >
                      {uploadingId === f.invoice_id ? "Enviando..." : "📤 Comprovante PDF"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}


    </div>
  );
}

export default Cobranca;

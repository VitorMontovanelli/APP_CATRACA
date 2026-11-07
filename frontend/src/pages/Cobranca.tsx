import { useState, useEffect, useRef } from "react";
import {
  ListarFaturasCobranca,
  SalvarComprovante,
  ObterConfigAvisoVencimento,
  SalvarConfigAvisoVencimento,
} from "../../wailsjs/go/main/App";
import PagamentoModal from "../components/PagamentoModal";
import { useDueNotifications } from "../components/DueNotificationsProvider";

interface FaturaCobranca {
  invoice_id: number;
  student_id: number;
  nome: string;
  cpf: string;
  plano_nome: string;
  plano_preco: number;
  valor_devido: number;
  due_date: string;
  dias_restantes: number;
  no_prazo_aviso: boolean;
  total_em_aberto: number;
  comprovante?: string | null;
  student_ativo: boolean;
}

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function formatDate(d: string): string {
  if (!d) return "—";
  const iso = d.length > 10 ? d.slice(0, 10) : d;
  const [ano, mes, dia] = iso.split("-");
  return ano && mes && dia ? `${dia}/${mes}/${ano}` : d;
}

function normalizar(texto: string): string {
  return (texto || "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase();
}

function prazoDaFatura(f: FaturaCobranca): { texto: string; classe: string } {
  if (f.dias_restantes < 0) {
    const dias = -f.dias_restantes;
    return {
      texto: dias === 1 ? "Vencida há 1 dia" : `Vencida há ${dias} dias`,
      classe: "vmd-badge vmd-badge-danger",
    };
  }
  if (f.dias_restantes === 0) {
    return { texto: "Vence hoje", classe: "vmd-badge vmd-badge-danger" };
  }
  if (f.dias_restantes === 1) {
    return { texto: "Vence amanhã", classe: "vmd-badge vmd-badge-warning" };
  }
  return {
    texto: `Vence em ${f.dias_restantes} dias`,
    classe: "vmd-badge vmd-badge-warning",
  };
}

function Cobranca() {
  const [invoices, setInvoices] = useState<FaturaCobranca[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [uploadingId, setUploadingId] = useState<number | null>(null);
  const [pendingUploadId, setPendingUploadId] = useState<number | null>(null);
  const [pagamento, setPagamento] = useState<FaturaCobranca | null>(null);
  const [antecedencia, setAntecedencia] = useState(1);
  const [configLoading, setConfigLoading] = useState(true);
  const [configSaving, setConfigSaving] = useState(false);
  const [configFeedback, setConfigFeedback] = useState("");
  const [configError, setConfigError] = useState("");
  const [highlightId, setHighlightId] = useState<number | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { refresh: refreshAvisos, focusInvoiceId, clearFocus } = useDueNotifications();

  useEffect(() => {
    load();
    loadConfig();
  }, []);

  useEffect(() => {
    if (focusInvoiceId == null) return;
    setSearchTerm("");
    setHighlightId(focusInvoiceId);
    const clearTimer = window.setTimeout(() => {
      setHighlightId(null);
      clearFocus();
    }, 6000);
    return () => {
      window.clearTimeout(clearTimer);
    };
  }, [focusInvoiceId, clearFocus]);

  useEffect(() => {
    if (highlightId == null || loading) return;
    document
      .getElementById(`fatura-${highlightId}`)
      ?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [highlightId, loading]);

  async function load() {
    setLoading(true);
    try {
      const data = await ListarFaturasCobranca();
      setInvoices((data as unknown as FaturaCobranca[]) || []);
      setError("");
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  async function loadConfig() {
    try {
      const cfg = (await ObterConfigAvisoVencimento()) as unknown as {
        dias_antecedencia: number;
      };
      setAntecedencia(cfg?.dias_antecedencia ?? 1);
    } catch (e) {
      setConfigError("Não foi possível carregar a antecedência. " + String(e));
    } finally {
      setConfigLoading(false);
    }
  }

  async function salvarConfig() {
    setConfigError("");
    setConfigFeedback("");
    if (!Number.isInteger(antecedencia) || antecedencia < 0 || antecedencia > 30) {
      setConfigError("Informe uma antecedência entre 0 e 30 dias.");
      return;
    }
    setConfigSaving(true);
    try {
      await SalvarConfigAvisoVencimento(antecedencia);
      await load();
      await refreshAvisos();
      setConfigFeedback("Antecedência atualizada.");
    } catch (e) {
      setConfigError("Não foi possível salvar. " + String(e));
    } finally {
      setConfigSaving(false);
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

  async function confirmouPagamento() {
    await load();
    await refreshAvisos();
  }

  const termo = normalizar(searchTerm.trim());
  const combinaBusca = (f: FaturaCobranca) => !termo || normalizar(f.nome).includes(termo);
  const vencendo = invoices.filter(
    (f) => f.dias_restantes >= 0 && f.no_prazo_aviso && combinaBusca(f),
  );
  const atrasadas = invoices.filter((f) => f.dias_restantes < 0 && combinaBusca(f));

  function renderFatura(f: FaturaCobranca) {
    const prazo = prazoDaFatura(f);
    const destacada = highlightId === f.invoice_id;
    return (
      <div
        key={f.invoice_id}
        id={`fatura-${f.invoice_id}`}
        className="vmd-card"
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          flexWrap: "wrap",
          gap: 16,
          padding: "16px 20px",
          border: destacada ? "1px solid var(--primary)" : undefined,
          boxShadow: destacada ? "0 0 0 3px rgba(56, 189, 248, 0.15)" : undefined,
          transition: "border-color 0.2s, box-shadow 0.2s",
        }}
      >
        <div style={{ flex: 1, minWidth: 240 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap" }}>
            <span style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)" }}>{f.nome}</span>
            <span className={prazo.classe}>{prazo.texto}</span>
          </div>
          <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
            CPF: {f.cpf || "—"} · Plano: <strong style={{ color: "var(--text-main)" }}>{f.plano_nome}</strong>
          </div>
          <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
            Vencimento: <strong style={{ color: "var(--text-main)" }}>{formatDate(f.due_date)}</strong>
            {" · "}Valor: <strong style={{ color: "var(--text-main)" }}>{formatCents(f.valor_devido)}</strong>
            {" · "}Total em aberto: <strong style={{ color: "var(--danger)" }}>{formatCents(f.total_em_aberto)}</strong>
          </div>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: 8, flexShrink: 0 }}>
          <button
            onClick={() => setPagamento(f)}
            className="vmd-btn vmd-btn-success"
            style={{ padding: "6px 12px", height: 32, fontSize: 11 }}
          >
            Confirmar pago
          </button>
          <button
            onClick={() => handleUpload(f.invoice_id)}
            disabled={uploadingId === f.invoice_id}
            className="vmd-btn vmd-btn-secondary"
            style={{ padding: "5px 10px", height: 28, fontSize: 11 }}
          >
            {uploadingId === f.invoice_id ? "Enviando..." : "📤 Comprovante PDF"}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <input
        ref={fileInputRef}
        type="file"
        accept=".pdf"
        onChange={handleFileSelected}
        style={{ display: "none" }}
      />

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20, flexWrap: "wrap", gap: 16 }}>
        <div>
          <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Painel de Cobranças</h2>
        </div>
        <input
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Buscar por aluno..."
          className="vmd-input"
          style={{ width: 320, maxWidth: "100%" }}
        />
      </div>

      <div className="vmd-card" style={{ padding: "14px 18px", marginBottom: 20 }}>
        <div style={{ display: "flex", alignItems: "flex-end", gap: 12, flexWrap: "wrap" }}>
          <div className="vmd-form-group" style={{ marginBottom: 0, width: 180 }}>
            <label className="vmd-label">Antecedência do aviso (dias)</label>
            <input
              type="number"
              min={0}
              max={30}
              step={1}
              value={antecedencia}
              disabled={configLoading || configSaving}
              onChange={(e) => setAntecedencia(Number(e.target.value))}
              className="vmd-input"
            />
          </div>
          <button
            className="vmd-btn vmd-btn-primary"
            onClick={salvarConfig}
            disabled={configLoading || configSaving}
            style={{ height: 38 }}
          >
            {configSaving ? "Salvando..." : "Salvar"}
          </button>
          <div style={{ fontSize: 12, color: "var(--text-muted)", paddingBottom: 9 }}>
            Padrão: 1 dia. Avisos que caem no sábado ou domingo são antecipados para a sexta anterior.
          </div>
        </div>
        {configError && <div style={{ fontSize: 12, color: "var(--danger)", marginTop: 8 }}>{configError}</div>}
        {configFeedback && <div style={{ fontSize: 12, color: "var(--success)", marginTop: 8 }}>{configFeedback}</div>}
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 16 }}>
          <span>⚠️ {error}</span>
        </div>
      )}

      {loading ? (
        <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
          Carregando faturas...
        </div>
      ) : (
        <>
          <div style={{ marginBottom: 24 }}>
            <div style={{ marginBottom: 12 }}>
              <h3 style={{ fontSize: 16, fontWeight: 600, margin: 0 }}>Vencendo em breve</h3>
              <p style={{ fontSize: 12, color: "var(--text-muted)", margin: "4px 0 0" }}>
                Faturas pendentes que entraram no prazo de aviso.
              </p>
            </div>
            {vencendo.length === 0 ? (
              <div className="vmd-card" style={{ textAlign: "center", padding: 28, color: "var(--text-muted)", fontSize: 13 }}>
                {searchTerm
                  ? "Nenhuma fatura vencendo encontrada para esta busca."
                  : "Nenhuma fatura vencendo dentro do prazo de aviso."}
              </div>
            ) : (
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {vencendo.map(renderFatura)}
              </div>
            )}
          </div>

          <div>
            <div style={{ marginBottom: 12 }}>
              <h3 style={{ fontSize: 16, fontWeight: 600, margin: 0 }}>Faturas em atraso</h3>
            </div>
            {atrasadas.length === 0 ? (
              <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
                {searchTerm
                  ? "Nenhuma fatura em atraso encontrada para esta busca."
                  : "Parabéns! Não existem faturas em atraso."}
              </div>
            ) : (
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {atrasadas.map(renderFatura)}
              </div>
            )}
          </div>
        </>
      )}

      {pagamento && (
        <PagamentoModal
          invoice={{
            id: pagamento.invoice_id,
            student_name: pagamento.nome,
            plan_name: pagamento.plano_nome,
            amount_cents: pagamento.valor_devido,
          }}
          onClose={() => setPagamento(null)}
          onConfirmed={confirmouPagamento}
        />
      )}
    </div>
  );
}

export default Cobranca;

import { useState, useEffect, useCallback, useRef } from "react";
import {
  ListarStudentsComPlanos,
  CriarStudentComPlano,
  AtualizarStudent,
  AtualizarPlanoAluno,
  AtivarStudent,
  ListarPaymentMethods,
  ListarPlanos,
  GerarInvoice,
  SalvarLaudoAluno,
  RemoverLaudoAluno,
  BaixarLaudo,
  SalvarFotoAluno,
  RemoverFotoAluno,
  ObterFotoAluno,
  CapturarFotoWebcam,
} from "../../wailsjs/go/main/App";

interface StudentComPlano {
  id: number;
  nome: string;
  cpf: string;
  data_nascimento?: string;
  telefone?: string;
  telefone_urgencia?: string;
  email?: string;
  laudo_medico?: string;
  foto?: string;
  fotoDataUrl?: string;
  forma_pagamento_id?: number;
  data_entrada: string;
  observacao?: string;
  ativo: boolean;
  created_at: string;
  plano_nome?: string;
  plano_preco?: number;
  plano_status?: string;
  student_plan_id?: number;
  payment_method?: string;
  due_day?: number;
  ultima_fatura?: string;
  fatura_valor?: number;
  fatura_vencimento?: string;
}

interface Plan {
  id: number;
  name: string;
  price_cents: number;
  duration_days: number;
  active: boolean;
}

interface PaymentMethod {
  id: number;
  name: string;
  type: string;
  enabled: boolean;
}

interface AlunosProps {
  user: { cargo: string };
}

function maskCPF(v: string): string {
  const d = v.replace(/\D/g, "").slice(0, 11);
  return d.replace(/^(\d{3})(\d{3})(\d{3})(\d{2})$/, "$1.$2.$3-$4")
          .replace(/^(\d{3})(\d{3})(\d{1,3})$/, "$1.$2.$3")
          .replace(/^(\d{3})(\d{1,3})$/, "$1.$2");
}

function maskTel(v: string): string {
  const d = v.replace(/\D/g, "").slice(0, 11);
  return d.replace(/^(\d{2})(\d{5})(\d{4})$/, "($1) $2-$3")
          .replace(/^(\d{2})(\d{1,5})$/, "($1) $2")
          .replace(/^(\d{1,2})$/, "($1");
}

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

const statusBadgeClasses: Record<string, string> = {
  pending: "vmd-badge vmd-badge-warning",
  paid: "vmd-badge vmd-badge-success",
  overdue: "vmd-badge vmd-badge-danger",
  cancelled: "vmd-badge vmd-badge-secondary",
};

const statusLabels: Record<string, string> = {
  pending: "Pendente",
  paid: "Pago",
  overdue: "Vencido",
  cancelled: "Cancelado",
};

function Alunos({ user }: AlunosProps) {
  const [students, setStudents] = useState<StudentComPlano[]>([]);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editId, setEditId] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; msg: string } | null>(null);
  const [loadingInvoice, setLoadingInvoice] = useState<number | null>(null);
  const [searchTerm, setSearchTerm] = useState("");

  const filteredStudents = searchTerm
    ? students.filter((s) => s.nome.toLowerCase().includes(searchTerm.toLowerCase()))
    : students;

  const [nome, setNome] = useState("");
  const [cpf, setCpf] = useState("");
  const [dataNasc, setDataNasc] = useState("");
  const [telefone, setTelefone] = useState("");
  const [telefoneUrgencia, setTelefoneUrgencia] = useState("");
  const [email, setEmail] = useState("");
  const [formaPagamentoId, setFormaPagamentoId] = useState(0);
  const [planId, setPlanId] = useState(0);
  const [dueDay, setDueDay] = useState(5);

  const [laudoFile, setLaudoFile] = useState<File | null>(null);
  const [laudoRemoved, setLaudoRemoved] = useState(false);
  const [uploadingLaudo, setUploadingLaudo] = useState(false);
  const laudoInputRef = useRef<HTMLInputElement>(null);

  const [fotoDataUrl, setFotoDataUrl] = useState("");
  const [loadingCameraGo, setLoadingCameraGo] = useState(false);
  const [photoChanged, setPhotoChanged] = useState(false);
  const [photoRemoved, setPhotoRemoved] = useState(false);
  const fotoInputRef = useRef<HTMLInputElement>(null);

  const canManage = user.cargo === "super_admin" || user.cargo === "admin";

  const showFeedback = useCallback((type: "success" | "error", msg: string) => {
    setFeedback({ type, msg });
    setTimeout(() => setFeedback(null), 4000);
  }, []);

  useEffect(() => { load(); }, []);



  async function handleRegistrarPagamento(studentPlanId: number) {
    setLoadingInvoice(studentPlanId);
    try {
      await GerarInvoice(studentPlanId);
      showFeedback("success", "Fatura gerada com sucesso!");
      await load();
    } catch (e) {
      showFeedback("error", String(e));
    } finally {
      setLoadingInvoice(null);
    }
  }

  async function load() {
    try {
      const [s, p, m] = await Promise.all([
        ListarStudentsComPlanos(),
        ListarPlanos(),
        ListarPaymentMethods(),
      ]);
      const list = (s as unknown as StudentComPlano[]) || [];
      setStudents(await Promise.all(list.map(async (st) => {
        if (st.foto) {
          try {
            const url = (await ObterFotoAluno(st.id)) as unknown as string;
            return { ...st, fotoDataUrl: url || "" };
          } catch { return st; }
        }
        return st;
      })));
      setPlans(((p as unknown as Plan[]) || []).filter((pl) => pl.active));
      setMethods(((m as unknown as PaymentMethod[]) || []).filter((pm) => pm.enabled));
    } catch (e) { setError(String(e)); }
  }

  function openCreate() {
    setEditId(null);
    setNome(""); setCpf(""); setDataNasc(""); setTelefone(""); setTelefoneUrgencia(""); setEmail("");
    setFormaPagamentoId(0); setPlanId(0); setDueDay(5);
    setLaudoFile(null); setLaudoRemoved(false);
    setFotoDataUrl(""); setPhotoChanged(false); setPhotoRemoved(false);
    setError(""); setShowModal(true);
  }

  async function openEdit(s: StudentComPlano) {
    setEditId(s.id);
    setNome(s.nome); setCpf(s.cpf);
    setDataNasc(s.data_nascimento || "");
    setTelefone(s.telefone || ""); setTelefoneUrgencia(s.telefone_urgencia || ""); setEmail(s.email || "");
    setFormaPagamentoId(s.forma_pagamento_id || 0);
    setPlanId(s.student_plan_id ? (plans.find(p => p.name === s.plano_nome)?.id || 0) : 0);
    setDueDay(s.due_day || 5);
    setLaudoFile(null); setLaudoRemoved(false);
    let existing = "";
    try {
      existing = ((await ObterFotoAluno(s.id)) as unknown as string) || "";
    } catch { existing = ""; }
    setFotoDataUrl(existing);
    setPhotoChanged(false); setPhotoRemoved(false);
    setError(""); setShowModal(true);
  }

  async function fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result as string);
      reader.onerror = () => reject(reader.error);
      reader.readAsDataURL(file);
    });
  }

  async function handleCapturarGo() {
    setError("");
    setLoadingCameraGo(true);
    try {
      const dataUrl = await CapturarFotoWebcam();
      if (dataUrl) {
        setFotoDataUrl(dataUrl);
        setPhotoChanged(true);
        setPhotoRemoved(false);
      }
    } catch (e: any) {
      setError("Erro ao capturar da webcam: " + (e?.message || String(e)));
    } finally {
      setLoadingCameraGo(false);
    }
  }

  function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    fileToBase64(file).then((b64) => {
      setFotoDataUrl(b64);
      setPhotoChanged(true);
      setPhotoRemoved(false);
    }).catch((err) => {
      setError("Erro ao carregar arquivo de foto: " + String(err));
    });
    e.target.value = "";
  }

  function clearPhoto() {
    setFotoDataUrl("");
    setPhotoChanged(false);
    if (editId != null) setPhotoRemoved(true);
  }

  async function handleSave() {
    setError("");
    setUploadingLaudo(true);
    try {
      let studentId: number | null = editId;
      if (studentId) {
        await AtualizarStudent(studentId, nome, cpf, dataNasc, telefone, telefoneUrgencia, email, formaPagamentoId);
        await AtualizarPlanoAluno(studentId, planId, formaPagamentoId, dueDay);
      } else {
        const created = (await CriarStudentComPlano(
          nome, cpf, dataNasc, telefone, telefoneUrgencia, email, formaPagamentoId, planId, dueDay
        )) as unknown as StudentComPlano;
        studentId = created.id;
      }

      if (studentId) {
        const student = students.find((st) => st.id === studentId);
        const hadLaudo = !!student?.laudo_medico;
        if (laudoRemoved && hadLaudo) {
          await RemoverLaudoAluno(studentId);
        }
        if (laudoFile) {
          const base64 = await fileToBase64(laudoFile);
          await SalvarLaudoAluno(studentId, base64, laudoFile.name);
        }
        const hadFoto = !!student?.foto;
        if (photoRemoved && hadFoto) {
          await RemoverFotoAluno(studentId);
        }
        if (photoChanged && fotoDataUrl) {
          await SalvarFotoAluno(studentId, fotoDataUrl);
        }
      }

      setShowModal(false);
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setUploadingLaudo(false);
    }
  }

  async function toggleAtivo(s: StudentComPlano) {
    try {
      await AtivarStudent(s.id, !s.ativo);
      await load();
    } catch (e) { setError(String(e)); }
  }

  const currentStudent = editId != null ? students.find((st) => st.id === editId) : undefined;
  const existingLaudoName = currentStudent?.laudo_medico
    ? currentStudent.laudo_medico.split(/[\\/]/).pop() || currentStudent.laudo_medico
    : "";

  const canRemoveLaudo =
    (laudoFile != null) || (!!existingLaudoName && !laudoRemoved);

  const laudoStatusText = laudoFile
    ? "Novo arquivo: " + laudoFile.name
    : (!!existingLaudoName && !laudoRemoved)
      ? "Laudo anexado: " + existingLaudoName
      : (existingLaudoName && laudoRemoved)
        ? "Laudo removido"
        : "";

  const laudoStatusColor = laudoRemoved
    ? "var(--danger)"
    : (laudoFile || existingLaudoName)
      ? "var(--success)"
      : "var(--text-muted)";

  function handleLaudoSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.type && file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
      setError("Apenas arquivos em formato PDF são aceitos.");
      return;
    }
    setLaudoFile(file);
    setLaudoRemoved(false);
    setError("");
    if (laudoInputRef.current) laudoInputRef.current.value = "";
  }

  function handleRemoveLaudo() {
    setLaudoFile(null);
    setLaudoRemoved(true);
  }

  function handleBaixarLaudo(id: number) {
    BaixarLaudo(id).catch((e) => setError(String(e)));
  }

  function formatDate(d: string): string {
    if (!d) return "—";
    return d.split("T")[0] || d;
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <div>
          <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Gestão de Alunos</h2>
          <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
            Visualize, edite e ative matrículas de alunos e seus planos.
          </p>
        </div>
        {canManage && (
          <button onClick={openCreate} className="vmd-btn vmd-btn-primary">
            ➕ Novo Aluno
          </button>
        )}
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ maxWidth: 600 }}>
          <span>⚠️ {error}</span>
        </div>
      )}
      {feedback && (
        <div className={`vmd-alert ${feedback.type === "success" ? "vmd-alert-success" : "vmd-alert-danger"}`} style={{ maxWidth: 600 }}>
          <span>{feedback.type === "success" ? "✨" : "⚠️"} {feedback.msg}</span>
        </div>
      )}

      <div style={{ marginBottom: 20 }}>
        <input
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Buscar aluno por nome..."
          className="vmd-input"
          style={{ width: 320, maxWidth: "100%" }}
        />
        {searchTerm && (
          <span style={{ fontSize: 12, color: "var(--text-muted)", marginLeft: 12 }}>
            {filteredStudents.length} de {students.length} aluno(s)
          </span>
        )}
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {filteredStudents.map((s) => (
          <div
            key={s.id}
            className="vmd-card"
            style={{
              padding: "16px 20px",
              opacity: s.ativo ? 1 : 0.6,
              background: s.ativo ? "var(--bg-card)" : "rgba(255,255,255,0.01)",
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 12 }}>
              <div style={{ flex: 1, minWidth: 200 }}>
                <div style={{ fontSize: 15, fontWeight: 600, display: "flex", alignItems: "center", gap: 10 }}>
                  {s.fotoDataUrl ? (
                    <img
                      src={s.fotoDataUrl}
                      alt={s.nome}
                      style={{ width: 36, height: 36, borderRadius: "50%", objectFit: "cover", flexShrink: 0 }}
                    />
                  ) : (
                    <span
                      style={{
                        width: 36, height: 36, borderRadius: "50%", flexShrink: 0,
                        display: "flex", alignItems: "center", justifyContent: "center",
                        background: "rgba(255,255,255,0.08)", color: "var(--text-muted)",
                        fontSize: 14, fontWeight: 700,
                      }}
                    >
                      {s.nome.trim().charAt(0).toUpperCase() || "?"}
                    </span>
                  )}
                  {s.nome}
                  <span className={`vmd-badge ${s.ativo ? "vmd-badge-success" : "vmd-badge-danger"}`}>
                    {s.ativo ? "Ativo" : "Inativo"}
                  </span>
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>
                  CPF: {s.cpf} {s.data_nascimento ? `· Nasc: ${formatDate(s.data_nascimento)}` : ""}
                </div>
                {s.telefone_urgencia && (
                  <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>
                    Urgência: {s.telefone_urgencia}
                  </div>
                )}
              </div>
              
              <div style={{ display: "flex", gap: 8, flexShrink: 0 }}>
                {s.laudo_medico && (
                  <button
                    onClick={() => handleBaixarLaudo(s.id)}
                    className="vmd-btn vmd-btn-secondary"
                    style={{ padding: "6px 12px", height: 32 }}
                  >
                    📄 Laudo
                  </button>
                )}
                {canManage && (
                  <>
                    <button onClick={() => openEdit(s)} className="vmd-btn vmd-btn-secondary" style={{ padding: "6px 12px", height: 32 }}>
                      Editar
                    </button>
                    <button
                      onClick={() => toggleAtivo(s)}
                      className={`vmd-btn ${s.ativo ? "vmd-btn-danger" : "vmd-btn-success"}`}
                      style={{ padding: "6px 12px", height: 32 }}
                    >
                      {s.ativo ? "Desativar" : "Ativar"}
                    </button>
                  </>
                )}
              </div>
            </div>

            {s.plano_nome && (
              <div
                style={{
                  marginTop: 14,
                  display: "grid",
                  gridTemplateColumns: "repeat(auto-fit, minmax(140px, 1fr))",
                  gap: 16,
                  padding: "12px 16px",
                  background: "rgba(255, 255, 255, 0.02)",
                  border: "1px solid var(--border-color)",
                  borderRadius: "var(--radius-sm)",
                }}
              >
                <div>
                  <div style={{ fontSize: 10, color: "var(--text-dim)", textTransform: "uppercase", fontWeight: 600 }}>Plano</div>
                  <div style={{ fontSize: 13, fontWeight: 500, marginTop: 2, color: "var(--text-main)" }}>
                    {s.plano_nome} {s.plano_preco ? `(${formatCents(s.plano_preco)})` : ""}
                  </div>
                </div>
                <div>
                  <div style={{ fontSize: 10, color: "var(--text-dim)", textTransform: "uppercase", fontWeight: 600 }}>Pagamento</div>
                  <div style={{ fontSize: 13, marginTop: 2, color: "var(--text-main)" }}>{s.payment_method || "—"}</div>
                </div>
                <div>
                  <div style={{ fontSize: 10, color: "var(--text-dim)", textTransform: "uppercase", fontWeight: 600 }}>Vencimento</div>
                  <div style={{ fontSize: 13, marginTop: 2, color: "var(--text-main)" }}>Dia {s.due_day || "—"}</div>
                </div>
                <div>
                  <div style={{ fontSize: 10, color: "var(--text-dim)", textTransform: "uppercase", fontWeight: 600 }}>Status do Plano</div>
                  <div
                    style={{
                      fontSize: 12,
                      marginTop: 2,
                      fontWeight: 600,
                      color: s.plano_status === "active" ? "var(--success)" :
                             s.plano_status === "overdue" ? "var(--danger)" : "var(--text-muted)",
                    }}
                  >
                    {s.plano_status === "active" ? "● Regular" :
                     s.plano_status === "overdue" ? "● Atrasado" :
                     s.plano_status === "suspended" ? "● Suspenso" : s.plano_status || "—"}
                  </div>
                </div>
                <div>
                  <div style={{ fontSize: 10, color: "var(--text-dim)", textTransform: "uppercase", fontWeight: 600 }}>Última Fatura</div>
                  <div style={{ fontSize: 12, marginTop: 2 }}>
                    {s.ultima_fatura ? (
                      <span className={statusBadgeClasses[s.ultima_fatura] || "vmd-badge"}>
                        {statusLabels[s.ultima_fatura] || s.ultima_fatura}
                        {s.fatura_valor ? ` (${formatCents(s.fatura_valor)})` : ""}
                      </span>
                    ) : "Nenhuma"}
                  </div>
                </div>
              </div>
            )}

            {canManage && s.student_plan_id && (
              <div style={{ marginTop: 12, display: "flex", justifyContent: "flex-end" }}>
                {loadingInvoice !== s.student_plan_id ? (
                  <button
                    onClick={() => handleRegistrarPagamento(s.student_plan_id!)}
                    className="vmd-btn vmd-btn-success"
                    style={{ padding: "6px 12px", height: 30, fontSize: 11 }}
                  >
                    💸 Registrar Novo Pagamento
                  </button>
                ) : (
                  <span style={{ fontSize: 12, color: "var(--text-muted)" }}>Gerando fatura...</span>
                )}
              </div>
            )}
          </div>
        ))}
        {filteredStudents.length === 0 && (
          <div className="vmd-card" style={{ textAlign: "center", padding: 48, color: "var(--text-muted)" }}>
            {searchTerm ? "Nenhum aluno encontrado para esta busca." : "Nenhum aluno cadastrado."}
          </div>
        )}
      </div>

      {showModal && (
        <div className="vmd-modal-overlay" onClick={() => setShowModal(false)}>
          <div className="vmd-modal-content" style={{ maxWidth: 540 }} onClick={(e) => e.stopPropagation()}>
            <div className="vmd-modal-header">
              <h3 style={{ fontSize: 16, margin: 0 }}>{editId ? "Editar Cadastro de Aluno" : "Matricular Novo Aluno"}</h3>
              <button onClick={() => setShowModal(false)} className="vmd-btn vmd-btn-ghost" style={{ padding: 4, minWidth: "auto" }}>
                ✕
              </button>
            </div>

            <div className="vmd-modal-body">
              <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
                <div className="vmd-form-group" style={{ marginBottom: 0, display: "flex", flexDirection: "column", alignItems: "center", gap: 10 }}>
                  <label className="vmd-label" style={{ alignSelf: "center" }}>Foto do Aluno</label>
                  <div
                    style={{
                      width: 360, height: 360, borderRadius: 12,
                      background: "rgba(255,255,255,0.04)",
                      border: "1px solid var(--border-color)",
                      overflow: "hidden",
                      display: "flex", alignItems: "center", justifyContent: "center",
                      position: "relative",
                    }}
                  >
                    {fotoDataUrl ? (
                      <img src={fotoDataUrl} alt="Foto do aluno" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                    ) : (
                      <div style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: 6, color: "var(--text-dim)" }}>
                        <span style={{ fontSize: 38, lineHeight: 1 }}>📷</span>
                        <span style={{ fontSize: 11 }}>Sem foto</span>
                      </div>
                    )}
                  </div>

                  <input
                    ref={fotoInputRef}
                    type="file"
                    accept="image/*"
                    onChange={handleFileSelect}
                    style={{ display: "none" }}
                  />

                  <div style={{ display: "flex", gap: 8, flexWrap: "wrap", justifyContent: "center" }}>
                    <button
                      type="button"
                      onClick={handleCapturarGo}
                      disabled={loadingCameraGo}
                      className="vmd-btn vmd-btn-primary"
                      style={{ padding: "6px 16px", height: 34, fontWeight: 500 }}
                      title="Capturar foto diretamente da webcam conectada (hardware)"
                    >
                      {loadingCameraGo ? "⏳ Acessando Webcam..." : "📸 Tirar Foto"}
                    </button>

                    <button
                      type="button"
                      onClick={() => fotoInputRef.current?.click()}
                      className="vmd-btn vmd-btn-secondary"
                      style={{ padding: "6px 14px", height: 34 }}
                      title="Selecionar foto já existente no computador"
                    >
                      📁 Escolher Arquivo
                    </button>

                    {fotoDataUrl && (
                      <button
                        type="button"
                        onClick={clearPhoto}
                        className="vmd-btn vmd-btn-danger"
                        style={{ padding: "6px 12px", height: 34, background: "transparent" }}
                        title="Remover foto atual"
                      >
                        🗑️ Remover
                      </button>
                    )}
                  </div>
                  <div style={{ fontSize: 11, color: "var(--text-dim)", textAlign: "center" }}>
                    Clique em "Tirar Foto" para capturar da webcam conectada ou "Escolher Arquivo" para usar uma foto existente.
                  </div>
                </div>

                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Nome Completo</label>
                  <input
                    value={nome}
                    onChange={(e) => setNome(e.target.value)}
                    placeholder="Nome completo do aluno"
                    className="vmd-input"
                  />
                </div>

                <div className="vmd-grid-2">
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">CPF</label>
                    <input
                      value={cpf}
                      onChange={(e) => setCpf(maskCPF(e.target.value))}
                      placeholder="000.000.000-00"
                      maxLength={14}
                      className="vmd-input"
                    />
                  </div>
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Data de Nascimento</label>
                    <input
                      type="date"
                      value={dataNasc}
                      onChange={(e) => setDataNasc(e.target.value)}
                      className="vmd-input"
                      style={{ colorScheme: "dark" }}
                    />
                  </div>
                </div>

                <div className="vmd-grid-2">
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Telefone / WhatsApp</label>
                    <input
                      value={telefone}
                      onChange={(e) => setTelefone(maskTel(e.target.value))}
                      placeholder="(00) 90000-0000"
                      maxLength={15}
                      className="vmd-input"
                    />
                  </div>
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">E-mail</label>
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="aluno@exemplo.com"
                      className="vmd-input"
                    />
                  </div>
                </div>

                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Telefone de Urgência</label>
                  <input
                    value={telefoneUrgencia}
                    onChange={(e) => setTelefoneUrgencia(maskTel(e.target.value))}
                    placeholder="(00) 90000-0000"
                    maxLength={15}
                    className="vmd-input"
                  />
                </div>

                <div style={{ borderTop: "1px solid var(--border-color)", margin: "8px 0" }} />

                <div className="vmd-grid-2">
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Plano</label>
                    <select
                      value={planId}
                      onChange={(e) => setPlanId(Number(e.target.value))}
                      className="vmd-select"
                      style={{ colorScheme: "dark" }}
                    >
                      <option value={0}>Sem plano ativo</option>
                      {plans.map((pl) => (
                        <option key={pl.id} value={pl.id}>
                          {pl.name} — {formatCents(pl.price_cents)}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Forma de Pagamento</label>
                    <select
                      value={formaPagamentoId}
                      onChange={(e) => setFormaPagamentoId(Number(e.target.value))}
                      className="vmd-select"
                      style={{ colorScheme: "dark" }}
                    >
                      <option value={0}>Selecione...</option>
                      {methods.map((m) => (
                        <option key={m.id} value={m.id}>{m.name}</option>
                      ))}
                    </select>
                  </div>
                </div>

                <div className="vmd-form-group" style={{ marginBottom: 0, width: "50%" }}>
                  <label className="vmd-label">Dia do Vencimento Mensal</label>
                  <input
                    type="number"
                    min={1}
                    max={28}
                    value={dueDay}
                    onChange={(e) => setDueDay(Number(e.target.value))}
                    className="vmd-input"
                  />
                </div>

                <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                  <label className="vmd-label">Laudo do Aluno</label>
                  <div style={{ display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
                    <input
                      ref={laudoInputRef}
                      type="file"
                      accept=".pdf,application/pdf"
                      onChange={handleLaudoSelect}
                      style={{ display: "none" }}
                    />
                    <button
                      type="button"
                      onClick={() => laudoInputRef.current?.click()}
                      className="vmd-btn vmd-btn-secondary"
                      style={{ padding: "6px 12px", height: 32 }}
                    >
                      📎 Selecionar Laudo PDF
                    </button>
                    {laudoStatusText && (
                      <span style={{ fontSize: 12, color: laudoStatusColor }}>{laudoStatusText}</span>
                    )}
                    {canRemoveLaudo && (
                      <button
                        type="button"
                        onClick={handleRemoveLaudo}
                        className="vmd-btn vmd-btn-danger"
                        style={{ padding: "6px 12px", height: 32, background: "transparent" }}
                      >
                        Remover
                      </button>
                    )}
                  </div>
                  {uploadingLaudo && (
                    <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 6 }}>Enviando laudo...</div>
                  )}
                  <div style={{ fontSize: 11, color: "var(--text-dim)", marginTop: 6 }}>
                    Apenas arquivos em formato PDF (ex.: liberação para atividade física).
                  </div>
                </div>
              </div>
            </div>

            <div className="vmd-modal-footer">
              <button onClick={() => setShowModal(false)} className="vmd-btn vmd-btn-secondary">
                Cancelar
              </button>
              <button onClick={handleSave} className="vmd-btn vmd-btn-primary">
                {editId ? "Salvar Alterações" : "Efetivar Matrícula"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Alunos;

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AdicionarAgendamento,
  AdicionarAgendamentoIndividual,
  AtualizarAgendamentoIndividual,
  DefinirCapacidadeDia,
  ListarAgendamentosDia,
  ListarAgendamentosIndividuaisDia,
  ListarAgendamentosIndividuaisMes,
  ListarAgendamentosMes,
  ListarCapacidadesMes,
  ListarStudents,
  ObterConfigAgendaIndividual,
  RemoverAgendamento,
  RemoverAgendamentoIndividual,
} from "../../wailsjs/go/main/App";
import "./Agenda.css";
import "./AgendaIndividual.css";

const CAPACIDADE_PADRAO = 10;

const TURNOS = [
  { key: "manha", label: "Manhã", horario: "06:00 às 11:00" },
  { key: "tarde", label: "Tarde", horario: "15:00 às 19:00" },
  { key: "noite", label: "Noite", horario: "19:00 às 20:00" },
];

const HORARIOS_IND = [
  { hora: "06:00", turno: "Manhã" },
  { hora: "07:00", turno: "Manhã" },
  { hora: "08:00", turno: "Manhã" },
  { hora: "09:00", turno: "Manhã" },
  { hora: "10:00", turno: "Manhã" },
  { hora: "15:00", turno: "Tarde" },
  { hora: "16:00", turno: "Tarde" },
  { hora: "17:00", turno: "Tarde" },
  { hora: "18:00", turno: "Tarde" },
  { hora: "19:00", turno: "Noite" },
];

const TURNOS_IND = [
  { key: "Manhã", label: "Manhã", horario: "06:00 às 11:00" },
  { key: "Tarde", label: "Tarde", horario: "15:00 às 19:00" },
  { key: "Noite", label: "Noite", horario: "19:00 às 20:00" },
];

const MENU_ITEMS = [
  { key: "pilates", label: "Pilates", icon: "🧘", color: "var(--primary)" },
  { key: "ventosaterapia", label: "Ventosaterapia", icon: "🟣", color: "#a78bfa" },
  { key: "liberacao_miofascial", label: "Liberação Miofascial", icon: "🔵", color: "#22d3ee" },
  { key: "personal_trainer", label: "Personal Trainer", icon: "🟡", color: "#fbbf24" },
  { key: "kinesio_tape", label: "Kinesio Tape", icon: "🩷", color: "#f472b6" },
];

const PRATICAS = [
  { key: "ventosaterapia", label: "Ventosaterapia" },
  { key: "liberacao_miofascial", label: "Liberação Miofascial" },
  { key: "personal_trainer", label: "Personal Trainer" },
  { key: "kinesio_tape", label: "Kinesio Tape" },
];

const MESES = [
  "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
  "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
];

const DIAS_SEMANA = ["Dom", "Seg", "Ter", "Qua", "Qui", "Sex", "Sáb"];

interface Student {
  id: number;
  nome: string;
  ativo: boolean;
}

interface Agendamento {
  id: number;
  data: string;
  student_id?: number;
  nome: string;
  turno: string;
  telefone?: string;
  observacao?: string;
  created_at: string;
}

interface AgendamentoIndividual {
  id: number;
  data: string;
  hora: string;
  student_id?: number;
  nome: string;
  pratica: string;
  observacao?: string;
  created_at: string;
}

function hojeStr(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function formatBR(data: string): string {
  const [y, m, d] = data.split("-");
  return `${d}/${m}/${y}`;
}

function turnoInfo(key: string) {
  return TURNOS.find((t) => t.key === key) || { key, label: key, horario: "" };
}

function praticaColor(key: string): string {
  const map: Record<string, string> = {
    ventosaterapia: "#a78bfa",
    liberacao_miofascial: "#22d3ee",
    personal_trainer: "#fbbf24",
    kinesio_tape: "#f472b6",
  };
  return map[key] || "#9ca3af";
}

function ocupacaoClass(total: number, capacidade: number): string {
  if (total === 0) return "";
  if (total <= 3) return "agenda-day-verde";
  if (total < capacidade) return "agenda-day-laranja";
  return "agenda-day-vermelho";
}

function ocupacaoLabel(total: number, capacidade: number): { label: string; color: string; bg: string } {
  if (total === 0) return { label: "Livre", color: "#6b7280", bg: "rgba(255,255,255,0.05)" };
  if (total <= 3) return { label: "Livre", color: "#34d399", bg: "rgba(16,185,129,0.15)" };
  if (total < capacidade) return { label: "Média Ocupação", color: "#fbbf24", bg: "rgba(245,158,11,0.15)" };
  return { label: "Lotado", color: "#f87171", bg: "rgba(239,68,68,0.18)" };
}

function parsePraticas(s: string): string[] {
  if (!s) return [];
  return [...new Set(s.split(","))];
}

function makeDias(ano: number, mes: number): (string | null)[] {
  const firstDay = new Date(ano, mes, 1).getDay();
  const daysInMonth = new Date(ano, mes + 1, 0).getDate();
  const cells: (string | null)[] = [];
  for (let i = 0; i < firstDay; i++) cells.push(null);
  for (let d = 1; d <= daysInMonth; d++) {
    cells.push(`${ano}-${String(mes + 1).padStart(2, "0")}-${String(d).padStart(2, "0")}`);
  }
  return cells;
}

// ======================== MAIN CALENDAR (PILATES) ========================

function CalendarioPilates({
  students,
  loadStudents,
}: {
  students: Student[];
  loadStudents: () => void;
}) {
  const today = hojeStr();
  const now = new Date();
  const [ano, setAno] = useState(now.getFullYear());
  const [mes, setMes] = useState(now.getMonth());
  const [ocupacao, setOcupacao] = useState<Record<string, { total: number; capacidade: number }>>({});
  const [loadingMes, setLoadingMes] = useState(false);

  const [selectedDay, setSelectedDay] = useState<string | null>(null);
  const [agendamentos, setAgendamentos] = useState<Agendamento[]>([]);
  const [loadingDia, setLoadingDia] = useState(false);

  const [capacidadeDia, setCapacidadeDia] = useState(CAPACIDADE_PADRAO);
  const [editCapacidade, setEditCapacidade] = useState(false);
  const [novaCapacidade, setNovaCapacidade] = useState(String(CAPACIDADE_PADRAO));
  const [savingCapacidade, setSavingCapacidade] = useState(false);

  const [modoNovo, setModoNovo] = useState(false);
  const [alunoId, setAlunoId] = useState(0);
  const [novoNome, setNovoNome] = useState("");
  const [turno, setTurno] = useState("manha");
  const [error, setError] = useState("");
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; msg: string } | null>(null);

  const loadMes = useCallback(async (a: number, m: number) => {
    setLoadingMes(true);
    try {
      const [list, caps] = await Promise.all([
        ListarAgendamentosMes(a, m + 1) as unknown as Promise<{ data: string; total: number; capacidade: number }[]>,
        ListarCapacidadesMes(a, m + 1) as unknown as Promise<{ data: string; capacidade: number }[]>,
      ]);
      const map: Record<string, { total: number; capacidade: number }> = {};
      for (const d of caps) map[d.data] = { total: 0, capacidade: d.capacidade };
      for (const d of list) map[d.data] = { total: d.total, capacidade: d.capacidade || CAPACIDADE_PADRAO };
      setOcupacao(map);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoadingMes(false);
    }
  }, []);

  useEffect(() => { loadMes(ano, mes); }, [ano, mes, loadMes]);

  const loadDia = useCallback(async (data: string) => {
    setLoadingDia(true);
    try {
      const list = await ListarAgendamentosDia(data) as unknown as Agendamento[];
      setAgendamentos(list || []);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoadingDia(false);
    }
  }, []);

  function navegar(delta: number) {
    const d = new Date(ano, mes + delta, 1);
    setAno(d.getFullYear());
    setMes(d.getMonth());
  }

  function abrirDia(data: string) {
    setSelectedDay(data);
    setError("");
    setFeedback(null);
    setModoNovo(false);
    setAlunoId(0);
    setNovoNome("");
    setTurno("manha");
    setEditCapacidade(false);
    setNovaCapacidade(String(ocupacao[data]?.capacidade ?? CAPACIDADE_PADRAO));
    setCapacidadeDia(ocupacao[data]?.capacidade ?? CAPACIDADE_PADRAO);
    loadDia(data);
    if (students.length === 0) loadStudents();
  }

  async function handleSalvarCapacidade() {
    if (!selectedDay) return;
    const cap = Number(novaCapacidade);
    if (Number.isNaN(cap) || cap < 1) { setError("Capacidade inválida"); return; }
    setSavingCapacidade(true);
    setError("");
    try {
      await DefinirCapacidadeDia(selectedDay, cap);
      setCapacidadeDia(cap);
      setEditCapacidade(false);
      setFeedback({ type: "success", msg: `Capacidade atualizada para ${cap} vagas.` });
      await loadMes(ano, mes);
    } catch (e) {
      setError(String(e));
    } finally {
      setSavingCapacidade(false);
    }
  }

  const totalDia = selectedDay ? agendamentos.length : 0;
  const lotado = totalDia >= capacidadeDia;

  async function handleAdicionar() {
    if (!selectedDay || lotado) return;
    setError("");
    setFeedback(null);
    if (modoNovo && !novoNome.trim()) { setError("Informe o nome do aluno"); return; }
    if (!modoNovo && alunoId === 0) { setError("Selecione um aluno ou alterne para 'Novo aluno'"); return; }
    try {
      await AdicionarAgendamento(selectedDay, alunoId, novoNome.trim(), turno);
      setFeedback({ type: "success", msg: "Vaga reservada com sucesso!" });
      await Promise.all([loadDia(selectedDay), loadMes(ano, mes)]);
      setNovoNome("");
      setAlunoId(0);
    } catch (e) {
      setError(String(e));
    }
  }

  async function handleRemover(id: number, nome: string) {
    if (!window.confirm(`Remover o agendamento de ${nome}?`)) return;
    try {
      await RemoverAgendamento(id);
      setFeedback({ type: "success", msg: "Agendamento removido." });
      if (selectedDay) await Promise.all([loadDia(selectedDay), loadMes(ano, mes)]);
    } catch (e) {
      setError(String(e));
    }
  }

  const dias = useMemo(() => makeDias(ano, mes), [ano, mes]);
  const vagas = ocupacaoLabel(totalDia, capacidadeDia);

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <h3 style={{ fontSize: 17, fontWeight: 700, margin: 0 }}>Calendário Principal — Pilates</h3>
        <p style={{ color: "var(--text-muted)", fontSize: 12, marginTop: 2 }}>
          Capacidade padrão de {CAPACIDADE_PADRAO} alunos por dia (editável).
        </p>
      </div>

      {error && <div className="vmd-alert vmd-alert-danger"><span>⚠️ {error}</span></div>}
      {feedback && <div className={`vmd-alert ${feedback.type === "success" ? "vmd-alert-success" : "vmd-alert-danger"}`}><span>✨ {feedback.msg}</span></div>}

      <div className="vmd-card">
        <div className="agenda-header">
          <div className="agenda-nav">
            <button className="agenda-nav-btn" onClick={() => navegar(-1)}>‹</button>
            <div className="agenda-nav-title">{MESES[mes]} {ano}</div>
            <button className="agenda-nav-btn" onClick={() => navegar(1)}>›</button>
            <button className="vmd-btn vmd-btn-ghost" onClick={() => { setAno(now.getFullYear()); setMes(now.getMonth()); }}>Hoje</button>
          </div>
          {loadingMes && <span style={{ fontSize: 12, color: "var(--text-dim)" }}>Carregando...</span>}
        </div>

        {loadingMes && <span style={{ fontSize: 12, color: "var(--text-dim)" }}>Carregando...</span>}

        <div className="agenda-weekdays">
          {DIAS_SEMANA.map((d) => <div key={d} className="agenda-weekday">{d}</div>)}
        </div>

        <div className="agenda-grid">
          {dias.map((data, idx) => {
            if (data === null) return <div key={`v-${idx}`} className="agenda-vazio" />;
            const dia = ocupacao[data] || { total: 0, capacidade: CAPACIDADE_PADRAO };
            return (
              <button
                key={data}
                className={`agenda-day ${ocupacaoClass(dia.total, dia.capacidade)} ${data === today ? "hoje" : ""} ${data === selectedDay ? "selecionado" : ""}`}
                onClick={() => abrirDia(data)}
              >
                <span className="agenda-day-num">{Number(data.split("-")[2])}</span>
                <span className="agenda-day-count">{dia.total}/{dia.capacidade} vagas</span>
              </button>
            );
          })}
        </div>
      </div>

      {selectedDay && (
        <div className="vmd-modal-overlay" onClick={() => setSelectedDay(null)}>
          <div className="vmd-modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 560 }}>
            <div className="vmd-modal-header">
              <div>
                <h3 style={{ fontSize: 16, margin: 0 }}>{formatBR(selectedDay)}</h3>
                <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
                  {DIAS_SEMANA[new Date(ano, mes, Number(selectedDay.split("-")[2])).getDay()]}
                </span>
              </div>
              <button onClick={() => setSelectedDay(null)} className="vmd-btn vmd-btn-ghost" style={{ padding: 4, minWidth: "auto" }}>✕</button>
            </div>

            <div className="vmd-modal-body">
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16 }}>
                <span className="agenda-vagas-pill" style={{ color: vagas.color, background: vagas.bg }}>
                  {totalDia}/{capacidadeDia} vagas — {vagas.label}
                </span>
                <span style={{ fontSize: 11, color: "var(--text-dim)" }}>
                  {loadingDia ? "Carregando..." : `${agendamentos.length} agendamento(s)`}
                </span>
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 16, padding: "10px 12px", border: "1px solid var(--border-color)", borderRadius: "var(--radius-sm)", background: "rgba(255,255,255,0.02)" }}>
                <span style={{ fontSize: 12, fontWeight: 600, color: "var(--text-main)" }}>Capacidade:</span>
                {editCapacidade ? (
                  <>
                    <input type="number" min={1} value={novaCapacidade} onChange={(e) => setNovaCapacidade(e.target.value)} className="vmd-input" style={{ width: 90, padding: "6px 10px", fontSize: 13 }} />
                    <button onClick={handleSalvarCapacidade} disabled={savingCapacidade} className="vmd-btn vmd-btn-primary" style={{ padding: "6px 12px", fontSize: 12 }}>{savingCapacidade ? "Salvando..." : "Salvar"}</button>
                    <button onClick={() => { setEditCapacidade(false); setNovaCapacidade(String(capacidadeDia)); }} className="vmd-btn vmd-btn-ghost" style={{ padding: "6px 10px", fontSize: 12 }}>Cancelar</button>
                  </>
                ) : (
                  <>
                    <span style={{ fontSize: 13, fontWeight: 600, color: vagas.color }}>{capacidadeDia} vagas/dia</span>
                    <button onClick={() => { setEditCapacidade(true); setNovaCapacidade(String(capacidadeDia)); }} className="vmd-btn vmd-btn-ghost" style={{ padding: "6px 10px", fontSize: 12 }}>✏️ Editar</button>
                  </>
                )}
              </div>

              {agendamentos.length === 0 ? (
                <div style={{ textAlign: "center", padding: "24px 0", color: "var(--text-dim)", fontSize: 13 }}>Nenhum aluno agendado neste dia.</div>
              ) : (
                <div className="agenda-lista">
                  {agendamentos.map((ag) => (
                    <div key={ag.id} className="agenda-item">
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div className="agenda-item-nome">{ag.nome}</div>
                        <div className="agenda-item-sub">{ag.student_id ? "Aluno cadastrado" : "Novo aluno"}{ag.telefone ? ` · ${ag.telefone}` : ""}</div>
                      </div>
                      <span className="agenda-turno-badge">{turnoInfo(ag.turno).label} · {turnoInfo(ag.turno).horario}</span>
                      <button onClick={() => handleRemover(ag.id, ag.nome)} className="vmd-btn vmd-btn-danger" style={{ padding: "5px 10px", fontSize: 11 }}>Remover</button>
                    </div>
                  ))}
                </div>
              )}

              <div style={{ borderTop: "1px solid var(--border-color)", margin: "18px 0 14px", paddingTop: 16 }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: "var(--text-main)", marginBottom: 12 }}>
                  {lotado ? "Dia lotado" : "Adicionar aluno"}
                </div>
                {lotado ? (
                  <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 0 }}>⚠️ Capacidade máxima de {capacidadeDia} alunos atingida.</div>
                ) : (
                  <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
                    <div className="agenda-seg-toggle">
                      <button className={`agenda-seg-btn ${!modoNovo ? "ativo" : ""}`} onClick={() => setModoNovo(false)}>🎓 Aluno cadastrado</button>
                      <button className={`agenda-seg-btn ${modoNovo ? "ativo" : ""}`} onClick={() => setModoNovo(true)}>🆕 Novo aluno</button>
                    </div>
                    {modoNovo ? (
                      <div>
                        <label className="vmd-label">Nome do aluno</label>
                        <input value={novoNome} onChange={(e) => setNovoNome(e.target.value)} placeholder="Ex: Pedro Silva" className="vmd-input" />
                      </div>
                    ) : (
                      <div>
                        <label className="vmd-label">Selecionar aluno</label>
                        <select value={alunoId} onChange={(e) => setAlunoId(Number(e.target.value))} className="vmd-select">
                          <option value={0}>— Escolha um aluno —</option>
                          {students.filter((s) => s.ativo).map((s) => <option key={s.id} value={s.id}>{s.nome}</option>)}
                        </select>
                      </div>
                    )}
                    <div>
                      <label className="vmd-label">Turno</label>
                      <div style={{ display: "flex", gap: 6 }}>
                        {TURNOS.map((t) => (
                          <button key={t.key} className={`agenda-seg-btn ${turno === t.key ? "ativo" : ""}`} style={{ flex: 1 }} onClick={() => setTurno(t.key)}>
                            {t.label}
                            <span style={{ display: "block", fontSize: 10, opacity: 0.75, marginTop: 2 }}>{t.horario}</span>
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>

            <div className="vmd-modal-footer">
              <button onClick={() => setSelectedDay(null)} className="vmd-btn vmd-btn-secondary">Fechar</button>
              {!lotado && <button onClick={handleAdicionar} className="vmd-btn vmd-btn-primary">➕ Reservar vaga</button>}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ======================== INDIVIDUAL PRACTICE CALENDAR ========================

function CalendarioPratica({
  praticaKey,
  praticaNome,
  students,
  loadStudents,
}: {
  praticaKey: string;
  praticaNome: string;
  students: Student[];
  loadStudents: () => void;
}) {
  const today = hojeStr();
  const now = new Date();
  const [ano, setAno] = useState(now.getFullYear());
  const [mes, setMes] = useState(now.getMonth());
  const [ocupacao, setOcupacao] = useState<Record<string, { total: number; capacidade: number; praticas: string }>>({});
  const [loadingMes, setLoadingMes] = useState(false);

  const [selectedDay, setSelectedDay] = useState<string | null>(null);
  const [atendimentos, setAtendimentos] = useState<AgendamentoIndividual[]>([]);
  const [loadingDia, setLoadingDia] = useState(false);

  const [limite, setLimite] = useState(CAPACIDADE_PADRAO);

  const [capacidadeDia, setCapacidadeDia] = useState(CAPACIDADE_PADRAO);
  const [editCapacidade, setEditCapacidade] = useState(false);
  const [novaCapacidade, setNovaCapacidade] = useState(String(CAPACIDADE_PADRAO));
  const [savingCapacidade, setSavingCapacidade] = useState(false);

  const [modoEdicao, setModoEdicao] = useState<number | null>(null);
  const [alunoId, setAlunoId] = useState(0);
  const [novoNome, setNovoNome] = useState("");
  const [hora, setHora] = useState("06:00");
  const [observacao, setObservacao] = useState("");
  const [error, setError] = useState("");
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; msg: string } | null>(null);

  const loadMes = useCallback(async (a: number, m: number) => {
    setLoadingMes(true);
    try {
      const list = await ListarAgendamentosIndividuaisMes(a, m + 1) as unknown as { data: string; total: number; capacidade: number; praticas: string }[];
      const map: Record<string, { total: number; capacidade: number; praticas: string }> = {};
      for (const d of list) map[d.data] = { total: d.total, capacidade: d.capacidade, praticas: d.praticas || "" };
      setOcupacao(map);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoadingMes(false);
    }
  }, []);

  const loadConfig = useCallback(async () => {
    try {
      const cfg = await ObterConfigAgendaIndividual() as unknown as { limite_diario: number };
      setLimite(cfg.limite_diario);
    } catch { /* keep default */ }
  }, []);

  useEffect(() => { loadMes(ano, mes); }, [ano, mes, loadMes]);
  useEffect(() => { loadConfig(); }, [loadConfig]);

  const loadDia = useCallback(async (data: string) => {
    setLoadingDia(true);
    try {
      const list = await ListarAgendamentosIndividuaisDia(data) as unknown as AgendamentoIndividual[];
      setAtendimentos((list || []).filter((a) => a.pratica === praticaKey));
    } catch (e) {
      setError(String(e));
    } finally {
      setLoadingDia(false);
    }
  }, [praticaKey]);

  function navegar(delta: number) {
    const d = new Date(ano, mes + delta, 1);
    setAno(d.getFullYear());
    setMes(d.getMonth());
  }

  function abrirDia(data: string) {
    setSelectedDay(data);
    setError("");
    setFeedback(null);
    resetForm();
    const cap = ocupacao[data]?.capacidade ?? limite;
    setCapacidadeDia(cap);
    setNovaCapacidade(String(cap));
    setEditCapacidade(false);
    loadDia(data);
    if (students.length === 0) loadStudents();
  }

  function resetForm() {
    setModoEdicao(null);
    setAlunoId(0);
    setNovoNome("");
    setHora("06:00");
    setObservacao("");
  }

  async function handleSalvarCapacidade() {
    if (!selectedDay) return;
    const cap = Number(novaCapacidade);
    if (Number.isNaN(cap) || cap < 1 || cap > CAPACIDADE_PADRAO) { setError("Capacidade inválida"); return; }
    setSavingCapacidade(true);
    setError("");
    try {
      await DefinirCapacidadeDia(selectedDay, cap);
      setCapacidadeDia(cap);
      setEditCapacidade(false);
      setFeedback({ type: "success", msg: `Capacidade atualizada para ${cap} vagas.` });
      await loadMes(ano, mes);
    } catch (e) {
      setError(String(e));
    } finally {
      setSavingCapacidade(false);
    }
  }

  const horasDisponiveis = useMemo(() => {
    const ocupadas = new Set(atendimentos.map((a) => a.hora));
    return HORARIOS_IND.filter((h) => !ocupadas.has(h.hora) || (modoEdicao !== null && h.hora === atendimentos.find((a) => a.id === modoEdicao)?.hora));
  }, [atendimentos, modoEdicao]);

  const totalDia = selectedDay ? atendimentos.length : 0;
  const lotado = totalDia >= capacidadeDia;

  function iniciarEdicao(ag: AgendamentoIndividual) {
    setModoEdicao(ag.id);
    setAlunoId(ag.student_id || 0);
    setNovoNome(ag.student_id ? "" : ag.nome);
    setHora(ag.hora);
    setObservacao(ag.observacao || "");
    setError("");
    setFeedback(null);
  }

  async function handleSalvar() {
    if (!selectedDay) return;
    setError("");
    setFeedback(null);

    if (modoEdicao !== null) {
      try {
        await AtualizarAgendamentoIndividual(modoEdicao, hora, praticaKey, observacao);
        setFeedback({ type: "success", msg: "Atendimento atualizado!" });
        resetForm();
        await Promise.all([loadDia(selectedDay), loadMes(ano, mes)]);
      } catch (e) { setError(String(e)); }
      return;
    }

    if (lotado) { setError(`Limite de ${capacidadeDia} atendimentos de ${praticaNome} atingido`); return; }
    if (alunoId === 0 && !novoNome.trim()) { setError("Selecione um aluno ou informe o nome"); return; }

    try {
      await AdicionarAgendamentoIndividual(selectedDay, alunoId, novoNome.trim(), hora, praticaKey, observacao);
      setFeedback({ type: "success", msg: "Atendimento agendado!" });
      resetForm();
      await Promise.all([loadDia(selectedDay), loadMes(ano, mes)]);
    } catch (e) { setError(String(e)); }
  }

  async function handleRemover(id: number, nome: string) {
    if (!window.confirm(`Remover atendimento de ${nome}?`)) return;
    try {
      await RemoverAgendamentoIndividual(id);
      setFeedback({ type: "success", msg: "Atendimento removido." });
      if (modoEdicao === id) resetForm();
      if (selectedDay) await Promise.all([loadDia(selectedDay), loadMes(ano, mes)]);
    } catch (e) { setError(String(e)); }
  }

  const dias = useMemo(() => makeDias(ano, mes), [ano, mes]);
  const vagas = ocupacaoLabel(totalDia, capacidadeDia);

  return (
    <div>
      <div style={{ marginBottom: 16, display: "flex", alignItems: "center", gap: 10 }}>
        <span className="agi-pratica-dot" style={{ background: praticaColor(praticaKey), width: 12, height: 12, borderRadius: "50%", flexShrink: 0 }} />
        <h3 style={{ fontSize: 17, fontWeight: 700, margin: 0 }}>{praticaNome}</h3>
      </div>

      {error && <div className="vmd-alert vmd-alert-danger"><span>⚠️ {error}</span></div>}
      {feedback && <div className={`vmd-alert ${feedback.type === "success" ? "vmd-alert-success" : "vmd-alert-danger"}`}><span>✨ {feedback.msg}</span></div>}

      <div className="vmd-card">
        <div className="agenda-header">
          <div className="agenda-nav">
            <button className="agenda-nav-btn" onClick={() => navegar(-1)}>‹</button>
            <div className="agenda-nav-title">{MESES[mes]} {ano}</div>
            <button className="agenda-nav-btn" onClick={() => navegar(1)}>›</button>
            <button className="vmd-btn vmd-btn-ghost" onClick={() => { setAno(now.getFullYear()); setMes(now.getMonth()); }}>Hoje</button>
          </div>
          {loadingMes && <span style={{ fontSize: 12, color: "var(--text-dim)" }}>Carregando...</span>}
        </div>

        <div className="agenda-weekdays">
          {DIAS_SEMANA.map((d) => <div key={d} className="agenda-weekday">{d}</div>)}
        </div>

        <div className="agenda-grid">
          {dias.map((data, idx) => {
            if (data === null) return <div key={`v-${idx}`} className="agenda-vazio" />;
            const dia = ocupacao[data] || { total: 0, capacidade: limite, praticas: "" };
            const dayPraticas = parsePraticas(dia.praticas);
            const thisCount = dayPraticas.includes(praticaKey) ? dia.total : 0;
            const oClass = ocupacaoClass(thisCount, dia.capacidade);
            return (
              <button
                key={data}
                className={`agenda-day ${oClass} ${data === today ? "hoje" : ""} ${data === selectedDay ? "selecionado" : ""}`}
                onClick={() => abrirDia(data)}
              >
                <span className="agenda-day-num">{Number(data.split("-")[2])}</span>
                {thisCount > 0 && (
                  <span className="agenda-day-count">{thisCount}/{dia.capacidade} vagas</span>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {selectedDay && (
        <div className="vmd-modal-overlay" onClick={() => setSelectedDay(null)}>
          <div className="vmd-modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 560 }}>
            <div className="vmd-modal-header">
              <div>
                <h3 style={{ fontSize: 16, margin: 0 }}>{formatBR(selectedDay)}</h3>
                <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
                  {DIAS_SEMANA[new Date(ano, mes, Number(selectedDay.split("-")[2])).getDay()]}
                </span>
              </div>
              <button onClick={() => setSelectedDay(null)} className="vmd-btn vmd-btn-ghost" style={{ padding: 4, minWidth: "auto" }}>✕</button>
            </div>

            <div className="vmd-modal-body">
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16 }}>
                <span className="agenda-vagas-pill" style={{ color: vagas.color, background: vagas.bg }}>
                  {totalDia}/{capacidadeDia} vagas — {vagas.label}
                </span>
                <span style={{ fontSize: 11, color: "var(--text-dim)" }}>
                  {loadingDia ? "Carregando..." : `${atendimentos.length} atendimento(s)`}
                </span>
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 16, padding: "10px 12px", border: "1px solid var(--border-color)", borderRadius: "var(--radius-sm)", background: "rgba(255,255,255,0.02)" }}>
                <span style={{ fontSize: 12, fontWeight: 600, color: "var(--text-main)" }}>Capacidade:</span>
                {editCapacidade ? (
                  <>
                    <input type="number" min={1} max={CAPACIDADE_PADRAO} value={novaCapacidade} onChange={(e) => setNovaCapacidade(e.target.value)} className="vmd-input" style={{ width: 90, padding: "6px 10px", fontSize: 13 }} />
                    <button onClick={handleSalvarCapacidade} disabled={savingCapacidade} className="vmd-btn vmd-btn-primary" style={{ padding: "6px 12px", fontSize: 12 }}>{savingCapacidade ? "Salvando..." : "Salvar"}</button>
                    <button onClick={() => { setEditCapacidade(false); setNovaCapacidade(String(capacidadeDia)); }} className="vmd-btn vmd-btn-ghost" style={{ padding: "6px 10px", fontSize: 12 }}>Cancelar</button>
                  </>
                ) : (
                  <>
                    <span style={{ fontSize: 13, fontWeight: 600, color: vagas.color }}>{capacidadeDia} vagas/dia</span>
                    <button onClick={() => { setEditCapacidade(true); setNovaCapacidade(String(capacidadeDia)); }} className="vmd-btn vmd-btn-ghost" style={{ padding: "6px 10px", fontSize: 12 }}>✏️ Editar</button>
                  </>
                )}
              </div>

              {atendimentos.length === 0 ? (
                <div style={{ textAlign: "center", padding: "24px 0", color: "var(--text-dim)", fontSize: 13 }}>Nenhum atendimento neste dia.</div>
              ) : (
                <div className="agenda-lista">
                  {atendimentos.map((ag) => (
                    <div key={ag.id} className="agenda-item">
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div className="agenda-item-nome">{ag.nome}</div>
                        <div className="agenda-item-sub">{ag.student_id ? "Aluno cadastrado" : "Novo aluno"}{ag.observacao ? ` · ${ag.observacao}` : ""}</div>
                      </div>
                      <span className="agenda-turno-badge">{ag.hora}</span>
                      <button onClick={() => iniciarEdicao(ag)} className="vmd-btn vmd-btn-ghost" style={{ padding: "5px 10px", fontSize: 11 }}>✏️</button>
                      <button onClick={() => handleRemover(ag.id, ag.nome)} className="vmd-btn vmd-btn-danger" style={{ padding: "5px 10px", fontSize: 11 }}>Remover</button>
                    </div>
                  ))}
                </div>
              )}

              <div style={{ borderTop: "1px solid var(--border-color)", margin: "18px 0 14px", paddingTop: 16 }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: "var(--text-main)", marginBottom: 12 }}>
                  {modoEdicao ? "Editar atendimento" : lotado ? "Limite atingido" : "Novo atendimento"}
                </div>
                {lotado && modoEdicao === null ? (
                  <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 0 }}>⚠️ Capacidade máxima de {capacidadeDia} vagas atingida para este dia.</div>
                ) : (
                  <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
                    <div>
                      <label className="vmd-label">Aluno</label>
                      <select value={alunoId} onChange={(e) => setAlunoId(Number(e.target.value))} className="vmd-select" disabled={modoEdicao !== null}>
                        <option value={0}>— Escolha um aluno —</option>
                        {students.filter((s) => s.ativo).map((s) => <option key={s.id} value={s.id}>{s.nome}</option>)}
                      </select>
                      {alunoId === 0 && (
                        <input value={novoNome} onChange={(e) => setNovoNome(e.target.value)} placeholder="Ou digite o nome do aluno" className="vmd-input" style={{ marginTop: 8 }} disabled={modoEdicao !== null} />
                      )}
                    </div>
                    <div>
                      <label className="vmd-label">Horário</label>
                      <div className="agi-grade-horarios">
                        {TURNOS_IND.map((t) => {
                          const horasTurno = HORARIOS_IND.filter((h) => h.turno === t.key);
                          const temHoras = horasTurno.some((h) => horasDisponiveis.some((hd) => hd.hora === h.hora));
                          if (!temHoras && modoEdicao === null) return null;
                          return (
                            <div key={t.key}>
                              <div style={{ fontSize: 10, fontWeight: 600, color: "var(--text-dim)", marginBottom: 4, textTransform: "uppercase", letterSpacing: "0.5px" }}>
                                {t.label} <span style={{ opacity: 0.7, textTransform: "none", letterSpacing: 0 }}>{t.horario}</span>
                              </div>
                              <div className="agi-grade-linha">
                                {horasTurno.map((h) => {
                                  const disp = horasDisponiveis.some((hd) => hd.hora === h.hora);
                                  return (
                                    <button key={h.hora} className={`agi-seg-btn ${hora === h.hora ? "ativo" : ""}`} disabled={!disp} onClick={() => setHora(h.hora)}>
                                      {h.hora}
                                    </button>
                                  );
                                })}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                    <div>
                      <label className="vmd-label">Observação (opcional)</label>
                      <input value={observacao} onChange={(e) => setObservacao(e.target.value)} placeholder="Ex: Foco em glúteos" className="vmd-input" />
                    </div>
                  </div>
                )}
              </div>
            </div>

            <div className="vmd-modal-footer">
              <button onClick={() => { setSelectedDay(null); resetForm(); }} className="vmd-btn vmd-btn-secondary">Fechar</button>
              {(modoEdicao !== null || !lotado) && (
                <button onClick={handleSalvar} className="vmd-btn vmd-btn-primary">
                  {modoEdicao ? "💾 Salvar alterações" : "➕ Agendar atendimento"}
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ======================== UNIFIED PAGE ========================

function CalendariosAgendamento() {
  const [students, setStudents] = useState<Student[]>([]);
  const [activeTab, setActiveTab] = useState("pilates");

  const loadStudents = useCallback(async () => {
    try {
      const list = await ListarStudents() as unknown as Student[];
      setStudents(list || []);
    } catch {
      setStudents([]);
    }
  }, []);

  useEffect(() => {
    loadStudents();
  }, [loadStudents]);

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out", display: "flex", gap: 20, minHeight: "calc(100vh - 120px)" }}>
      <aside style={{
        width: 220,
        flexShrink: 0,
        border: "1px solid var(--border-color)",
        borderRadius: "var(--radius-lg)",
        background: "var(--bg-card)",
        padding: "16px 0",
        alignSelf: "flex-start",
        position: "sticky",
        top: 0,
      }}>
        <div style={{ padding: "0 16px 14px", borderBottom: "1px solid var(--border-color)", marginBottom: 8 }}>
          <h3 style={{ fontSize: 14, fontWeight: 700, margin: 0, letterSpacing: "-0.3px" }}>Calendários</h3>
          <p style={{ color: "var(--text-dim)", fontSize: 11, marginTop: 2 }}>Selecione uma modalidade</p>
        </div>
        <nav style={{ display: "flex", flexDirection: "column", gap: 2, padding: "0 8px" }}>
          {MENU_ITEMS.map((item) => {
            const active = activeTab === item.key;
            return (
              <button
                key={item.key}
                onClick={() => setActiveTab(item.key)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  width: "100%",
                  padding: "10px 12px",
                  border: "none",
                  borderRadius: "var(--radius-sm)",
                  background: active ? "rgba(199, 157, 51, 0.14)" : "transparent",
                  color: active ? "var(--primary)" : "var(--text-muted)",
                  fontFamily: "var(--font-sans)",
                  fontSize: 13,
                  fontWeight: active ? 600 : 400,
                  cursor: "pointer",
                  transition: "all 0.15s",
                  textAlign: "left",
                  borderLeft: active ? `3px solid ${item.color}` : "3px solid transparent",
                }}
                onMouseEnter={(e) => {
                  if (!active) e.currentTarget.style.background = "rgba(255,255,255,0.04)";
                }}
                onMouseLeave={(e) => {
                  if (!active) e.currentTarget.style.background = "transparent";
                }}
              >
                <span style={{ fontSize: 16, width: 22, textAlign: "center", flexShrink: 0 }}>{item.icon}</span>
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>
      </aside>

      <main style={{ flex: 1, minWidth: 0 }}>
        {activeTab === "pilates" && (
          <CalendarioPilates students={students} loadStudents={loadStudents} />
        )}
        {PRATICAS.filter((p) => p.key === activeTab).map((p) => (
          <CalendarioPratica
            key={p.key}
            praticaKey={p.key}
            praticaNome={p.label}
            students={students}
            loadStudents={loadStudents}
          />
        ))}
      </main>
    </div>
  );
}

export default CalendariosAgendamento;

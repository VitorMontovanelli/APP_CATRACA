import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AdicionarAgendamentoIndividual,
  AtualizarAgendamentoIndividual,
  DefinirLimiteDiarioIndividual,
  ListarAgendamentosIndividuaisDia,
  ListarAgendamentosIndividuaisMes,
  ListarStudents,
  ObterConfigAgendaIndividual,
  RemoverAgendamentoIndividual,
} from "../../wailsjs/go/main/App";
import "./AgendaIndividual.css";

const LIMITE_PADRAO = 10;

const HORARIOS = [
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

const TURNOS = [
  { key: "Manhã", label: "Manhã", horario: "06:00 às 11:00" },
  { key: "Tarde", label: "Tarde", horario: "15:00 às 19:00" },
  { key: "Noite", label: "Noite", horario: "19:00 às 20:00" },
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

interface DiaOcupacao {
  data: string;
  total: number;
  capacidade: number;
  praticas: string;
}

function hojeStr(): string {
  const d = new Date();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${mm}-${dd}`;
}

function formatBR(data: string): string {
  const [y, m, d] = data.split("-");
  return `${d}/${m}/${y}`;
}

function praticaLabel(key: string): string {
  return PRATICAS.find((p) => p.key === key)?.label || key;
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
  if (total <= 3) return "agi-day-verde";
  if (total < capacidade) return "agi-day-laranja";
  return "agi-day-vermelho";
}

function ocupacaoLabel(total: number, capacidade: number): { label: string; color: string; bg: string } {
  if (total === 0) return { label: "Livre", color: "#6b7280", bg: "rgba(255,255,255,0.05)" };
  if (total <= 3) return { label: "Livre", color: "#34d399", bg: "rgba(16,185,129,0.15)" };
  if (total < capacidade) return { label: "Média Ocupação", color: "#fbbf24", bg: "rgba(245,158,11,0.15)" };
  return { label: "Lotado", color: "#f87171", bg: "rgba(239,68,68,0.18)" };
}

function parsePraticas(praticasStr: string): string[] {
  if (!praticasStr) return [];
  return [...new Set(praticasStr.split(","))];
}

function AgendaIndividual() {
  const today = hojeStr();
  const now = new Date();
  const [ano, setAno] = useState(now.getFullYear());
  const [mes, setMes] = useState(now.getMonth());
  const [ocupacao, setOcupacao] = useState<Record<string, DiaOcupacao>>({});
  const [loadingMes, setLoadingMes] = useState(false);

  const [selectedDay, setSelectedDay] = useState<string | null>(null);
  const [atendimentos, setAtendimentos] = useState<AgendamentoIndividual[]>([]);
  const [students, setStudents] = useState<Student[]>([]);
  const [loadingDia, setLoadingDia] = useState(false);

  const [limite, setLimite] = useState(LIMITE_PADRAO);
  const [editLimite, setEditLimite] = useState(false);
  const [novoLimite, setNovoLimite] = useState(String(LIMITE_PADRAO));
  const [savingLimite, setSavingLimite] = useState(false);

  const [modoEdicao, setModoEdicao] = useState<number | null>(null);
  const [alunoId, setAlunoId] = useState(0);
  const [novoNome, setNovoNome] = useState("");
  const [hora, setHora] = useState("06:00");
  const [pratica, setPratica] = useState("ventosaterapia");
  const [observacao, setObservacao] = useState("");
  const [error, setError] = useState("");
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; msg: string } | null>(null);

  const [filtroPratica, setFiltroPratica] = useState<string | null>(null);

  const loadMes = useCallback(async (a: number, m: number) => {
    setLoadingMes(true);
    try {
      const list = await ListarAgendamentosIndividuaisMes(a, m + 1) as unknown as DiaOcupacao[];
      const map: Record<string, DiaOcupacao> = {};
      for (const d of list) {
        map[d.data] = { data: d.data, total: d.total, capacidade: d.capacidade, praticas: d.praticas || "" };
      }
      setOcupacao(map);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoadingMes(false);
    }
  }, []);

  const loadConfig = useCallback(async () => {
    try {
      const cfg = await ObterConfigAgendaIndividual() as unknown as { limite_diario: number; limite_maximo: number };
      setLimite(cfg.limite_diario);
      setNovoLimite(String(cfg.limite_diario));
    } catch {
      setLimite(LIMITE_PADRAO);
    }
  }, []);

  useEffect(() => {
    loadMes(ano, mes);
  }, [ano, mes, loadMes]);

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const loadDia = useCallback(async (data: string) => {
    setLoadingDia(true);
    try {
      const list = await ListarAgendamentosIndividuaisDia(data) as unknown as AgendamentoIndividual[];
      setAtendimentos(list || []);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoadingDia(false);
    }
  }, []);

  const loadStudents = useCallback(async () => {
    try {
      const list = await ListarStudents() as unknown as Student[];
      setStudents(list || []);
    } catch {
      setStudents([]);
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
    resetForm();
    loadDia(data);
    if (students.length === 0) loadStudents();
  }

  function resetForm() {
    setModoEdicao(null);
    setAlunoId(0);
    setNovoNome("");
    setHora("06:00");
    setPratica("ventosaterapia");
    setObservacao("");
  }

  async function handleSalvarLimite() {
    const n = Number(novoLimite);
    if (Number.isNaN(n) || n < 1 || n > LIMITE_PADRAO) {
      setError(`O limite deve estar entre 1 e ${LIMITE_PADRAO}`);
      return;
    }
    setSavingLimite(true);
    setError("");
    try {
      await DefinirLimiteDiarioIndividual(n);
      setLimite(n);
      setEditLimite(false);
      setFeedback({ type: "success", msg: `Limite diário ajustado para ${n} atendimentos.` });
      await loadMes(ano, mes);
    } catch (e) {
      setError(String(e));
    } finally {
      setSavingLimite(false);
    }
  }

  const horasDisponiveis = useMemo(() => {
    const ocupadas = new Set(atendimentos.map((a) => a.hora));
    return HORARIOS.filter((h) => !ocupadas.has(h.hora) || (modoEdicao !== null && h.hora === atendimentos.find((a) => a.id === modoEdicao)?.hora));
  }, [atendimentos, modoEdicao]);

  const totalDia = selectedDay ? (ocupacao[selectedDay]?.total || 0) : 0;
  const lotado = totalDia >= limite;

  function iniciarEdicao(ag: AgendamentoIndividual) {
    setModoEdicao(ag.id);
    setAlunoId(ag.student_id || 0);
    setNovoNome(ag.student_id ? "" : ag.nome);
    setHora(ag.hora);
    setPratica(ag.pratica);
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
        await AtualizarAgendamentoIndividual(modoEdicao, hora, pratica, observacao);
        setFeedback({ type: "success", msg: "Atendimento atualizado!" });
        resetForm();
        await Promise.all([loadDia(selectedDay), loadMes(ano, mes)]);
      } catch (e) {
        setError(String(e));
      }
      return;
    }

    if (lotado) {
      setError(`Limite diário de ${limite} atendimentos atingido`);
      return;
    }
    if (alunoId === 0 && !novoNome.trim()) {
      setError("Selecione um aluno ou informe o nome");
      return;
    }

    try {
      await AdicionarAgendamentoIndividual(selectedDay, alunoId, novoNome.trim(), hora, pratica, observacao);
      setFeedback({ type: "success", msg: "Atendimento agendado com sucesso!" });
      resetForm();
      await Promise.all([loadDia(selectedDay), loadMes(ano, mes)]);
    } catch (e) {
      setError(String(e));
    }
  }

  async function handleRemover(id: number, nome: string) {
    if (!window.confirm(`Remover o atendimento de ${nome}?`)) return;
    try {
      await RemoverAgendamentoIndividual(id);
      setFeedback({ type: "success", msg: "Atendimento removido." });
      if (modoEdicao === id) resetForm();
      if (selectedDay) {
        await Promise.all([loadDia(selectedDay), loadMes(ano, mes)]);
      }
    } catch (e) {
      setError(String(e));
    }
  }

  const dias = useMemo(() => {
    const firstDay = new Date(ano, mes, 1).getDay();
    const daysInMonth = new Date(ano, mes + 1, 0).getDate();
    const cells: (string | null)[] = [];
    for (let i = 0; i < firstDay; i++) cells.push(null);
    for (let d = 1; d <= daysInMonth; d++) {
      cells.push(`${ano}-${String(mes + 1).padStart(2, "0")}-${String(d).padStart(2, "0")}`);
    }
    return cells;
  }, [ano, mes]);

  const atendimentosFiltrados = useMemo(() => {
    if (!filtroPratica) return atendimentos;
    return atendimentos.filter((a) => a.pratica === filtroPratica);
  }, [atendimentos, filtroPratica]);

  const atendimentosPorTurno = useMemo(() => {
    const grupos: Record<string, AgendamentoIndividual[]> = { "Manhã": [], "Tarde": [], "Noite": [] };
    for (const ag of atendimentosFiltrados) {
      const h = HORARIOS.find((x) => x.hora === ag.hora);
      const turno = h?.turno || "Manhã";
      grupos[turno].push(ag);
    }
    return grupos;
  }, [atendimentosFiltrados]);

  const vagas = ocupacaoLabel(totalDia, limite);
  const dataSelecionada = selectedDay ? formatBR(selectedDay) : "";

  const praticasUnicasNoDia = selectedDay ? parsePraticas(ocupacao[selectedDay]?.praticas || "") : [];

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 24 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Agenda Individual</h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Atendimentos personalizados da professora. Limite de {limite} atendimentos por dia.
        </p>
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ maxWidth: 640 }}>
          <span>⚠️ {error}</span>
        </div>
      )}
      {feedback && (
        <div className={`vmd-alert ${feedback.type === "success" ? "vmd-alert-success" : "vmd-alert-danger"}`} style={{ maxWidth: 640 }}>
          <span>✨ {feedback.msg}</span>
        </div>
      )}

      <div className="vmd-card" style={{ maxWidth: 900 }}>
        <div className="agi-header">
          <div className="agi-nav">
            <button className="agi-nav-btn" onClick={() => navegar(-1)} aria-label="Mês anterior">‹</button>
            <div className="agi-nav-title">{MESES[mes]} {ano}</div>
            <button className="agi-nav-btn" onClick={() => navegar(1)} aria-label="Próximo mês">›</button>
            <button
              className="vmd-btn vmd-btn-ghost"
              onClick={() => { setAno(now.getFullYear()); setMes(now.getMonth()); }}
            >
              Hoje
            </button>
          </div>
          {loadingMes && <span style={{ fontSize: 12, color: "var(--text-dim)" }}>Carregando...</span>}
        </div>

        <div className="agi-legend">
          <span style={{ fontSize: 12, fontWeight: 600, color: "var(--text-main)" }}>Ocupação:</span>
          <span className="agi-legend-item"><span className="agi-legend-dot" style={{ background: "rgba(16,185,129,0.35)", border: "1px solid rgba(16,185,129,0.6)" }} />Livre (0–3)</span>
          <span className="agi-legend-item"><span className="agi-legend-dot" style={{ background: "rgba(245,158,11,0.35)", border: "1px solid rgba(245,158,11,0.6)" }} />Média (4–{limite - 1})</span>
          <span className="agi-legend-item"><span className="agi-legend-dot" style={{ background: "rgba(239,68,68,0.4)", border: "1px solid rgba(239,68,68,0.6)" }} />Lotado ({limite})</span>
        </div>

        <div className="agi-filtros">
          <span className="agi-filtro-label">Filtrar por prática:</span>
          <button
            className={`agi-filtro-pill ${filtroPratica === null ? "ativo" : ""}`}
            style={filtroPratica === null ? { borderColor: "var(--primary)", color: "var(--text-main)" } : {}}
            onClick={() => setFiltroPratica(null)}
          >
            Todas
          </button>
          {PRATICAS.map((p) => (
            <button
              key={p.key}
              className={`agi-filtro-pill ${filtroPratica === p.key ? "ativo" : ""}`}
              style={filtroPratica === p.key ? { borderColor: praticaColor(p.key), color: praticaColor(p.key), background: `${praticaColor(p.key)}18` } : {}}
              onClick={() => setFiltroPratica(filtroPratica === p.key ? null : p.key)}
            >
              <span className="agi-pratica-dot" style={{ background: praticaColor(p.key) }} />
              {p.label}
            </button>
          ))}
        </div>

        <div className="agi-weekdays">
          {DIAS_SEMANA.map((d) => (
            <div key={d} className="agi-weekday">{d}</div>
          ))}
        </div>

        <div className="agi-grid">
          {dias.map((data, idx) => {
            if (data === null) return <div key={`v-${idx}`} className="agi-vazio" />;
            const dia = ocupacao[data] || { total: 0, capacidade: limite, praticas: "" };
            const dayPraticas = parsePraticas(dia.praticas);
            return (
              <button
                key={data}
                className={`agi-day ${ocupacaoClass(dia.total, dia.capacidade)} ${data === today ? "hoje" : ""} ${data === selectedDay ? "selecionado" : ""}`}
                onClick={() => abrirDia(data)}
              >
                <span className="agi-day-num">{Number(data.split("-")[2])}</span>
                {dayPraticas.length > 0 && (
                  <div className="agi-day-dots">
                    {dayPraticas.map((p) => (
                      <span key={p} style={{ background: praticaColor(p) }} />
                    ))}
                  </div>
                )}
                <span className="agi-day-count">{dia.total}/{dia.capacidade}</span>
              </button>
            );
          })}
        </div>
      </div>

      {selectedDay && (
        <div className="vmd-modal-overlay" onClick={() => setSelectedDay(null)}>
          <div className="vmd-modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 600 }}>
            <div className="vmd-modal-header">
              <div>
                <h3 style={{ fontSize: 16, margin: 0 }}>{dataSelecionada}</h3>
                <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
                  {DIAS_SEMANA[new Date(ano, mes, Number(selectedDay.split("-")[2])).getDay()]}
                </span>
              </div>
              <button onClick={() => setSelectedDay(null)} className="vmd-btn vmd-btn-ghost" style={{ padding: 4, minWidth: "auto" }}>✕</button>
            </div>

            <div className="vmd-modal-body">
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16 }}>
                <span className="agi-vagas-pill" style={{ color: vagas.color, background: vagas.bg }}>
                  {totalDia}/{limite} vagas — {vagas.label}
                </span>
                <span style={{ fontSize: 11, color: "var(--text-dim)" }}>
                  {loadingDia ? "Carregando..." : `${atendimentos.length} atendimento(s)`}
                </span>
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 16, padding: "10px 12px", border: "1px solid var(--border-color)", borderRadius: "var(--radius-sm)", background: "rgba(255,255,255,0.02)" }}>
                <span style={{ fontSize: 12, fontWeight: 600, color: "var(--text-main)" }}>Limite diário:</span>
                {editLimite ? (
                  <>
                    <input
                      type="number"
                      min={1}
                      max={LIMITE_PADRAO}
                      value={novoLimite}
                      onChange={(e) => setNovoLimite(e.target.value)}
                      className="vmd-input"
                      style={{ width: 90, padding: "6px 10px", fontSize: 13 }}
                    />
                    <button
                      onClick={handleSalvarLimite}
                      disabled={savingLimite}
                      className="vmd-btn vmd-btn-primary"
                      style={{ padding: "6px 12px", fontSize: 12 }}
                    >
                      {savingLimite ? "Salvando..." : "Salvar"}
                    </button>
                    <button
                      onClick={() => { setEditLimite(false); setNovoLimite(String(limite)); }}
                      className="vmd-btn vmd-btn-ghost"
                      style={{ padding: "6px 10px", fontSize: 12 }}
                    >
                      Cancelar
                    </button>
                  </>
                ) : (
                  <>
                    <span style={{ fontSize: 13, fontWeight: 600, color: vagas.color }}>
                      {limite} atendimentos/dia
                    </span>
                    <button
                      onClick={() => { setEditLimite(true); setNovoLimite(String(limite)); }}
                      className="vmd-btn vmd-btn-ghost"
                      style={{ padding: "6px 10px", fontSize: 12 }}
                    >
                      ✏️ Editar
                    </button>
                  </>
                )}
              </div>

              {praticasUnicasNoDia.length > 0 && (
                <div style={{ display: "flex", gap: 6, marginBottom: 14, flexWrap: "wrap" }}>
                  {praticasUnicasNoDia.map((p) => (
                    <span key={p} className={`agi-pratica-badge pill-pratica-${p}`}>
                      <span className="agi-pratica-dot" style={{ background: praticaColor(p) }} />
                      {praticaLabel(p)}
                    </span>
                  ))}
                </div>
              )}

              {atendimentos.length === 0 ? (
                <div style={{ textAlign: "center", padding: "24px 0", color: "var(--text-dim)", fontSize: 13 }}>
                  Nenhum atendimento agendado neste dia.
                </div>
              ) : (
                TURNOS.map((turno) => {
                  const itens = atendimentosPorTurno[turno.key] || [];
                  if (itens.length === 0) return null;
                  return (
                    <div key={turno.key} className="agi-turno-bloco">
                      <div className="agi-turno-titulo">
                        {turno.label}
                        <span className="agi-turno-horario">{turno.horario}</span>
                      </div>
                      <div className="agi-lista">
                        {itens.map((ag) => (
                          <div key={ag.id} className="agi-item" style={modoEdicao === ag.id ? { borderColor: "var(--primary)", background: "rgba(199,157,51,0.06)" } : {}}>
                            <span className="agi-item-hora">{ag.hora}</span>
                            <span className={`agi-pratica-badge pill-pratica-${ag.pratica}`}>
                              <span className="agi-pratica-dot" style={{ background: praticaColor(ag.pratica) }} />
                              {praticaLabel(ag.pratica)}
                            </span>
                            <div className="agi-item-info">
                              <div className="agi-item-nome">{ag.nome}</div>
                              {ag.observacao && <div className="agi-item-sub">{ag.observacao}</div>}
                            </div>
                            <div className="agi-item-acoes">
                              <button
                                onClick={() => iniciarEdicao(ag)}
                                className="vmd-btn vmd-btn-ghost"
                                style={{ padding: "4px 8px", fontSize: 11 }}
                              >
                                ✏️
                              </button>
                              <button
                                onClick={() => handleRemover(ag.id, ag.nome)}
                                className="vmd-btn vmd-btn-danger"
                                style={{ padding: "4px 8px", fontSize: 11 }}
                              >
                                ✕
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  );
                })
              )}

              <div style={{ borderTop: "1px solid var(--border-color)", margin: "18px 0 14px", paddingTop: 16 }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: "var(--text-main)", marginBottom: 12 }}>
                  {modoEdicao ? "Editar atendimento" : lotado ? "Limite atingido" : "Novo atendimento"}
                </div>

                {lotado && modoEdicao === null ? (
                  <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 0 }}>
                    <span>⚠️ Limite diário de {limite} atendimentos atingido para este dia.</span>
                  </div>
                ) : (
                  <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
                    <div>
                      <label className="vmd-label">Aluno</label>
                      <select
                        value={alunoId}
                        onChange={(e) => setAlunoId(Number(e.target.value))}
                        className="vmd-select"
                        disabled={modoEdicao !== null}
                      >
                        <option value={0}>— Informe o nome abaixo —</option>
                        {students.filter((s) => s.ativo).map((s) => (
                          <option key={s.id} value={s.id}>{s.nome}</option>
                        ))}
                      </select>
                      {alunoId === 0 && (
                        <input
                          value={novoNome}
                          onChange={(e) => setNovoNome(e.target.value)}
                          placeholder="Ou digite o nome do aluno"
                          className="vmd-input"
                          style={{ marginTop: 8 }}
                          disabled={modoEdicao !== null}
                        />
                      )}
                    </div>

                    <div>
                      <label className="vmd-label">Horário</label>
                      <div className="agi-grade-horarios">
                        {TURNOS.map((t) => {
                          const horasTurno = HORARIOS.filter((h) => h.turno === t.key);
                          const temHoras = horasTurno.some((h) => horasDisponiveis.some((hd) => hd.hora === h.hora));
                          if (!temHoras && modoEdicao === null) return null;
                          return (
                            <div key={t.key}>
                              <div style={{ fontSize: 10, fontWeight: 600, color: "var(--text-dim)", marginBottom: 4, textTransform: "uppercase", letterSpacing: "0.5px" }}>
                                {t.label} <span style={{ opacity: 0.7, textTransform: "none", letterSpacing: 0 }}>{t.horario}</span>
                              </div>
                              <div className="agi-grade-linha">
                                {horasTurno.map((h) => {
                                  const disponivel = horasDisponiveis.some((hd) => hd.hora === h.hora);
                                  return (
                                    <button
                                      key={h.hora}
                                      className={`agi-seg-btn ${hora === h.hora ? "ativo" : ""}`}
                                      disabled={!disponivel}
                                      onClick={() => setHora(h.hora)}
                                    >
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
                      <label className="vmd-label">Prática</label>
                      <div className="agi-seg-toggle" style={{ flexWrap: "wrap" }}>
                        {PRATICAS.map((p) => (
                          <button
                            key={p.key}
                            className={`agi-seg-btn ${pratica === p.key ? "ativo" : ""}`}
                            style={pratica === p.key ? { borderColor: praticaColor(p.key), color: praticaColor(p.key), background: `${praticaColor(p.key)}18` } : {}}
                            onClick={() => setPratica(p.key)}
                          >
                            <span className="agi-pratica-dot" style={{ background: praticaColor(p.key), display: "inline-block" }} />
                            {p.label}
                          </button>
                        ))}
                      </div>
                    </div>

                    <div>
                      <label className="vmd-label">Observação (opcional)</label>
                      <input
                        value={observacao}
                        onChange={(e) => setObservacao(e.target.value)}
                        placeholder="Ex: Foco em glúteos"
                        className="vmd-input"
                      />
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

export default AgendaIndividual;

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AdicionarAgendamento,
  DefinirCapacidadeDia,
  ListarAgendamentosDia,
  ListarAgendamentosMes,
  ListarCapacidadesMes,
  ListarStudents,
  RemoverAgendamento,
} from "../../wailsjs/go/main/App";
import "./Agenda.css";

const CAPACIDADE_PADRAO = 10;

interface DiaOcupacao {
  total: number;
  capacidade: number;
}

const TURNOS = [
  { key: "manha", label: "Manhã", horario: "06:00 às 11:00" },
  { key: "tarde", label: "Tarde", horario: "15:00 às 19:00" },
  { key: "noite", label: "Noite", horario: "19:00 às 20:00" },
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

function turnoInfo(key: string) {
  return TURNOS.find((t) => t.key === key) || { key, label: key, horario: "" };
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

function Agenda() {
  const today = hojeStr();
  const now = new Date();
  const [ano, setAno] = useState(now.getFullYear());
  const [mes, setMes] = useState(now.getMonth());
  const [ocupacao, setOcupacao] = useState<Record<string, DiaOcupacao>>({});
  const [loadingMes, setLoadingMes] = useState(false);

  const [selectedDay, setSelectedDay] = useState<string | null>(null);
  const [agendamentos, setAgendamentos] = useState<Agendamento[]>([]);
  const [students, setStudents] = useState<Student[]>([]);
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
      const map: Record<string, DiaOcupacao> = {};
      for (const d of caps) {
        map[d.data] = { total: 0, capacidade: d.capacidade };
      }
      for (const d of list) {
        map[d.data] = {
          total: d.total,
          capacidade: d.capacidade || CAPACIDADE_PADRAO,
        };
      }
      setOcupacao(map);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoadingMes(false);
    }
  }, []);

  useEffect(() => {
    loadMes(ano, mes);
  }, [ano, mes, loadMes]);

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

  function handleCapacidadeChange(valor: string) {
    const n = Number(valor);
    if (!Number.isNaN(n) && n >= 1) setNovaCapacidade(String(n));
    else setNovaCapacidade(valor);
  }

  async function handleSalvarCapacidade() {
    if (!selectedDay) return;
    const cap = Number(novaCapacidade);
    if (Number.isNaN(cap) || cap < 1) {
      setError("Informe uma capacidade válida (mínimo 1 vaga)");
      return;
    }
    setSavingCapacidade(true);
    setError("");
    setFeedback(null);
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
    if (!selectedDay) return;
    setError("");
    setFeedback(null);
    if (lotado) {
      setError(`Capacidade máxima de ${capacidadeDia} alunos por dia atingida`);
      return;
    }
    if (modoNovo && !novoNome.trim()) {
      setError("Informe o nome do novo aluno");
      return;
    }
    if (!modoNovo && alunoId === 0) {
      setError("Selecione um aluno cadastrado ou alterne para 'Novo aluno'");
      return;
    }
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

  const vagas = ocupacaoLabel(totalDia, capacidadeDia);
  const dataSelecionada = selectedDay ? formatBR(selectedDay) : "";

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 24 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Agenda Calendário</h2>
<p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
            Controle de vagas e agendamento de alunos. Capacidade padrão de {CAPACIDADE_PADRAO} alunos por dia (editável).
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
        <div className="agenda-header">
          <div className="agenda-nav">
            <button className="agenda-nav-btn" onClick={() => navegar(-1)} aria-label="Mês anterior">‹</button>
            <div className="agenda-nav-title">{MESES[mes]} {ano}</div>
            <button className="agenda-nav-btn" onClick={() => navegar(1)} aria-label="Próximo mês">›</button>
            <button
              className="vmd-btn vmd-btn-ghost"
              onClick={() => { setAno(now.getFullYear()); setMes(now.getMonth()); }}
            >
              Hoje
            </button>
          </div>
          {loadingMes && <span style={{ fontSize: 12, color: "var(--text-dim)" }}>Carregando...</span>}
        </div>

        <div className="agenda-legend">
          <span style={{ fontSize: 12, fontWeight: 600, color: "var(--text-main)" }}>Ocupação:</span>
          <span className="legend-item"><span className="legend-dot" style={{ background: "rgba(16,185,129,0.35)", border: "1px solid rgba(16,185,129,0.6)" }} />Livre (0 a 3)</span>
          <span className="legend-item"><span className="legend-dot" style={{ background: "rgba(245,158,11,0.35)", border: "1px solid rgba(245,158,11,0.6)" }} />Média (4 até abaixo da capacidade)</span>
          <span className="legend-item"><span className="legend-dot" style={{ background: "rgba(239,68,68,0.4)", border: "1px solid rgba(239,68,68,0.6)" }} />Lotado (capacidade atingida)</span>
        </div>

        <div className="agenda-weekdays">
          {DIAS_SEMANA.map((d) => (
            <div key={d} className="agenda-weekday">{d}</div>
          ))}
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
                <h3 style={{ fontSize: 16, margin: 0 }}>{dataSelecionada}</h3>
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
                <span style={{ fontSize: 12, fontWeight: 600, color: "var(--text-main)" }}>Ocupação (capacidade de vagas):</span>
                {editCapacidade ? (
                  <>
                    <input
                      type="number"
                      min={1}
                      value={novaCapacidade}
                      onChange={(e) => handleCapacidadeChange(e.target.value)}
                      className="vmd-input"
                      style={{ width: 90, padding: "6px 10px", fontSize: 13 }}
                    />
                    <button
                      onClick={handleSalvarCapacidade}
                      disabled={savingCapacidade}
                      className="vmd-btn vmd-btn-primary"
                      style={{ padding: "6px 12px", fontSize: 12 }}
                    >
                      {savingCapacidade ? "Salvando..." : "Salvar"}
                    </button>
                    <button
                      onClick={() => { setEditCapacidade(false); setNovaCapacidade(String(capacidadeDia)); }}
                      className="vmd-btn vmd-btn-ghost"
                      style={{ padding: "6px 10px", fontSize: 12 }}
                    >
                      Cancelar
                    </button>
                  </>
                ) : (
                  <>
                    <span style={{ fontSize: 13, fontWeight: 600, color: vagas.color }}>
                      {capacidadeDia} vagas/dia
                    </span>
                    <button
                      onClick={() => { setEditCapacidade(true); setNovaCapacidade(String(capacidadeDia)); }}
                      className="vmd-btn vmd-btn-ghost"
                      style={{ padding: "6px 10px", fontSize: 12 }}
                    >
                      ✏️ Editar
                    </button>
                  </>
                )}
              </div>

              {agendamentos.length === 0 ? (
                <div style={{ textAlign: "center", padding: "24px 0", color: "var(--text-dim)", fontSize: 13 }}>
                  Nenhum aluno agendado neste dia.
                </div>
              ) : (
                <div className="agenda-lista">
                  {agendamentos.map((ag) => (
                    <div key={ag.id} className="agenda-item">
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div className="agenda-item-nome">{ag.nome}</div>
                        <div className="agenda-item-sub">
                          {ag.student_id ? "Aluno cadastrado" : "Novo aluno"}
                          {ag.telefone ? ` · ${ag.telefone}` : ""}
                        </div>
                      </div>
                      <span className="agenda-turno-badge">
                        {turnoInfo(ag.turno).label} · {turnoInfo(ag.turno).horario}
                      </span>
                      <button
                        onClick={() => handleRemover(ag.id, ag.nome)}
                        className="vmd-btn vmd-btn-danger"
                        style={{ padding: "5px 10px", fontSize: 11 }}
                      >
                        Remover
                      </button>
                    </div>
                  ))}
                </div>
              )}

              <div style={{ borderTop: "1px solid var(--border-color)", margin: "18px 0 14px", paddingTop: 16 }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: "var(--text-main)", marginBottom: 12 }}>
                  {lotado ? "Dia lotado" : "Adicionar aluno"}
                </div>

                {lotado ? (
                  <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 0 }}>
                    <span>⚠️ Capacidade máxima de {capacidadeDia} alunos atingida para este dia.</span>
                  </div>
                ) : (
                  <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
                    <div className="agenda-seg-toggle">
                      <button
                        className={`agenda-seg-btn ${!modoNovo ? "ativo" : ""}`}
                        onClick={() => setModoNovo(false)}
                      >
                        🎓 Aluno cadastrado
                      </button>
                      <button
                        className={`agenda-seg-btn ${modoNovo ? "ativo" : ""}`}
                        onClick={() => setModoNovo(true)}
                      >
                        🆕 Novo aluno
                      </button>
                    </div>

                    {modoNovo ? (
                      <div>
                        <label className="vmd-label">Nome do aluno</label>
                        <input
                          value={novoNome}
                          onChange={(e) => setNovoNome(e.target.value)}
                          placeholder="Ex: Pedro Silva"
                          className="vmd-input"
                        />
                      </div>
                    ) : (
                      <div>
                        <label className="vmd-label">Selecionar aluno</label>
                        <select
                          value={alunoId}
                          onChange={(e) => setAlunoId(Number(e.target.value))}
                          className="vmd-select"
                        >
                          <option value={0}>— Escolha um aluno —</option>
                          {students
                            .filter((s) => s.ativo)
                            .map((s) => (
                              <option key={s.id} value={s.id}>{s.nome}</option>
                            ))}
                        </select>
                      </div>
                    )}

                    <div>
                      <label className="vmd-label">Turno</label>
                      <div style={{ display: "flex", gap: 6 }}>
                        {TURNOS.map((t) => (
                          <button
                            key={t.key}
                            className={`agenda-seg-btn ${turno === t.key ? "ativo" : ""}`}
                            style={{ flex: 1 }}
                            onClick={() => setTurno(t.key)}
                          >
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
              {!lotado && (
                <button onClick={handleAdicionar} className="vmd-btn vmd-btn-primary">➕ Reservar vaga</button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Agenda;

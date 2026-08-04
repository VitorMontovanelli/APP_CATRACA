import { useState, useEffect } from "react";
import { AppVersion, ObterMetricasHome, ObterEstatisticasAcesso } from "../../wailsjs/go/main/App";

interface PlanoContador {
  nome: string;
  qtd: number;
}

interface MetricasHome {
  total_usuarios: number;
  total_admins: number;
  total_alunos: number;
  total_faturado_cents: number;
  total_atraso_cents: number;
  total_pendente_cents: number;
  planos_distribuicao: PlanoContador[];
}

interface AcessoEstatistica {
  label: string;
  liberado: number;
  negado: number;
}

function formatCents(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function Home() {
  const [metrics, setMetrics] = useState<MetricasHome | null>(null);
  const [version, setVersion] = useState("");
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<"dia" | "semana" | "mes">("dia");
  const [accessData, setAccessData] = useState<AcessoEstatistica[]>([]);
  const [loadingChart, setLoadingChart] = useState(false);
  const [hoveredBar, setHoveredBar] = useState<number | null>(null);

  useEffect(() => {
    async function loadMetrics() {
      try {
        const data = await ObterMetricasHome();
        setMetrics(data as unknown as MetricasHome);
        setVersion(await AppVersion());
      } catch (e) {
        console.error("Erro ao carregar métricas", e);
      } finally {
        setLoading(false);
      }
    }
    loadMetrics();
  }, []);

  useEffect(() => {
    async function loadChart() {
      setLoadingChart(true);
      try {
        const data = await ObterEstatisticasAcesso(filter);
        setAccessData((data as unknown as AcessoEstatistica[]) || []);
      } catch (e) {
        console.error("Erro ao carregar estatísticas do gráfico", e);
      } finally {
        setLoadingChart(false);
      }
    }
    loadChart();
  }, [filter]);

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 48, color: "var(--text-muted)", animation: "fadeIn 0.3s ease-out" }}>
        Carregando painel de controle...
      </div>
    );
  }

  // Financial and Student KPI Cards (Grid 4)
  const kpiCards = [
    {
      label: "Alunos Matriculados",
      value: metrics?.total_alunos || 0,
      color: "var(--cyan)",
      icon: "🎓",
      isCurrency: false,
    },
    {
      label: "Total Pago",
      value: metrics?.total_faturado_cents || 0,
      color: "var(--success)",
      icon: "💵",
      isCurrency: true,
    },
    {
      label: "Valor Pendente",
      value: metrics?.total_pendente_cents || 0,
      color: "var(--warning)",
      icon: "⏳",
      isCurrency: true,
    },
    {
      label: "Valor em Atraso",
      value: metrics?.total_atraso_cents || 0,
      color: "var(--danger)",
      icon: "⚠️",
      isCurrency: true,
    },
  ];

  // Access chart limits
  const maxAccessVal = Math.max(
    ...accessData.map((d) => d.liberado + d.negado),
    10 // fallback min
  );

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      {/* Title */}
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Visão Geral</h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Dashboard executivo e resumo de faturamento da academia.
        </p>
      </div>

      {/* Grid 4 - Main KPIs */}
      <div className="vmd-grid-4" style={{ marginBottom: 28 }}>
        {kpiCards.map((card) => (
          <div
            key={card.label}
            className="vmd-card"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              padding: "20px 24px",
            }}
          >
            <div>
              <div style={{ fontSize: 11, fontWeight: 600, color: "var(--text-dim)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
                {card.label}
              </div>
              <div style={{ fontSize: 22, fontWeight: 700, color: card.color, marginTop: 8 }}>
                {card.isCurrency ? formatCents(card.value) : card.value}
              </div>
            </div>
            <div style={{ fontSize: 28, opacity: 0.8 }}>{card.icon}</div>
          </div>
        ))}
      </div>

      {/* Access Log Chart Section */}
      <div className="vmd-card" style={{ marginBottom: 28, display: "flex", flexDirection: "column", gap: 20 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 12 }}>
          <div>
            <h3 style={{ fontSize: 16, fontWeight: 600, margin: 0 }}>Fluxo de Entradas (Catraca)</h3>
            <p style={{ color: "var(--text-muted)", fontSize: 12, marginTop: 2 }}>
              Histórico de acessos liberados vs negados na academia.
            </p>
          </div>
          {/* Filters Selectors */}
          <div style={{ display: "flex", gap: 4, background: "rgba(255,255,255,0.03)", padding: 4, borderRadius: 8, border: "1px solid var(--border-color)" }}>
            <button
              onClick={() => setFilter("dia")}
              className={`vmd-btn ${filter === "dia" ? "vmd-btn-primary" : "vmd-btn-ghost"}`}
              style={{ padding: "4px 12px", height: 28, fontSize: 12, minWidth: "auto" }}
            >
              Diário
            </button>
            <button
              onClick={() => setFilter("semana")}
              className={`vmd-btn ${filter === "semana" ? "vmd-btn-primary" : "vmd-btn-ghost"}`}
              style={{ padding: "4px 12px", height: 28, fontSize: 12, minWidth: "auto" }}
            >
              Semanal
            </button>
            <button
              onClick={() => setFilter("mes")}
              className={`vmd-btn ${filter === "mes" ? "vmd-btn-primary" : "vmd-btn-ghost"}`}
              style={{ padding: "4px 12px", height: 28, fontSize: 12, minWidth: "auto" }}
            >
              Mensal
            </button>
          </div>
        </div>

        {/* Custom Visual Bar Chart */}
        {loadingChart ? (
          <div style={{ height: 180, display: "flex", alignItems: "center", justifyContent: "center", color: "var(--text-muted)", fontSize: 13 }}>
            Carregando estatísticas...
          </div>
        ) : accessData.length === 0 ? (
          <div style={{ height: 180, display: "flex", alignItems: "center", justifyContent: "center", color: "var(--text-muted)", fontSize: 13 }}>
            Nenhum registro de acesso encontrado para este período.
          </div>
        ) : (
          <div style={{ position: "relative", marginTop: 10 }}>
            {/* Chart Area */}
            <div
              style={{
                height: 180,
                display: "flex",
                alignItems: "flex-end",
                justifyContent: "space-between",
                borderBottom: "1px solid var(--border-color)",
                paddingBottom: 8,
                gap: 8,
              }}
            >
              {accessData.map((d, index) => {
                const total = d.liberado + d.negado;
                const hLiberado = maxAccessVal > 0 ? (d.liberado / maxAccessVal) * 100 : 0;
                const hNegado = maxAccessVal > 0 ? (d.negado / maxAccessVal) * 100 : 0;
                const isHovered = hoveredBar === index;

                return (
                  <div
                    key={d.label + index}
                    style={{
                      flex: 1,
                      display: "flex",
                      flexDirection: "column",
                      alignItems: "center",
                      height: "100%",
                      justifyContent: "flex-end",
                      position: "relative",
                      cursor: "pointer",
                    }}
                    onMouseEnter={() => setHoveredBar(index)}
                    onMouseLeave={() => setHoveredBar(null)}
                  >
                    {/* Tooltip on Hover */}
                    {isHovered && (
                      <div
                        style={{
                          position: "absolute",
                          bottom: "105%",
                          backgroundColor: "#0d0f22",
                          border: "1px solid var(--border-color)",
                          borderRadius: 8,
                          padding: "8px 12px",
                          zIndex: 10,
                          fontSize: 11,
                          minWidth: 100,
                          boxShadow: "0 4px 12px rgba(0,0,0,0.5)",
                          animation: "fadeIn 0.15s ease-out",
                          color: "var(--text-main)",
                        }}
                      >
                        <div style={{ fontWeight: 600, color: "var(--cyan)", marginBottom: 4 }}>{d.label}</div>
                        <div style={{ display: "flex", justifyContent: "space-between", gap: 8 }}>
                          <span style={{ color: "var(--success)" }}>Liberados:</span>
                          <strong>{d.liberado}</strong>
                        </div>
                        <div style={{ display: "flex", justifyContent: "space-between", gap: 8 }}>
                          <span style={{ color: "var(--danger)" }}>Negados:</span>
                          <strong>{d.negado}</strong>
                        </div>
                        <div style={{ borderTop: "1px solid rgba(255,255,255,0.05)", margin: "4px 0" }} />
                        <div style={{ display: "flex", justifyContent: "space-between", gap: 8, fontWeight: 600 }}>
                          <span>Total:</span>
                          <span>{total}</span>
                        </div>
                      </div>
                    )}

                    {/* Stacked Bars */}
                    <div
                      style={{
                        width: "100%",
                        maxWidth: 32,
                        minWidth: 12,
                        display: "flex",
                        flexDirection: "column",
                        justifyContent: "flex-end",
                        height: "100%",
                        borderRadius: "4px 4px 0 0",
                        overflow: "hidden",
                        backgroundColor: "rgba(255,255,255,0.01)",
                        transition: "all 0.2s",
                        transform: isHovered ? "scaleX(1.05)" : "none",
                      }}
                    >
                      {/* Negados (Red segment) */}
                      <div
                        style={{
                          height: `${hNegado}%`,
                          backgroundColor: "var(--danger)",
                          opacity: isHovered ? 1 : 0.8,
                          transition: "height 0.3s ease-out",
                        }}
                      />
                      {/* Liberados (Green segment) */}
                      <div
                        style={{
                          height: `${hLiberado}%`,
                          backgroundColor: "var(--success)",
                          opacity: isHovered ? 1 : 0.8,
                          transition: "height 0.3s ease-out",
                        }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Labels under bottom border */}
            <div style={{ display: "flex", justifyContent: "space-between", marginTop: 8, padding: "0 4px" }}>
              {accessData.map((d, index) => (
                <div
                  key={d.label + index}
                  style={{
                    flex: 1,
                    textAlign: "center",
                    fontSize: 10,
                    color: hoveredBar === index ? "var(--text-main)" : "var(--text-muted)",
                    fontWeight: hoveredBar === index ? 600 : 400,
                    transition: "color 0.2s",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  }}
                >
                  {d.label}
                </div>
              ))}
            </div>

            {/* Legend */}
            <div style={{ display: "flex", justifyContent: "center", gap: 20, marginTop: 20, fontSize: 12 }}>
              <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                <span style={{ width: 10, height: 10, borderRadius: 2, backgroundColor: "var(--success)" }} />
                <span style={{ color: "var(--text-muted)" }}>Acessos Liberados</span>
              </div>
              <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                <span style={{ width: 10, height: 10, borderRadius: 2, backgroundColor: "var(--danger)" }} />
                <span style={{ color: "var(--text-muted)" }}>Acessos Negados</span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Grid 2 - Bottom Panels */}
      <div className="vmd-grid-2">
        {/* Plan Distribution Card */}
        <div className="vmd-card" style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          <div>
            <h3 style={{ fontSize: 16, fontWeight: 600, margin: 0 }}>Planos Ativos por Modalidade</h3>
            <p style={{ color: "var(--text-muted)", fontSize: 12, marginTop: 2 }}>
              Adesão dos alunos por categoria de plano.
            </p>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
            {metrics?.planos_distribuicao && metrics.planos_distribuicao.length > 0 ? (
              metrics.planos_distribuicao.map((p) => {
                const totalAlunos = metrics.total_alunos || 1;
                const percent = Math.min((p.qtd / totalAlunos) * 100, 100);

                return (
                  <div key={p.nome} style={{ display: "flex", flexDirection: "column", gap: 6 }}>
                    <div style={{ display: "flex", justifyContent: "space-between", fontSize: 13 }}>
                      <span style={{ fontWeight: 500, color: "var(--text-main)" }}>{p.nome}</span>
                      <span style={{ color: "var(--text-muted)" }}>{p.qtd} aluno(s)</span>
                    </div>
                    <div
                      style={{
                        height: 6,
                        width: "100%",
                        backgroundColor: "rgba(255,255,255,0.03)",
                        borderRadius: 3,
                        overflow: "hidden",
                        border: "1px solid rgba(255,255,255,0.02)",
                      }}
                    >
                      <div
                        style={{
                          height: "100%",
                          width: `${percent}%`,
                          backgroundColor: "var(--primary)",
                          borderRadius: 3,
                        }}
                      />
                    </div>
                  </div>
                );
              })
            ) : (
              <div style={{ color: "var(--text-muted)", fontSize: 13, textAlign: "center", padding: 24 }}>
                Nenhum plano com matrículas ativas.
              </div>
            )}
          </div>
        </div>

        {/* System Stats Card */}
        <div className="vmd-card" style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          <div>
            <h3 style={{ fontSize: 16, fontWeight: 600, margin: 0 }}>Estatísticas Administrativas</h3>
            <p style={{ color: "var(--text-muted)", fontSize: 12, marginTop: 2 }}>
              Controle de usuários do sistema.
            </p>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: 12, marginTop: 8 }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <span style={{ fontSize: 13, color: "var(--text-muted)" }}>Total de Usuários Administrativos</span>
              <span style={{ fontSize: 14, fontWeight: 700, color: "var(--cyan)" }}>{metrics?.total_usuarios || 0}</span>
            </div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <span style={{ fontSize: 13, color: "var(--text-muted)" }}>Administradores Ativos</span>
              <span style={{ fontSize: 14, fontWeight: 700, color: "var(--primary)" }}>{metrics?.total_admins || 0}</span>
            </div>
            <div style={{ borderTop: "1px solid var(--border-color)", margin: "8px 0" }} />
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <span style={{ fontSize: 13, color: "var(--text-muted)" }}>Status do Banco de Dados</span>
              <span className="vmd-badge vmd-badge-success" style={{ fontSize: 11 }}>Operacional</span>
            </div>
          </div>
        </div>
      </div>

      {version && (
        <div style={{ marginTop: 48, fontSize: 11, color: "var(--text-dim)", textAlign: "center" }}>
          Versão do Sistema: v{version}
        </div>
      )}
    </div>
  );
}

export default Home;

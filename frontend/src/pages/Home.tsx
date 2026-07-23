import { useState, useEffect } from "react";
import { AppVersion, ListarUsuarios, ListarAccessLogs } from "../../wailsjs/go/main/App";

function Home() {
  const [stats, setStats] = useState({ total: 0, admins: 0, alunos: 0, logs: 0 });
  const [version, setVersion] = useState("");

  useEffect(() => {
    async function load() {
      const users = await ListarUsuarios();
      const logs = await ListarAccessLogs();
      setStats({
        total: users.length,
        admins: users.filter((u) => u.cargo === "super_admin" || u.cargo === "admin").length,
        alunos: users.filter((u) => u.cargo === "aluno").length,
        logs: logs.length,
      });
      try { setVersion(await AppVersion()); } catch {}
    }
    load();
  }, []);

  const cards = [
    { label: "Total de Usuários", value: stats.total, color: "#5eead4" },
    { label: "Administradores", value: stats.admins, color: "#60a5fa" },
    { label: "Alunos", value: stats.alunos, color: "#fbbf24" },
    { label: "Logs de Acesso", value: stats.logs, color: "#f87171" },
  ];

  return (
    <div>
      <h2 style={{ margin: "0 0 24px 0", fontSize: 20, fontWeight: 500 }}>Visão Geral</h2>
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 16 }}>
        {cards.map((card) => (
          <div
            key={card.label}
            style={{
              background: "#121a29",
              border: "1px solid #232e42",
              borderRadius: 10,
              padding: "20px",
            }}
          >
            <div style={{ fontSize: 13, color: "#7c8798", marginBottom: 8 }}>{card.label}</div>
            <div style={{ fontSize: 28, fontWeight: 600, color: card.color }}>{card.value}</div>
          </div>
        ))}
      </div>
      {version && (
        <div style={{ marginTop: 32, fontSize: 12, color: "#3a4255", textAlign: "center" }}>
          v{version}
        </div>
      )}
    </div>
  );
}

export default Home;

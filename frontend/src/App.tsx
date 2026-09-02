import { Component, useState, type ReactNode } from "react";
import Login from "./Login";
import Dashboard from "./Dashboard";
import Home from "./pages/Home";

import Usuarios from "./pages/Usuarios";
import Alunos from "./pages/Alunos";
import Logs from "./pages/Logs";
import Planos from "./pages/Planos";
import MetodosPagamento from "./pages/MetodosPagamento";
import Financeiro from "./pages/Financeiro";
import Cobranca from "./pages/Cobranca";
import Acesso from "./pages/Acesso";
import Backup from "./pages/Backup";
import CalendariosAgendamento from "./pages/CalendariosAgendamento";

interface User {
  id: number;
  name: string;
  email: string;
  cargo: string;
  ativo: boolean;
  foto?: string;
  permissoes: string;
  criado_em: string;
}

class PageErrorBoundary extends Component<{ children: ReactNode }, { hasError: boolean; message: string }> {
  state = { hasError: false, message: "" };

  static getDerivedStateFromError(error: unknown) {
    return { hasError: true, message: error instanceof Error ? error.message : String(error) };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            padding: 48,
            gap: 12,
            textAlign: "center",
          }}
        >
          <div style={{ fontSize: 40 }}>⚠️</div>
          <div style={{ fontSize: 18, fontWeight: 600, color: "var(--danger)" }}>
            Ocorreu um erro ao exibir esta página.
          </div>
          <div style={{ fontSize: 13, color: "var(--text-muted)", maxWidth: 480, whiteSpace: "pre-wrap", fontFamily: "monospace" }}>
            {this.state.message}
          </div>
          <button
            onClick={() => this.setState({ hasError: false, message: "" })}
            className="vmd-btn vmd-btn-secondary"
            style={{ marginTop: 8 }}
          >
            Tentar novamente
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

function App() {
  const [user, setUser] = useState<User | null>(null);

  if (!user) {
    return <Login onSuccess={(u) => setUser(u as unknown as User)} />;
  }

  return (
    <Dashboard
      user={user}
      onLogout={() => setUser(null)}
      onFotoUpdate={(foto) => setUser((prev) => (prev ? { ...prev, foto } : prev))}
    >
      {(page) => (
        <PageErrorBoundary key={page}>
          {(() => {
            switch (page) {
              case "home":
                return <Home />;
              case "usuarios":
                return <Usuarios user={user} />;
              case "alunos":
                return <Alunos user={user} />;
              case "agenda":
                return <CalendariosAgendamento />;
              case "planos":
                return <Planos user={user} />;
              case "metodos_pagamento":
                return <MetodosPagamento user={user} />;
              case "financeiro":
                return <Financeiro user={user} />;
              case "cobranca":
                return <Cobranca />;
              case "logs":
                return <Logs />;
              case "acesso":
                return <Acesso />;
              case "backup":
                return <Backup />;
            }
          })()}
        </PageErrorBoundary>
      )}
    </Dashboard>
  );
}

export default App;

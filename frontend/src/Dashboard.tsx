import { useState } from "react";
import PerfilModal from "./pages/Perfil";
import "./Dashboard.css";

interface User {
  id: number;
  name: string;
  email: string;
  cargo: string;
}

type Page = "home" | "usuarios" | "alunos" | "logs" | "planos" | "metodos_pagamento" | "financeiro" | "catraca" | "cobranca";

interface DashboardProps {
  user: User;
  onLogout: () => void;
  children: (page: Page) => React.ReactNode;
}

const fullNav: { page: Page; label: string; icon: string; minCargo: string }[] = [
  { page: "home", label: "Visão Geral", icon: "📊", minCargo: "admin" },
  { page: "catraca", label: "Catraca", icon: "🔑", minCargo: "admin" },
  { page: "alunos", label: "Alunos", icon: "🎓", minCargo: "admin" },
  { page: "planos", label: "Planos", icon: "📋", minCargo: "admin" },
  { page: "financeiro", label: "Financeiro", icon: "💰", minCargo: "admin" },
  { page: "cobranca", label: "Cobrança", icon: "📋", minCargo: "admin" },
  { page: "usuarios", label: "Usuários", icon: "👥", minCargo: "admin" },
  { page: "metodos_pagamento", label: "Pagamentos", icon: "💳", minCargo: "admin" },
  { page: "logs", label: "Logs de Acesso", icon: "📋", minCargo: "admin" },
];

const cargoLevel: Record<string, number> = {
  super_admin: 3,
  admin: 2,
};

function Dashboard({ user, onLogout, children }: DashboardProps) {
  const userLevel = cargoLevel[user.cargo] || 0;
  const navItems = fullNav.filter((item) => (cargoLevel[item.minCargo] || 0) <= userLevel);
  const [currentPage, setCurrentPage] = useState<Page>("home");
  const [showPerfil, setShowPerfil] = useState(false);

  return (
    <div className="dashboard">
      <PerfilModal user={user} show={showPerfil} onClose={() => setShowPerfil(false)} />

      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-logo">AcessoID</div>
          <button
            onClick={() => setShowPerfil(true)}
            style={{
              background: "none", border: "none", cursor: "pointer",
              textAlign: "left", padding: 0, width: "100%", color: "inherit",
              fontFamily: "inherit",
            }}
          >
            <div className="sidebar-user">
              <span className="sidebar-user-name">{user.name}</span>
              <span className="sidebar-user-cargo">{user.cargo}</span>
            </div>
          </button>
        </div>

        <nav className="sidebar-nav">
          {navItems.map((item) => (
            <button
              key={item.page}
              className={`sidebar-nav-item ${currentPage === item.page ? "active" : ""}`}
              onClick={() => setCurrentPage(item.page)}
            >
              <span className="sidebar-nav-icon">{item.icon}</span>
              {item.label}
            </button>
          ))}
        </nav>

        <div className="sidebar-footer">
          <button className="sidebar-logout" onClick={onLogout}>
            Sair
          </button>
        </div>
      </aside>

      <main className="main-content">
        {children(currentPage)}
      </main>
    </div>
  );
}

export default Dashboard;
export type { Page, User };

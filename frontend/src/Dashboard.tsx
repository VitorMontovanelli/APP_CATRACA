import { useState } from "react";
import PerfilModal from "./pages/Perfil";
import NotificationBell from "./components/NotificationBell";
import "./Dashboard.css";

interface User {
  id: number;
  name: string;
  email: string;
  cargo: string;
  foto?: string;
  permissoes?: string;
}

type Page = "home" | "usuarios" | "alunos" | "logs" | "planos" | "metodos_pagamento" | "financeiro" | "catraca" | "cobranca" | "acesso" | "backup";

interface DashboardProps {
  user: User;
  onLogout: () => void;
  onFotoUpdate?: (foto: string) => void;
  children: (page: Page) => React.ReactNode;
}

interface NavItem {
  page: Page;
  label: string;
  icon: string;
  minCargo: string;
}

interface NavGroup {
  title: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    title: "Principal",
    items: [
      { page: "home", label: "Visão Geral", icon: "📊", minCargo: "admin" },
      { page: "catraca", label: "Catraca Virtual", icon: "🔑", minCargo: "admin" },
    ],
  },
  {
    title: "Matrículas",
    items: [
      { page: "alunos", label: "Gestão de Alunos", icon: "🎓", minCargo: "admin" },
      { page: "planos", label: "Planos de Acesso", icon: "📋", minCargo: "admin" },
    ],
  },
  {
    title: "Financeiro",
    items: [
      { page: "financeiro", label: "Visão Financeira", icon: "💰", minCargo: "admin" },
      { page: "cobranca", label: "Cobranças", icon: "⚠️", minCargo: "admin" },
      { page: "metodos_pagamento", label: "Pagamentos", icon: "💳", minCargo: "admin" },
    ],
  },
  {
    title: "Sistema",
    items: [
      { page: "usuarios", label: "Usuários", icon: "👥", minCargo: "admin" },
      { page: "logs", label: "Logs de Acesso", icon: "📋", minCargo: "admin" },
      { page: "acesso", label: "Acessos", icon: "🔑", minCargo: "super_admin" },
      { page: "backup", label: "Backup Telegram", icon: "☁️", minCargo: "super_admin" },
    ],
  },
];

const cargoLevel: Record<string, number> = {
  super_admin: 3,
  admin: 2,
};

function Dashboard({ user, onLogout, onFotoUpdate, children }: DashboardProps) {
  const userLevel = cargoLevel[user.cargo] || 0;
  const [currentPage, setCurrentPage] = useState<Page>("home");
  const [showPerfil, setShowPerfil] = useState(false);

  // Filter items in each group by cargo and dynamic permissions
  const filteredGroups = navGroups
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => {
        // First check role requirements
        if ((cargoLevel[item.minCargo] || 0) > userLevel) return false;

        // super_admin has absolute access
        if (user.cargo === "super_admin") return true;

        // admin is subject to dynamic permissions
        if (item.page === "acesso") return false; // admin never sees accesses page

        // check list
        const perms = user.permissoes ? user.permissoes.split(",") : ["home", "alunos", "planos", "financeiro", "cobranca", "metodos_pagamento", "usuarios", "logs"];
        return perms.includes(item.page);
      }),
    }))
    .filter((group) => group.items.length > 0);

  return (
    <div className="dashboard">
      <PerfilModal user={user} show={showPerfil} onClose={() => setShowPerfil(false)} onFotoUpdate={onFotoUpdate} />
      <NotificationBell />

      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-logo">CatracaVMD</div>
          
          <button
            onClick={() => setShowPerfil(true)}
            className="sidebar-user-btn"
          >
            <div className="sidebar-user-avatar">
              {user.foto ? (
                <img src={user.foto} alt={user.name} style={{ width: "100%", height: "100%", objectFit: "cover" }} />
              ) : (
                user.name.charAt(0).toUpperCase()
              )}
            </div>
            <div className="sidebar-user-info">
              <span className="sidebar-user-name">{user.name}</span>
              <span className="sidebar-user-cargo">{user.cargo.replace("_", " ")}</span>
            </div>
          </button>
        </div>

        <nav className="sidebar-nav">
          {filteredGroups.map((group) => (
            <div key={group.title} className="sidebar-group">
              <div className="sidebar-group-title">{group.title}</div>
              <div className="sidebar-group-items">
                {group.items.map((item) => (
                  <button
                    key={item.page}
                    className={`sidebar-nav-item ${currentPage === item.page ? "active" : ""}`}
                    onClick={() => setCurrentPage(item.page)}
                  >
                    <span className="sidebar-nav-icon">{item.icon}</span>
                    <span className="sidebar-nav-label">{item.label}</span>
                  </button>
                ))}
              </div>
            </div>
          ))}
        </nav>

        <div className="sidebar-footer">
          <button className="sidebar-logout" onClick={onLogout}>
            <span>🚪</span> Sair da Conta
          </button>
        </div>
      </aside>

      <main className="main-content">
        <div className="content-container">
          {children(currentPage)}
        </div>
      </main>
    </div>
  );
}

export default Dashboard;
export type { Page, User };

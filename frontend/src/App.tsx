import { useState } from "react";
import Login from "./Login";
import Dashboard from "./Dashboard";
import Home from "./pages/Home";

import Usuarios from "./pages/Usuarios";
import Alunos from "./pages/Alunos";
import Catraca from "./pages/Catraca";
import Logs from "./pages/Logs";
import Planos from "./pages/Planos";
import MetodosPagamento from "./pages/MetodosPagamento";
import Financeiro from "./pages/Financeiro";
import Cobranca from "./pages/Cobranca";

interface User {
  id: number;
  name: string;
  email: string;
  cargo: string;
  ativo: boolean;
  criado_em: string;
}

function App() {
  const [user, setUser] = useState<User | null>(null);

  if (!user) {
    return <Login onSuccess={(u) => setUser(u as unknown as User)} />;
  }

  return (
    <Dashboard user={user} onLogout={() => setUser(null)}>
      {(page) => {
        switch (page) {
          case "home":
            return <Home />;
          case "usuarios":
            return <Usuarios user={user} />;
          case "alunos":
            return <Alunos user={user} />;
          case "catraca":
            return <Catraca />;
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
        }
      }}
    </Dashboard>
  );
}

export default App;

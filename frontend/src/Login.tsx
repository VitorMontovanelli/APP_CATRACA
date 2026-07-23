import { useState } from "react";
import { Login as WailsLogin } from "../wailsjs/go/main/App";
import "./Login.css";

interface User {
  id: number;
  name: string;
  email: string;
  cargo: string;
}

interface LoginProps {
  onSuccess?: (user: User) => void;
}

function FingerprintIcon() {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M12 2a10 10 0 0 0-7.07 2.93" />
      <path d="M12 6a6 6 0 0 0-4.24 1.76" />
      <path d="M12 10a2 2 0 0 0-1.41 3.41" />
      <path d="M2.93 16.93A10 10 0 0 1 2 12" />
      <path d="M21.07 7.07A10 10 0 0 1 22 12" />
      <path d="M4.93 19.07A10 10 0 0 1 2 17" />
      <path d="M19.07 4.93A10 10 0 0 1 22 7" />
      <path d="M12 18a10 10 0 0 1-4.93-1.07" />
      <path d="M17 12a5 5 0 0 1-5 5" />
      <path d="M12 14a2 2 0 0 0 0 4" />
    </svg>
  );
}

function Login({ onSuccess }: LoginProps) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setIsLoading(true);
    setErrorMessage("");

    try {
      const user = await WailsLogin(email, password);
      if (onSuccess) {
        onSuccess(user as unknown as User);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setErrorMessage(msg);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="login-page">
      <header className="topbar">
        <FingerprintIcon />
        <span className="topbar-title">AcessoID</span>
      </header>

      <main className="login-content">
        <div className="login-card">
          <h1 className="login-card-title">Entrar no painel</h1>
          <p className="login-card-subtitle">
            Use suas credenciais de administrador.
          </p>

          <form className="login-form" onSubmit={handleSubmit}>
            <label className="login-label" htmlFor="email">
              E-mail
            </label>
            <input
              id="email"
              className="login-input"
              type="email"
              placeholder="seu@email.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />

            <label className="login-label" htmlFor="password">
              Senha
            </label>
            <input
              id="password"
              className="login-input"
              type="password"
              placeholder="Sua senha"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />

            {errorMessage && (
              <p className="login-error">{errorMessage}</p>
            )}

            <button
              className="login-button"
              type="submit"
              disabled={isLoading}
            >
              {isLoading ? "Entrando..." : "Entrar"}
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}

export default Login;

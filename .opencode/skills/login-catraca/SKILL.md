---
name: login-catraca
description: Use when creating or recreating a login screen for the catraca (biometric access control) system. Provides a dark-themed React + TypeScript + CSS login component with fingerprint icon, email/password fields, loading and error states.
---

# Login Screen — Catraca (Acesso Biométrico)

Reusable dark-themed login screen for the catraca desktop app (Tauri + React + TypeScript).

## Files

### `src/Login.tsx`

```tsx
import { useState } from "react";
import "./Login.css";

interface LoginProps {
  onSubmit: (email: string, password: string) => void;
  isLoading?: boolean;
  errorMessage?: string;
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

function Login({ onSubmit, isLoading, errorMessage }: LoginProps) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSubmit(email, password);
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
```

### `src/Login.css`

```css
.login-page {
  background-color: #0b1220;
}

.topbar {
  height: 44px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px;
  border-bottom: 1px solid #1c2537;
  color: #5eead4;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
  background-color: #0b1220;
}

.topbar-title {
  font-size: 13px;
  font-weight: 500;
  color: #e5e7eb;
}

.login-content {
  position: fixed;
  inset: 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

.login-card {
  width: 100%;
  max-width: 340px;
  background-color: #121a29;
  border: 1px solid #232e42;
  border-radius: 12px;
  padding: 28px 26px;
}

.login-card-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
  color: #f4f6f8;
}

.login-card-subtitle {
  margin: 4px 0 20px 0;
  font-size: 13px;
  color: #7c8798;
}

.login-form {
  display: flex;
  flex-direction: column;
}

.login-label {
  font-size: 12px;
  color: #a7b0bf;
  margin-bottom: 4px;
}

.login-input {
  height: 38px;
  padding: 0 12px;
  background-color: #0e1420;
  border: 1px solid #232e42;
  border-radius: 6px;
  color: #e5e7eb;
  font-size: 14px;
  outline: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
  width: 100%;
  font-family: inherit;
}

.login-input + .login-label {
  margin-top: 14px;
}

.login-input:focus {
  border-color: #5eead4;
}

.login-input::placeholder {
  color: #4a5570;
}

.login-error {
  color: #f87171;
  font-size: 12px;
  margin: 8px 0 0 0;
}

.login-button {
  height: 38px;
  width: 100%;
  margin-top: 18px;
  background-color: #5eead4;
  color: #053b32;
  font-size: 14px;
  font-weight: 500;
  font-family: inherit;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: opacity 0.15s;
}

.login-button:hover {
  opacity: 0.9;
}

.login-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
```

### `src/App.css`

Global reset and body background:

```css
*, *::before, *::after {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #root {
  height: 100%;
  background-color: #0b1220;
}

:root {
  font-family: Arial, sans-serif;
  font-size: 12px;
  line-height: 1.4;
  font-weight: 400;
  color: #e5e7eb;
  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  -webkit-text-size-adjust: 100%;
}
```

Color palette reference:

| Token               | Hex       | Usage                    |
|---------------------|-----------|--------------------------|
| `bg-page`           | `#0b1220` | Page/body background     |
| `bg-card`           | `#121a29` | Card background          |
| `bg-input`          | `#0e1420` | Input field background   |
| `border-default`    | `#232e42` | Card and input borders   |
| `border-focus`      | `#5eead4` | Input focus border       |
| `accent`            | `#5eead4` | Button + icon color      |
| `text-primary`      | `#e5e7eb` | Primary text             |
| `text-secondary`    | `#7c8798` | Secondary/muted text     |
| `text-label`        | `#a7b0bf` | Form labels              |
| `text-placeholder`  | `#4a5570` | Input placeholders       |
| `text-error`        | `#f87171` | Error message            |
| `text-button`       | `#053b32` | Button text              |
| `topbar-divider`    | `#1c2537` | Topbar bottom border     |

import { useState, useEffect } from "react";
import {
  ObterConfigTelegram,
  SalvarConfigTelegram,
  TestarTelegram,
  EnviarBackupTelegram,
} from "../../wailsjs/go/main/App";

interface BackupConfig {
  token: string;
  chat_id: string;
  auto_backup: boolean;
  last_backup: string;
  updated_at: string;
}

function formatLast(t: string): string {
  if (!t) return "Nenhum backup enviado ainda";
  const d = new Date(t);
  if (isNaN(d.getTime())) return t;
  return d.toLocaleString("pt-BR");
}

function Backup() {
  const [token, setToken] = useState("");
  const [chatId, setChatId] = useState("");
  const [autoBackup, setAutoBackup] = useState(false);
  const [lastBackup, setLastBackup] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    load();
  }, []);

  async function load() {
    try {
      setLoading(true);
      const cfg = (await ObterConfigTelegram()) as unknown as BackupConfig;
      setToken(cfg.token || "");
      setChatId(cfg.chat_id || "");
      setAutoBackup(cfg.auto_backup || false);
      setLastBackup(cfg.last_backup || "");
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  async function handleSave() {
    setError("");
    setSuccess("");
    setSaving(true);
    try {
      await SalvarConfigTelegram(token, chatId, autoBackup);
      setSuccess("Configuração do backup salva com sucesso!");
      await load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  }

  async function handleTest() {
    setError("");
    setSuccess("");
    setTesting(true);
    try {
      await SalvarConfigTelegram(token, chatId, autoBackup);
      await TestarTelegram();
      setSuccess("Conexão com o Telegram validada com sucesso!");
    } catch (e) {
      setError(String(e));
    } finally {
      setTesting(false);
    }
  }

  async function handleSend() {
    if (!confirm("Enviar backup do banco de dados para o Telegram agora?")) return;
    setError("");
    setSuccess("");
    setSending(true);
    try {
      await SalvarConfigTelegram(token, chatId, autoBackup);
      const when = (await EnviarBackupTelegram()) as unknown as string;
      setLastBackup(when || "");
      setSuccess("Backup enviado para o Telegram com sucesso!");
    } catch (e) {
      setError(String(e));
    } finally {
      setSending(false);
    }
  }

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 48, color: "var(--text-muted)", animation: "fadeIn 0.3s ease-out" }}>
        Carregando configurações de backup...
      </div>
    );
  }

  return (
    <div style={{ animation: "fadeIn 0.3s ease-out" }}>
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 24, fontWeight: 700, margin: 0, letterSpacing: "-0.5px" }}>Backup Telegram</h2>
        <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 4 }}>
          Envie cópias de segurança do banco de dados da academia diretamente para o seu Telegram.
        </p>
      </div>

      {error && (
        <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 16 }}>
          <span>⚠️ {error}</span>
        </div>
      )}
      {success && (
        <div className="vmd-alert vmd-alert-success" style={{ marginBottom: 16 }}>
          <span>✓ {success}</span>
        </div>
      )}

      <div className="vmd-card" style={{ marginBottom: 16, display: "flex", flexDirection: "column", gap: 16 }}>
        <div>
          <h3 style={{ fontSize: 15, fontWeight: 600, margin: 0 }}>Configuração do Bot</h3>
          <p style={{ color: "var(--text-muted)", fontSize: 12, marginTop: 4 }}>
            Crie um bot no Telegram com o <strong>@BotFather</strong>, obtenha o token e adicione o bot ao chat de destino.
          </p>
        </div>

        <div>
          <label className="vmd-label">Bot Token</label>
          <input
            type="password"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="123456789:AA...-token-do-bot"
            className="vmd-input"
            style={{ fontFamily: "monospace" }}
          />
          <div style={{ color: "var(--text-muted)", fontSize: 11, marginTop: 4 }}>
            Obtido no @BotFather (formato <code>123456:ABC-...&lt;token&gt;</code>).
          </div>
        </div>

        <div>
          <label className="vmd-label">Chat ID</label>
          <input
            value={chatId}
            onChange={(e) => setChatId(e.target.value)}
            placeholder="Ex.: -1001234567890 (grupo) ou 123456789 (usuário)"
            className="vmd-input"
            style={{ fontFamily: "monospace" }}
          />
          <div style={{ color: "var(--text-muted)", fontSize: 11, marginTop: 4 }}>
            Para descobrir, use o bot <strong>@userinfobot</strong> ou o endpoint getUpdates da API.
          </div>
        </div>

        <label
          style={{
            display: "flex",
            alignItems: "center",
            gap: 12,
            padding: "12px 16px",
            borderRadius: "var(--radius-md)",
            border: "1px solid var(--border-color)",
            cursor: "pointer",
            backgroundColor: autoBackup ? "rgba(199, 157, 51, 0.03)" : "transparent",
          }}
        >
          <input
            type="checkbox"
            checked={autoBackup}
            onChange={(e) => setAutoBackup(e.target.checked)}
            style={{ width: 16, height: 16, accentColor: "var(--primary)", cursor: "pointer" }}
          />
          <div>
            <div style={{ fontWeight: 600, color: "var(--text-main)", fontSize: 13 }}>Backup automático ao fechar o app</div>
            <div style={{ color: "var(--text-muted)", fontSize: 11, marginTop: 2 }}>
              Envia uma cópia de segurança sempre que o aplicativo for encerrado.
            </div>
          </div>
        </label>

        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <button onClick={handleSave} disabled={saving} className="vmd-btn vmd-btn-primary">
            {saving ? "Salvando..." : "💾 Salvar Configuração"}
          </button>
          <button onClick={handleTest} disabled={testing} className="vmd-btn vmd-btn-secondary">
            {testing ? "Testando..." : "🔌 Testar Conexão"}
          </button>
          <button onClick={handleSend} disabled={sending} className="vmd-btn vmd-btn-success">
            {sending ? "Enviando..." : "📤 Enviar Backup Agora"}
          </button>
        </div>
      </div>

      <div className="vmd-card" style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 12 }}>
        <div>
          <div style={{ fontSize: 15, fontWeight: 600, color: "var(--text-main)" }}>Último backup enviado</div>
          <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 4 }}>{formatLast(lastBackup)}</div>
        </div>
        <span className="vmd-badge" style={{ fontSize: 11, background: "var(--bg-active)", color: "var(--text-main)" }}>
          ☁️ Telegram
        </span>
      </div>
    </div>
  );
}

export default Backup;

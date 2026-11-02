import { useState } from "react";
import { RedefinirSenhaUsuario } from "../../../wailsjs/go/main/App";

interface User {
  id: number;
  name: string;
}

interface RedefinirSenhaModalProps {
  user: User;
  onClose: () => void;
  onSaved: () => void;
}

export default function RedefinirSenhaModal({ user, onClose, onSaved }: RedefinirSenhaModalProps) {
  const [novaSenha, setNovaSenha] = useState("");
  const [confirmaSenha, setConfirmaSenha] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (novaSenha !== confirmaSenha) {
      setError("As senhas não coincidem.");
      return;
    }
    if (novaSenha.length < 4) {
      setError("A senha deve ter no mínimo 4 caracteres.");
      return;
    }

    setLoading(true);
    try {
      await RedefinirSenhaUsuario(user.id, novaSenha);
      setSuccess("Senha atualizada com sucesso!");
      setTimeout(() => {
        onSaved();
        onClose();
      }, 1000);
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="vmd-modal-overlay" onClick={onClose}>
      <form
        onSubmit={handleSubmit}
        className="vmd-modal-content"
        style={{ maxWidth: 380 }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="vmd-modal-header">
          <h3 style={{ fontSize: 16, margin: 0 }}>Alterar Senha: {user.name}</h3>
          <button type="button" className="vmd-modal-close" onClick={onClose}>✕</button>
        </div>

        <div className="vmd-form-group" style={{ marginTop: 16 }}>
          <label className="vmd-label">Nova Senha</label>
          <input
            className="vmd-input"
            type="password"
            placeholder="••••••••"
            value={novaSenha}
            onChange={(e) => setNovaSenha(e.target.value)}
            required
          />
        </div>

        <div className="vmd-form-group">
          <label className="vmd-label">Confirmar Nova Senha</label>
          <input
            className="vmd-input"
            type="password"
            placeholder="••••••••"
            value={confirmaSenha}
            onChange={(e) => setConfirmaSenha(e.target.value)}
            required
          />
        </div>

        {error && <div className="vmd-alert vmd-alert-danger" style={{ marginBottom: 12 }}>⚠️ {error}</div>}
        {success && <div className="vmd-alert vmd-alert-success" style={{ marginBottom: 12 }}>✅ {success}</div>}

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <button type="button" className="vmd-btn vmd-btn-secondary" onClick={onClose}>
            Cancelar
          </button>
          <button type="submit" className="vmd-btn vmd-btn-primary" disabled={loading}>
            {loading ? "Salvando..." : "Salvar Nova Senha"}
          </button>
        </div>
      </form>
    </div>
  );
}

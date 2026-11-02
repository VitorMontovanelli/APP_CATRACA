import { useState } from "react";
import { CriarUsuario } from "../../../wailsjs/go/main/App";

interface NovoUsuarioModalProps {
  allowedCargos: string[];
  onClose: () => void;
  onCreated: () => void;
}

export default function NovoUsuarioModal({ allowedCargos, onClose, onCreated }: NovoUsuarioModalProps) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [cargo, setCargo] = useState(allowedCargos[0] || "admin");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await CriarUsuario(name, email, senha, cargo, 0);
      onCreated();
      onClose();
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="vmd-modal-overlay" onClick={onClose}>
      <form onSubmit={handleCreate} className="vmd-modal-content" style={{ maxWidth: 420 }} onClick={(e) => e.stopPropagation()}>
        <div className="vmd-modal-header">
          <h3 style={{ fontSize: 16, margin: 0 }}>Criar Novo Usuário</h3>
          <button type="button" className="vmd-modal-close" onClick={onClose}>✕</button>
        </div>

        {error && <div className="vmd-alert vmd-alert-danger" style={{ margin: "12px 0" }}>⚠️ {error}</div>}

        <div className="vmd-form-group" style={{ marginTop: 12 }}>
          <label className="vmd-label">Nome Completo</label>
          <input className="vmd-input" value={name} onChange={(e) => setName(e.target.value)} required />
        </div>

        <div className="vmd-form-group">
          <label className="vmd-label">Usuário ou E-mail</label>
          <input className="vmd-input" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </div>

        <div className="vmd-form-group">
          <label className="vmd-label">Senha Inicial</label>
          <input className="vmd-input" type="password" value={senha} onChange={(e) => setSenha(e.target.value)} required />
        </div>

        <div className="vmd-form-group">
          <label className="vmd-label">Cargo</label>
          <select value={cargo} onChange={(e) => setCargo(e.target.value)} className="vmd-select">
            {allowedCargos.map((c) => (
              <option key={c} value={c}>{c === "admin" ? "Administrador" : c}</option>
            ))}
          </select>
        </div>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <button type="button" className="vmd-btn vmd-btn-secondary" onClick={onClose}>Cancelar</button>
          <button type="submit" className="vmd-btn vmd-btn-primary" disabled={loading}>
            {loading ? "Criando..." : "Criar Usuário"}
          </button>
        </div>
      </form>
    </div>
  );
}

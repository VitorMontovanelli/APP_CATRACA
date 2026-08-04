import { useState, useEffect, useRef } from "react";
import { MeuPerfil, AlterarSenha, AlterarFoto } from "../../wailsjs/go/main/App";

interface User {
  id: number;
  name: string;
  email: string;
  cargo: string;
  foto?: string;
}

interface PerfilProps {
  user: User;
  show: boolean;
  onClose: () => void;
  onFotoUpdate?: (foto: string) => void;
}

function PerfilModal({ user, show, onClose, onFotoUpdate }: PerfilProps) {
  const [profile, setProfile] = useState<User>(user);
  const [senhaAtual, setSenhaAtual] = useState("");
  const [novaSenha, setNovaSenha] = useState("");
  const [confirmarSenha, setConfirmarSenha] = useState("");
  const [msg, setMsg] = useState("");
  const [msgType, setMsgType] = useState<"success" | "error">("success");
  const fileRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (show) {
      MeuPerfil().then((u) => setProfile(u as unknown as User)).catch(console.error);
      setSenhaAtual(""); setNovaSenha(""); setConfirmarSenha(""); setMsg("");
    }
  }, [show]);

  async function handleFoto(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = async () => {
      try {
        const base64 = reader.result as string;
        await AlterarFoto(base64);
        setProfile((prev) => prev ? { ...prev, foto: base64 } : prev);
        if (onFotoUpdate) onFotoUpdate(base64);
        mostrarMsg("Foto atualizada com sucesso!", "success");
      } catch (err) {
        mostrarMsg(String(err), "error");
      }
    };
    reader.readAsDataURL(file);
  }

  async function handleSenha(e: React.FormEvent) {
    e.preventDefault();
    setMsg("");
    if (novaSenha !== confirmarSenha) {
      mostrarMsg("Nova senha e confirmação não conferem", "error");
      return;
    }
    if (novaSenha.length < 4) {
      mostrarMsg("Nova senha deve ter no mínimo 4 caracteres", "error");
      return;
    }
    try {
      await AlterarSenha(senhaAtual, novaSenha);
      mostrarMsg("Senha alterada com sucesso!", "success");
      setSenhaAtual(""); setNovaSenha(""); setConfirmarSenha("");
    } catch (err) {
      mostrarMsg(String(err), "error");
    }
  }

  function mostrarMsg(texto: string, tipo: "success" | "error") {
    setMsg(texto);
    setMsgType(tipo);
    setTimeout(() => setMsg(""), 4000);
  }

  if (!show) return null;

  const cargoLabels: Record<string, string> = {
    super_admin: "Super Administrador",
    admin: "Administrador",
    aluno: "Aluno",
  };

  return (
    <div className="vmd-modal-overlay" onClick={onClose}>
      <div className="vmd-modal-content" style={{ maxWidth: 500 }} onClick={(e) => e.stopPropagation()}>
        <div className="vmd-modal-header">
          <h3 style={{ fontSize: 16, margin: 0 }}>Meu Perfil de Acesso</h3>
          <button onClick={onClose} className="vmd-btn vmd-btn-ghost" style={{ padding: 4, minWidth: "auto" }}>
            ✕
          </button>
        </div>

        <div className="vmd-modal-body">
          {msg && (
            <div className={`vmd-alert ${msgType === "success" ? "vmd-alert-success" : "vmd-alert-danger"}`} style={{ marginBottom: 20 }}>
              <span>{msgType === "success" ? "✨" : "⚠️"} {msg}</span>
            </div>
          )}

          {profile && (
            <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
              {/* Profile Card Header */}
              <div className="vmd-card" style={{ padding: 20, background: "rgba(255, 255, 255, 0.01)" }}>
                <div style={{ display: "flex", alignItems: "center", gap: 20 }}>
                  <div
                    style={{
                      width: 72,
                      height: 72,
                      borderRadius: "50%",
                      background: "var(--bg-input)",
                      border: "2px solid var(--border-color)",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      overflow: "hidden",
                      cursor: "pointer",
                      flexShrink: 0,
                      position: "relative",
                    }}
                    onClick={() => fileRef.current?.click()}
                  >
                    {profile.foto ? (
                      <img src={profile.foto} alt="Perfil" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                    ) : (
                      <span style={{ fontSize: 28, fontWeight: 700, color: "var(--primary)" }}>
                        {profile.name.charAt(0).toUpperCase()}
                      </span>
                    )}
                  </div>
                  <input ref={fileRef} type="file" accept="image/*" onChange={handleFoto} style={{ display: "none" }} />
                  
                  <div>
                    <div style={{ fontSize: 16, fontWeight: 600, color: "var(--text-main)" }}>{profile.name}</div>
                    <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>{profile.email}</div>
                    <div style={{ marginTop: 6 }}>
                      <span className="vmd-badge vmd-badge-secondary">
                        {cargoLabels[profile.cargo] || profile.cargo}
                      </span>
                    </div>
                    <button
                      onClick={() => fileRef.current?.click()}
                      className="vmd-btn vmd-btn-secondary"
                      style={{ padding: "4px 10px", height: 26, fontSize: 11, marginTop: 8 }}
                    >
                      Alterar Foto
                    </button>
                  </div>
                </div>
              </div>

              {/* Password Form Card */}
              <div className="vmd-card" style={{ padding: 20, background: "rgba(255, 255, 255, 0.01)" }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: "var(--text-main)", marginBottom: 14 }}>
                  Alterar Senha de Acesso
                </div>
                <form onSubmit={handleSenha} style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                  <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                    <label className="vmd-label">Senha Atual</label>
                    <input
                      type="password"
                      value={senhaAtual}
                      onChange={(e) => setSenhaAtual(e.target.value)}
                      required
                      placeholder="Sua senha atual"
                      className="vmd-input"
                    />
                  </div>
                  
                  <div className="vmd-grid-2">
                    <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                      <label className="vmd-label">Nova Senha</label>
                      <input
                        type="password"
                        value={novaSenha}
                        onChange={(e) => setNovaSenha(e.target.value)}
                        required
                        placeholder="Mín. 4 caracteres"
                        className="vmd-input"
                      />
                    </div>
                    <div className="vmd-form-group" style={{ marginBottom: 0 }}>
                      <label className="vmd-label">Confirmar Senha</label>
                      <input
                        type="password"
                        value={confirmarSenha}
                        onChange={(e) => setConfirmarSenha(e.target.value)}
                        required
                        placeholder="Repita a senha"
                        className="vmd-input"
                      />
                    </div>
                  </div>
                  
                  <button
                    type="submit"
                    className="vmd-btn vmd-btn-primary"
                    style={{ marginTop: 8, height: 38 }}
                  >
                    Confirmar Alteração de Senha
                  </button>
                </form>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default PerfilModal;

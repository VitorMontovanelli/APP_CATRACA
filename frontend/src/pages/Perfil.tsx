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
}

function PerfilModal({ user, show, onClose }: PerfilProps) {
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
        mostrarMsg("Foto atualizada!", "success");
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

  const labelStyle: React.CSSProperties = {
    fontSize: 12, color: "#7c8798", marginBottom: 4,
  };

  const inputStyle: React.CSSProperties = {
    height: 38, padding: "0 12px", background: "#0e1420",
    border: "1px solid #232e42", borderRadius: 6, color: "#e5e7eb",
    fontSize: 13, fontFamily: "inherit", outline: "none",
    width: "100%", boxSizing: "border-box",
  };

  return (
    <div style={{
      position: "fixed", inset: 0, background: "rgba(0,0,0,0.7)",
      display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
    }} onClick={onClose}>
      <div style={{
        background: "#121a29", border: "1px solid #232e42", borderRadius: 12,
        padding: 28, width: 500, maxWidth: "90vw", maxHeight: "90vh", overflow: "auto",
      }} onClick={(e) => e.stopPropagation()}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 500 }}>Meu Perfil</h2>
          <button onClick={onClose} style={{
            background: "none", border: "none", color: "#7c8798",
            fontSize: 20, cursor: "pointer", fontFamily: "inherit",
            padding: "4px 8px", borderRadius: 4,
          }}>✕</button>
        </div>

        {msg && (
          <div style={{
            padding: "10px 14px", borderRadius: 6, fontSize: 13, marginBottom: 16,
            background: msgType === "success" ? "#5eead422" : "#f8717122",
            color: msgType === "success" ? "#5eead4" : "#f87171",
            border: `1px solid ${msgType === "success" ? "#5eead4" : "#f87171"}`,
          }}>{msg}</div>
        )}

        {profile && (
          <>
            <div style={{
              background: "#0e1420", border: "1px solid #232e42", borderRadius: 10,
              padding: 20, marginBottom: 20,
            }}>
              <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
                <div style={{
                  width: 64, height: 64, borderRadius: "50%",
                  background: "#0e1420", border: "2px solid #232e42",
                  display: "flex", alignItems: "center", justifyContent: "center",
                  overflow: "hidden", cursor: "pointer", flexShrink: 0,
                }} onClick={() => fileRef.current?.click()}>
                  {profile.foto ? (
                    <img src={profile.foto} alt="foto"
                      style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                  ) : (
                    <span style={{ fontSize: 24, color: "#7c8798" }}>
                      {profile.name.charAt(0).toUpperCase()}
                    </span>
                  )}
                </div>
                <input ref={fileRef} type="file" accept="image/*"
                  onChange={handleFoto} style={{ display: "none" }} />
                <div>
                  <div style={{ fontSize: 15, fontWeight: 500 }}>{profile.name}</div>
                  <div style={{ fontSize: 12, color: "#7c8798" }}>{profile.email}</div>
                  <div style={{ fontSize: 11, color: "#60a5fa", textTransform: "uppercase", marginTop: 2 }}>
                    {profile.cargo}
                  </div>
                  <button onClick={() => fileRef.current?.click()} style={{
                    marginTop: 6, padding: "3px 10px", border: "1px solid #60a5fa",
                    borderRadius: 6, background: "transparent", color: "#60a5fa",
                    fontSize: 11, fontFamily: "inherit", cursor: "pointer",
                  }}>Alterar Foto</button>
                </div>
              </div>
            </div>

            <div style={{
              background: "#0e1420", border: "1px solid #232e42", borderRadius: 10, padding: 20,
            }}>
              <h3 style={{ margin: "0 0 16px 0", fontSize: 14, fontWeight: 500, color: "#a7b0bf" }}>
                Alterar Senha
              </h3>
              <form onSubmit={handleSenha} style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                <div>
                  <div style={labelStyle}>Senha Atual</div>
                  <input type="password" value={senhaAtual}
                    onChange={(e) => setSenhaAtual(e.target.value)} required style={inputStyle} />
                </div>
                <div style={{ display: "flex", gap: 12 }}>
                  <div style={{ flex: 1 }}>
                    <div style={labelStyle}>Nova Senha</div>
                    <input type="password" value={novaSenha}
                      onChange={(e) => setNovaSenha(e.target.value)} required style={inputStyle} />
                  </div>
                  <div style={{ flex: 1 }}>
                    <div style={labelStyle}>Confirmar</div>
                    <input type="password" value={confirmarSenha}
                      onChange={(e) => setConfirmarSenha(e.target.value)} required style={inputStyle} />
                  </div>
                </div>
                <button type="submit" style={{
                  padding: "10px", background: "#5eead4", color: "#053b32",
                  border: "none", borderRadius: 6, fontSize: 13, fontWeight: 500,
                  fontFamily: "inherit", cursor: "pointer", marginTop: 4,
                }}>Salvar Senha</button>
              </form>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default PerfilModal;

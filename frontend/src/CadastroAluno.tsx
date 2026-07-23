import { useState, useCallback } from "react";
import "./CadastroAluno.css";

interface CadastroAlunoProps {
  onSubmit: (dados: {
    nome: string;
    cpf: string;
    dataNascimento: string;
    email: string;
    telefone: string;
    plano: string;
    status: string;
  }) => void;
  onCancel: () => void;
  isLoading?: boolean;
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

function formatCPF(value: string): string {
  const digits = value.replace(/\D/g, "").slice(0, 11);
  return digits
    .replace(/^(\d{3})(\d)/, "$1.$2")
    .replace(/^(\d{3})\.(\d{3})(\d)/, "$1.$2.$3")
    .replace(/^(\d{3})\.(\d{3})\.(\d{3})(\d)/, "$1.$2.$3-$4");
}

function formatPhone(value: string): string {
  const digits = value.replace(/\D/g, "").slice(0, 11);
  if (digits.length <= 2) return `(${digits}`;
  if (digits.length <= 7)
    return `(${digits.slice(0, 2)}) ${digits.slice(2)}`;
  return `(${digits.slice(0, 2)}) ${digits.slice(2, 7)}-${digits.slice(7)}`;
}

function validarCPF(cpf: string): boolean {
  const digits = cpf.replace(/\D/g, "");
  if (digits.length !== 11) return false;
  if (/^(\d)\1{10}$/.test(digits)) return false;
  let sum = 0;
  for (let i = 0; i < 9; i++) sum += parseInt(digits[i]) * (10 - i);
  let rest = (sum * 10) % 11;
  if (rest === 10) rest = 0;
  if (rest !== parseInt(digits[9])) return false;
  sum = 0;
  for (let i = 0; i < 10; i++) sum += parseInt(digits[i]) * (11 - i);
  rest = (sum * 10) % 11;
  if (rest === 10) rest = 0;
  return rest === parseInt(digits[10]);
}

const planos = ["Mensal", "Trimestral", "Semestral", "Anual"];

const statusList = [
  { value: "ativo", label: "Ativo" },
  { value: "inadimplente", label: "Inadimplente" },
  { value: "suspenso", label: "Suspenso" },
];

const statusDotColor: Record<string, string> = {
  ativo: "#5eead4",
  inadimplente: "#f87171",
  suspenso: "#7c8798",
};

function CadastroAluno({ onSubmit, onCancel, isLoading }: CadastroAlunoProps) {
  const [nome, setNome] = useState("");
  const [cpf, setCpf] = useState("");
  const [dataNascimento, setDataNascimento] = useState("");
  const [email, setEmail] = useState("");
  const [telefone, setTelefone] = useState("");
  const [plano, setPlano] = useState("Mensal");
  const [status, setStatus] = useState("ativo");
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleCpfChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setCpf(formatCPF(e.target.value));
    },
    []
  );

  const handleTelefoneChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setTelefone(formatPhone(e.target.value));
    },
    []
  );

  function validate(): boolean {
    const errs: Record<string, string> = {};

    if (!nome.trim()) errs.nome = "Nome é obrigatório";

    const cpfDigits = cpf.replace(/\D/g, "");
    if (cpfDigits.length !== 11 || !validarCPF(cpf))
      errs.cpf = "CPF inválido";

    if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email))
      errs.email = "E-mail inválido";

    setErrors(errs);
    return Object.keys(errs).length === 0;
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!validate()) return;

    onSubmit({
      nome: nome.trim(),
      cpf: cpf.replace(/\D/g, ""),
      dataNascimento,
      email,
      telefone,
      plano,
      status,
    });
  }

  return (
    <div className="cadastro-page">
      <header className="cadastro-topbar">
        <FingerprintIcon />
        <span className="cadastro-topbar-title">AcessoID</span>
      </header>

      <main className="cadastro-content">
        <div className="cadastro-card">
          <h1 className="cadastro-title">Cadastrar aluno</h1>
          <p className="cadastro-subtitle">
            Preencha os dados básicos do novo aluno
          </p>

          <form className="cadastro-form" onSubmit={handleSubmit}>
            <div className="cadastro-field cadastro-field-full">
              <label className="cadastro-label">Nome completo</label>
              <input
                className={`cadastro-input ${errors.nome ? "cadastro-input-error" : ""}`}
                type="text"
                placeholder="Nome do aluno"
                value={nome}
                onChange={(e) => setNome(e.target.value)}
              />
              {errors.nome && (
                <span className="cadastro-field-error">{errors.nome}</span>
              )}
            </div>

            <div className="cadastro-field">
              <label className="cadastro-label">CPF</label>
              <input
                className={`cadastro-input ${errors.cpf ? "cadastro-input-error" : ""}`}
                type="text"
                placeholder="000.000.000-00"
                value={cpf}
                onChange={handleCpfChange}
              />
              {errors.cpf && (
                <span className="cadastro-field-error">{errors.cpf}</span>
              )}
            </div>

            <div className="cadastro-field">
              <label className="cadastro-label">Data de nascimento</label>
              <input
                className="cadastro-input"
                type="text"
                placeholder="dd/mm/aaaa"
                value={dataNascimento}
                onChange={(e) => setDataNascimento(e.target.value)}
              />
            </div>

            <div className="cadastro-field">
              <label className="cadastro-label">E-mail</label>
              <input
                className={`cadastro-input ${errors.email ? "cadastro-input-error" : ""}`}
                type="email"
                placeholder="aluno@email.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
              {errors.email && (
                <span className="cadastro-field-error">{errors.email}</span>
              )}
            </div>

            <div className="cadastro-field">
              <label className="cadastro-label">Telefone</label>
              <input
                className="cadastro-input"
                type="text"
                placeholder="(00) 00000-0000"
                value={telefone}
                onChange={handleTelefoneChange}
              />
            </div>

            <div className="cadastro-field">
              <label className="cadastro-label">Plano</label>
              <select
                className="cadastro-select"
                value={plano}
                onChange={(e) => setPlano(e.target.value)}
              >
                {planos.map((p) => (
                  <option key={p} value={p}>
                    {p}
                  </option>
                ))}
              </select>
            </div>

            <div className="cadastro-field">
              <label className="cadastro-label">Status</label>
              <div className="cadastro-select-wrapper">
                <select
                  className="cadastro-select cadastro-select-status"
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                >
                  {statusList.map((s) => (
                    <option key={s.value} value={s.value}>
                      {s.label}
                    </option>
                  ))}
                </select>
                <span
                  className="cadastro-status-dot"
                  style={{ backgroundColor: statusDotColor[status] || "#7c8798" }}
                />
              </div>
            </div>

            <div className="cadastro-footer">
              <button
                type="button"
                className="cadastro-btn cadastro-btn-cancel"
                onClick={onCancel}
              >
                Cancelar
              </button>
              <button
                type="submit"
                className="cadastro-btn cadastro-btn-submit"
                disabled={isLoading}
              >
                {isLoading ? "Salvando..." : "Salvar aluno"}
              </button>
            </div>
          </form>
        </div>
      </main>
    </div>
  );
}

export default CadastroAluno;

import { useEffect, useState, type CSSProperties } from "react";

interface UserAvatarProps {
  name: string;
  foto?: string | null;
  size?: number;
  fontSize?: number;
  className?: string;
  style?: CSSProperties;
  onClick?: () => void;
  title?: string;
}

export default function UserAvatar({
  name,
  foto,
  size = 38,
  fontSize,
  className,
  style,
  onClick,
  title,
}: UserAvatarProps) {
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    setFailed(false);
  }, [foto]);

  const inicial = (name || "").trim().charAt(0).toUpperCase() || "?";
  const mostrarFoto = Boolean(foto) && !failed;

  return (
    <div
      className={className}
      onClick={onClick}
      title={title}
      style={{
        width: size,
        height: size,
        borderRadius: "50%",
        overflow: "hidden",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        flexShrink: 0,
        ...style,
      }}
    >
      {mostrarFoto ? (
        <img
          src={foto as string}
          alt={`Foto de ${name}`}
          onError={() => setFailed(true)}
          style={{ width: "100%", height: "100%", objectFit: "cover", display: "block" }}
        />
      ) : (
        <span
          style={{
            fontSize: fontSize ?? Math.round(size * 0.4),
            fontWeight: 700,
            lineHeight: 1,
            color: "inherit",
          }}
        >
          {inicial}
        </span>
      )}
    </div>
  );
}

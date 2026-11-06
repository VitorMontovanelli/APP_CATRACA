const MAX_DIMENSAO = 256;
const QUALIDADE_JPEG = 0.82;

function carregarImagem(file: File): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file);
    const imagem = new Image();
    imagem.onload = () => {
      URL.revokeObjectURL(url);
      resolve(imagem);
    };
    imagem.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error("Não foi possível ler a imagem selecionada."));
    };
    imagem.src = url;
  });
}

/**
 * Redimensiona a foto para no máximo 256x256 e converte para JPEG
 * comprimido antes de enviar ao backend, para não inflar o banco
 * nem o backup.
 */
export async function prepararFotoPerfil(file: File): Promise<string> {
  if (!file.type.startsWith("image/")) {
    throw new Error("Selecione um arquivo de imagem.");
  }

  const imagem = await carregarImagem(file);
  const larguraOriginal = imagem.naturalWidth || imagem.width;
  const alturaOriginal = imagem.naturalHeight || imagem.height;
  const escala = Math.min(1, MAX_DIMENSAO / Math.max(larguraOriginal, alturaOriginal));
  const largura = Math.max(1, Math.round(larguraOriginal * escala));
  const altura = Math.max(1, Math.round(alturaOriginal * escala));

  const canvas = document.createElement("canvas");
  canvas.width = largura;
  canvas.height = altura;

  const ctx = canvas.getContext("2d");
  if (!ctx) {
    throw new Error("Não foi possível processar a imagem.");
  }

  // Fundo branco para evitar áreas pretas em imagens com transparência.
  ctx.fillStyle = "#ffffff";
  ctx.fillRect(0, 0, largura, altura);
  ctx.drawImage(imagem, 0, 0, largura, altura);

  return canvas.toDataURL("image/jpeg", QUALIDADE_JPEG);
}

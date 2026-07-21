#!/usr/bin/env bash

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$HOME/.local/bin"
TARGET_SCRIPT="$BIN_DIR/walkmoon"

echo "=== Configurando o Walkmoon ==="

if ! command -v go &>/dev/null; then
  echo "[!] Go não foi encontrado no sistema."
  echo "Instale Go manualmente (ex: sudo pacman -S go ou sudo apt isntall golang)"
  exit 1
fi

echo "[✓] Golang detectada: $(go version)"

mkdir -p "$BIN_DIR"

echo "[*] Compilando o walkmoon em walkmoon em $BIN_DIR..."
go build -o "$BIN_DIR/walkmoon" "$PROJECT_DIR/cmd/player"

chmod +x "$TARGET_SCRIPT"

echo "[✓] Walkmoon instalado com sucesso!"
echo ""
echo "Se o comando 'walkmoon' não funcionar imediatamente, garanta que '$BIN_DIR' está no seu PATH"
echo "Cheque o 'README' no repositório para mais dúvidas."

# 🎧 Walkmoon

A music player focused on quality, customization, and efficiency, built with Go.

> ⚠️ **Estado do Projeto:** Pré-MVP. Recursos e funcionalidades estão em desenvolvimento ativo.

---

## 📌 Sobre o Projeto

O **Walkmoon** nasce do desejo de ouvir música em um ambiente leve, integrado à rotina dev e adaptável à personalidade de cada usuário.

---

## 🛠️ Tecnologias Utilizadas

* **Linguagem:** [Go (Golang)](https://golang.org/)
* **Manipulação e Decodificação de Áudio:** [Beep](https://github.com/faiface/beep)
* **Design & Interface:** Ícaro

---

## 🚀 Como Executar

### Pré-requisitos

* Ambiente e scripts desenvolvidos com foco em sistemas operacionais **Linux**.
* Possuir a linguagem **Go** instalada no sistema.

### Passo a Passo

1. Clone o repositório em um diretório de sua preferência:
   ```bash
   git clone [https://github.com/seu-usuario/walkmoon.git](https://github.com/seu-usuario/walkmoon.git)
   cd walkmoon
    
2. Rode o instalador:
   na pasta raiz do projeto, rode
   ```bash
   ./install.sh
   ```
   Pronto. Rode walkmoon no seu terminal de qualquer diretório e a aplicação deve iniciar corretamente.

## Configurações específicas
   O __Walkmoon__ tem a funcionalidade de persistência de alguns dados que são salvos ao fim de cada execução. Os dados incluem:

   - Última música reproduzida e o diretório de origem.
   - Configurações de reprodução (modo aleatório, loop de playlist e loop de música única)
   - Volume.

   Caso deseje alterar esses valores manualmente, você pode editar diretamente o arquivo de configuração localizado em:
   > ~/.config/walkmoon/playerState.json

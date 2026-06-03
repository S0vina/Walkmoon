# Walkmoon
A music player that focus on quality, customization and efficiency.


## Lidando com problemas de seguranca e ambiente ao tentar rodar a Main:

### O problema

Acontece que ao rodar um script go com __go run__, o arquivo eh compilado na pasta onde voce definiu a instalacao da linguagem. Isso, ao menos no windows, pode nos resultar em restricoes de seguranca. 

### A solucao

Compilar e rodar o script a partir do executavel criado dentro da pasta onde o repositorio foi ancorado. Onde a seguranca do sistema faz vista grossa.

No windows (PowerSheel)
```PowerSheel
go build -o player.exe ./cmd/player/main.go; .\player.exe "assets/music"
```

No Linux/MacOs
```bash
go build -o player ./cmd/player/main.go && ./player "assets/music"
```
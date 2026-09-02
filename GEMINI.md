# HydraVault Project Guidelines (GEMINI.md)

Este documento define a arquitetura, design system cyberpunk, catálogo de skills, pipelines de curadoria/active learning e protocolo de auto-reparo do ecossistema **HydraVault** (Cofre de Dados de Visão Computacional, Auto-Rotulagem & Curadoria para YOLO).

---

## 1. Filosofia de Engenharia e Estrutura Arquitetural

O projeto segue estritamente a **Arquitetura Hexagonal (Ports & Adapters)** combinada com **Domain-Driven Design (DDD)**:

```text
hydravault
├── cmd/                          # Entrypoints principais dos binários
│   └── hydravault/main.go
├── internal/
│   ├── domain/                   # Entidades puras e regras de negócio (ZERO imports de infra/HTTP)
│   │   ├── dataset.go
│   │   ├── annotation.go
│   │   ├── edge_case.go
│   │   └── deduplication.go
│   ├── ports/                    # Interfaces de entrada (Driving) e saída (Driven)
│   ├── application/              # Casos de uso, auto-labeling pipeline e orquestração
│   └── adapters/
│       ├── primary/http/         # Controladores REST, gRPC e WebSockets
│       └── secondary/
│           ├── storage/          # Persistência de mídia, anotações e SQLite WAL
│           ├── autolabel/        # Adaptador SAM 2 / YOLO-World Zero-Shot
│           ├── deduplication/    # Extrator pHash e Cosine Similarity
│           └── exporter/         # Gerador de splits versionados (train/val/test data.yaml)
├── datasets/                     # Repositório de dados gerenciado
│   ├── inbox/                    # Frames não curados recebidos do HydraStream
│   └── curated/                  # Datasets aprovados e versionados
├── worker_python/                # Workers Python para SAM 2, embeddings e conversão COCO/YOLO
└── web/                          # Frontend SPA Cyberpunk High-Tech (Vite + React)
```

### Regras Inquebráveis de Backend:
1. **Pureza do Domínio:** O pacote `internal/domain/` **NUNCA** deve importar pacotes de infraestrutura como `net/http`, `database/sql` ou bibliotecas externas de rede.
2. **Concorrência Segura:** Toda operação de curadoria e ingestão de arquivos é sincronizada via mutexes e contexts Go (`context.WithCancel`).
3. **Integridade de Rótulos:** Validação estrita de anotações normalizadas no formato YOLO (`class_id x_center y_center width height` onde todos os floats estão no intervalo `[0.0, 1.0]`).

---

## 2. Design System & Estilo Visual (Cyberpunk High-Tech)

O frontend adota uma estética visual futurista de alta densidade inspirada no universo Cyberpunk 2077:

### Tokens e Cores Primárias:
- **Fundo Profundo:** `#07080c` (Main Background), `#0b0e14` (Surface), `#121824` (Elevated Cards).
- **Cores de Destaque Neon:**
  - Ciano: `#00f0ff` (Primary Glow & Destaques)
  - Amarelo Cyber: `#fcee0a` (Avisos & Títulos Secundários)
  - Magenta: `#ff003c` (Erros, Alertas Críticos & Quedas)
  - Verde Esmeralda: `#00ff9d` (Status Online, Sucesso & Aprovação de Labels)
- **Tipografia Modular:** Google Fonts (`Advent Pro`, `Barlow`, `Tomorrow`, `Oxanium` e `JetBrains Mono`).

### ⚠️ Regra Estrita de Modularização Frontend:
- **Limite Máximo de 100 Linhas por Arquivo:** Nenhum arquivo CSS, JS ou JSX em `web/` pode ultrapassar **100 linhas**.
- Se um componente crescer além de 100 linhas, ele **deve** ser dividido em sub-módulos focados (ex: `LabelCanvas.jsx`, `ClassSelector.jsx`, `DeduplicationCard.jsx`).

---

## 3. Catálogo de Skills & Ecossistema Ultralytics

O projeto possui integração direta com as skills oficiais de IA localizadas em `.agents/skills/`:

| Skill | Finalidade |
| :--- | :--- |
| **`validate-project`** | Executa auditoria de conformidade DDD, contagem de linhas web e testes unitários. |
| **`yolo-datasets`** | Estruturas de `data.yaml`, classes, splits train/val/test, formatos `.txt`, conversões COCO/DOTA e auto-labeling. |
| **`yolo`** | Router principal de comandos e CLI Ultralytics. |
| **`yolo-models`** | Guia de arquiteturas: YOLOv8, YOLO11, YOLO26, SAM 2 e YOLO-World. |
| **`yolo-export`** | Exportação e empacotamento de modelos. |

---

## 4. Pipeline do Ecossistema Hydra

```text
[ HydraStream ] ──► (Frames Ambíguos) ──► [ HydraVault (Curadoria & Auto-Labeling) ]
                                                            │
                                                            ▼ (Dataset 100k+ Limpo)
[ HydraStream ] ◄── (Modelo .engine)  ◄── [ HydraForge (Treino RTX 5090) ]
```

---

## 5. Protocolo de Auto-Reparo & Validação Pós-Prompt (Self-Healing Rules)

Ao final de **qualquer modificação de código**, o agente deve obrigatoriamente executar o seguinte checklist de auto-reparo antes de concluir:

```bash
# 1. Auditoria de Limite de Linhas Web (< 100 linhas por arquivo)
wc -l web/css/*.css web/js/*.js web/src/**/*.css web/src/**/*.jsx 2>/dev/null

# 2. Auditoria de Pureza DDD (Nenhum import de net/http no domínio)
grep -rn "net/http" internal/domain/ && echo "VIOLAÇÃO DDD: Remova net/http do domínio!"

# 3. Compilação e Testes Automatizados
make test && make build
```

# HydraVault Project Guidelines (GEMINI.md)

Este documento define a arquitetura, design system cyberpunk, catálogo de skills, pipelines de curadoria/active learning e protocolo de auto-reparo do **HydraVault** (Cofre de Dados de Visão Computacional, Auto-Rotulagem SAM 2 e Curadoria para os modelos do Marketplace Hydra).

---

## 0. Papel no Ecossistema & Cadeia de Valor de IA

O **HydraVault** é o **Cofre de Curadoria de Dados** do ecossistema:
1. **Ingestão Active Learning:** Recebe frames de baixa confiança enviados em produção pelo **HydraStream**.
2. **Auto-Rotulagem & Curadoria com SAM 2:** Permite anotação zero-shot e aprovação de instâncias com alta precisão.
3. **Exportação Versionada:** Fornece datasets `data.yaml` perfeitamente balanceados para o **HydraForge** treinar novos modelos analíticos do Marketplace.


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

### ⚠️ Regras Estritas de Frontend & Cyberpunk Design System:
- **Modularidade & Coesão:** Componentes `.jsx`/`.tsx` até 150 linhas (teto de 200 linhas para telas de anotação complexas), composables/lógica até 100-120 linhas; dicionários e tipos livres.
- **Proibição Absoluta de Emojis em Listas e Dropdowns:** Nunca usar emojis (como ⚡, 🎥, 📁, ⭐, 🏆, etc.) dentro de `<select>`, `<option>`, dropdowns, tabelas ou listas em nenhum projeto.
- **Linguagem Técnica Militar Cyberpunk:** Listas e opções devem usar terminologia técnica HUD (ex: `[TRAINED] YOLO26M // mAP 58.1%`, `[STREAM] CAM_01 // 1080P @ 30 FPS`, `[OK]`, `// RETICLE`). Ícones visuais devem remeter puramente à estética Cyberpunk HUD e nunca serem embutidos dentro de itens de listagem.

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
# 1. Auditoria de Limite de Linhas Web (Composables <= 120, Views <= 200, Locales/Types isentos)
wc -l web/css/*.css web/js/*.js web/src/**/*.css web/src/**/*.jsx 2>/dev/null | awk '$2 !~ /locales|types/ && $1 > 200 { print "VIOLATION: " $2 " has " $1 " lines (>200)" }'

# 2. Auditoria de Pureza DDD (Nenhum import de net/http no domínio)
grep -rn "net/http" internal/domain/ && echo "VIOLAÇÃO DDD: Remova net/http do domínio!"

# 3. Compilação e Testes Automatizados
make test && make build

# 4. Auditoria de Paridade de Internacionalização (i18n)
# Todo texto ou chave nova adicionada no frontend DEVE existir em todos os idiomas suportados (PT, EN, ES).
```

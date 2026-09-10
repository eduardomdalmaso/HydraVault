# 🏛️ HydraVault — Cofre Inteligente de Dados de Visão & Curadoria Active Learning

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Licença](https://img.shields.io/badge/Licenca-Apache%202.0-blue.svg)](LICENSE)
[![Ecosystem](https://img.shields.io/badge/Hydra-Ecosystem-00f0ff.svg)](https://github.com/eduardomdalmaso)
[![SAM 2](https://img.shields.io/badge/Modelos%20Fundacao-SAM%202%20%2B%20YOLO--World-emerald.svg)](https://github.com/facebookresearch/segment-anything-2)

[**English**](README.md) | [**Português do Brasil**]

> **Curadoria, Deduplicação Perceptual, Auto-Rotulagem com SAM 2 e Versionamento de Datasets para YOLO.**

---

## 🌐 Mapa de Portas do Ecossistema

| Serviço / App | Porta(s) | Papel & Tecnologia |
| :--- | :--- | :--- |
| **`hydra-vault` (Curador)** | `8082` | Control Plane em Go + SPA React Cyberpunk + Worker SAM 2 |
| **`hydra-stream` (Ingestão)** | `8080` | Ingestão Zero-Copy & Fan-Out de Memória Compartilhada (`/dev/shm`) |
| **`hydra-forge` (Treinamento)** | `8081` | Estúdio de Treinamento YOLO & Compilador TensorRT (RTX 5090) |
| **`hydra-vms-api` (Core VMS)** | `8083` | API REST em Go, Multi-Tenancy RLS & NATS Event Bridge |
| **`hydra-vms` (Frontend)** | `5173` | HUD de Operações em Vue 3 + Vite |
| **PostgreSQL 16** | `5432` | Container Podman `hydra_postgres` (Multi-Tenancy RLS) |
| **NATS JetStream** | `4222`, `8222` | Container Podman `hydra_nats` (Barramento de Eventos) |
| **MinIO S3** | `9000`, `9001` | Container Podman `hydra_minio` (Armazenamento de Objetos) |

---

## 🌟 O Papel do HydraVault no Ecossistema Hydra

Em sistemas de visão computacional em produção, **mais de 90% dos quadros de uma câmera de segurança são redundantes ou estáticos** (fundos sem movimento, carros idênticos, corredores vazios). Treinar redes neurais com vídeos brutos desperdiça horas de GPU e gera *overfitting* do modelo.

O **HydraVault é o elo que fecha o ciclo de aprendizado contínuo (*Active Learning Loop*):**

```text
       ┌────────────────────────────────────────────────────────┐
       │                 1. HYDRASTREAM / HYDRAVMS              │
       │  (Ingestão Zero-Copy, Inferência YOLO & Edge Cases)    │
       └──────────────────────────┬─────────────────────────────┘
                                  │
                    (Quadros Ambíguos / Baixa Confiança)
                                  │
                                  ▼
       ┌────────────────────────────────────────────────────────┐
       │                  2. HYDRAVAULT                         │
       │  (Deduplicação, Auto-Rotulagem SAM 2 & Curadoria)      │
       └──────────────────────────┬─────────────────────────────┘
                                  │
                    (Datasets Limpos e Balanceados 100k+)
                                  │
                                  ▼
       ┌────────────────────────────────────────────────────────┐
       │                  3. HYDRAFORGE                         │
       │   (Treinamento YOLO PyTorch na NVIDIA RTX 5090)        │
       └──────────────────────────┬─────────────────────────────┘
                                  │
                  (Modelo Compilado TensorRT .engine)
                                  │
                                  └──────────────► [ Re-implantado no HydraStream / HydraVMS! ]
```

---

## 🚀 Capacidades Principais

1. **📥 Coleta Automática de *Edge Cases*:** Captura automática de frames onde a IA teve baixa confiança (`< 40%`) durante o monitoramento ao vivo.
2. **🤖 Auto-Rotulagem com Modelos de Fundação:** Geração automática de caixas e polígonos usando **SAM 2 (Segment Anything 2)** e **YOLO-World** (*Zero-Shot*), exigindo apenas confirmação em 1-clique do operador humano.
3. **🔍 Deduplicação Perceptual (*pHash* e Cosseno):** Elimina imagens repetidas e redundantes, mantendo apenas exemplos informativos e de alta variância.
4. **⚖️ Auditoria de Saúde do Dataset:** Gráficos em tempo real de distribuição de classes, proporção de objetos pequenos (*small objects*) e integridade de anotações.
5. **📦 Exportação Versionada `data.yaml`:** Geração de splits reproduzíveis (`train`, `val`, `test`) prontos para consumo imediato no **HydraForge**.

---

## 🛠️ Inicialização Rápida

```bash
cd /home/hades/Documents/HydraVault

# 1. Compilar binário do backend
go build -o bin/hydravault ./cmd/hydravault

# 2. Compilar Frontend React
cd web && npm install && npm run build && cd ..

# 3. Iniciar via PM2
pm2 restart hydra-vault
```

---

## 📄 Licença
Distribuído sob a licença Apache 2.0.

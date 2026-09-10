# 🏛️ HydraVault — Intelligent Vision Data Engine & Active Learning Vault

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Ecosystem](https://img.shields.io/badge/Hydra-Ecosystem-00f0ff.svg)](https://github.com/eduardomdalmaso)
[![SAM 2](https://img.shields.io/badge/Foundation%20Model-SAM%202%20%2B%20YOLO--World-emerald.svg)](https://github.com/facebookresearch/segment-anything-2)

[**English**] | [**Português do Brasil**](README.pt-BR.md)

> **Curate, Deduplicate, Auto-Annotate, and Version-Control Vision Datasets at Scale for YOLO Architectures.**

---

## 🌐 Ecosystem Port Map

| Service / App | Port(s) | Role & Technology |
| :--- | :--- | :--- |
| **`hydra-vault` (Curator Engine)** | `8082` | Go Control Plane + React Cyberpunk SPA + SAM 2 Worker |
| **`hydra-stream` (Ingest)** | `8080` | Zero-Copy Frame Ingestion & SHM Fan-Out (`/dev/shm`) |
| **`hydra-forge` (Trainer)** | `8081` | YOLO Training Studio & TensorRT Compiler (RTX 5090) |
| **`hydra-vms-api` (VMS Core)** | `8083` | Go REST API, Multi-Tenancy & NATS Event Bridge |
| **`hydra-vms` (Frontend)** | `5173` | Vue 3 + Vite Live Operations HUD |
| **PostgreSQL 16** | `5432` | Podman container `hydra_postgres` (RLS Multi-Tenancy) |
| **NATS JetStream** | `4222`, `8222` | Podman container `hydra_nats` (Event Mesh) |
| **MinIO S3** | `9000`, `9001` | Podman container `hydra_minio` (Object Storage) |

---

## 🌟 The Role of HydraVault in the Hydra Ecosystem

**HydraVault** is the **curation and active learning vault** of the ecosystem. In production computer vision, 90% of camera frames are redundant (static background, empty hallways, identical cars). Deploying raw video to training pipelines causes model overfitting and wasted GPU compute.

**HydraVault closes the continuous AI improvement loop:**

```text
       ┌────────────────────────────────────────────────────────┐
       │                 1. HYDRASTREAM / HYDRAVMS              │
       │  (Zero-Copy Ingestion, YOLO Runtime & Hard-Negatives)  │
       └──────────────────────────┬─────────────────────────────┘
                                  │
                    (Ambiguous Frames / Low-Confidence)
                                  │
                                  ▼
       ┌────────────────────────────────────────────────────────┐
       │                  2. HYDRAVAULT                         │
       │  (Deduplication, SAM 2 Auto-Labeling & Curation Vault) │
       └──────────────────────────┬─────────────────────────────┘
                                  │
                    (Gold Curated 100k+ Image Datasets)
                                  │
                                  ▼
       ┌────────────────────────────────────────────────────────┐
       │                  3. HYDRAFORGE                         │
       │   (YOLO PyTorch Training Loop on NVIDIA RTX 5090)      │
       └──────────────────────────┬─────────────────────────────┘
                                  │
                  (Optimized TensorRT .engine Checkpoint)
                                  │
                                  └──────────────► [ Deployed back to HydraStream / HydraVMS! ]
```

---

## 🚀 Core Capabilities

1. **📥 Edge Frame Ingestion:** Automatically catches ambiguous frames, edge cases, and low-confidence detections from live cameras.
2. **🤖 Smart AI Pre-Annotation:** Zero-shot foundation model labeling with **SAM 2** (Segment Anything 2) and **YOLO-World**, generating high-quality polygon masks and bounding boxes.
3. **🔍 Perceptual Deduplication (pHash & Cosine Distance):** Identifies and discards visual duplicates, keeping only informative, high-variance frames.
4. **⚖️ Dataset Health Auditing:** Real-time analytics on class balance, bbox aspect ratios, and format verification.
5. **📦 Version-Controlled `data.yaml` Export:** Generates deterministic `train/val/test` splits ready for instant training in **HydraForge**.

---

## 🛠️ Quickstart

```bash
cd /home/hades/Documents/HydraVault

# 1. Build Control Plane binary
go build -o bin/hydravault ./cmd/hydravault

# 2. Build React Web SPA
cd web && npm install && npm run build && cd ..

# 3. Start via PM2
pm2 restart hydra-vault
```

---

## 📄 License
Licensed under the Apache 2.0 License.

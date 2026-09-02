# 🏛️ HydraVault — Intelligent Vision Data Engine & Active Learning Vault

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Ecosystem](https://img.shields.io/badge/Hydra-Ecosystem-00f0ff.svg)](https://github.com/eduardomdalmaso)

> **Curate, Deduplicate, Auto-Annotate, and Version-Control Vision Datasets at Scale for YOLO Architectures.**

---

## 🌟 Overview

**HydraVault** is the active learning and intelligent dataset curation backbone of the Hydra Vision Ecosystem. It closes the continuous learning loop between real-time camera ingestion (**HydraStream**) and high-throughput GPU model training (**HydraForge**).

```text
       ┌────────────────────────────────────────────────────────┐
       │                 1. HYDRASTREAM                         │
       │  (Zero-Copy RTSP Ingestion & TensorRT Live Inference)  │
       └──────────────────────────┬─────────────────────────────┘
                                  │
                   (Ambiguous Frames / Hard Negatives)
                                  │
                                  ▼
       ┌────────────────────────────────────────────────────────┐
       │                  2. HYDRAVAULT                         │
       │   (Data Curation, SAM 2 Auto-Labeling & Deduplication) │
       └──────────────────────────┬─────────────────────────────┘
                                  │
                   (Curated 100k+ Image Datasets)
                                  │
                                  ▼
       ┌────────────────────────────────────────────────────────┐
       │                  3. HYDRAFORGE                         │
       │   (YOLO PyTorch Studio, Benchmarking & TensorRT Export)│
       └──────────────────────────┬─────────────────────────────┘
                                  │
                 (Optimized .engine Model Checkpoint)
                                  │
                                  └──────────────► [ Deployed to HydraStream! ]
```

---

## 🚀 Core Features

* **📥 Zero-Friction Edge Harvesting:** Ingests difficult edge cases and low-confidence frames from HydraStream inference streams.
* **🤖 AI-Assisted Smart Pre-Annotation:** Zero-shot and few-shot auto-labeling using foundation models (**SAM 2** & **YOLO-World**) with 1-click human verification.
* **🔍 Perceptual Deduplication:** Removes visual redundancy and static frames using Perceptual Hashing (pHash) and cosine embedding distances.
* **⚖️ Dataset Health & Class Balance Auditing:** Real-time analytics on class distribution, small-object ratios, and bounding box integrity.
* **📦 Version-Controlled Splits:** Generates reproducible `data.yaml` splits (`train`, `val`, `test`) ready for immediate consumption by **HydraForge**.

---

## 🛠️ Quick Start

```bash
# Clone the repository
git clone https://github.com/eduardomdalmaso/HydraVault.git
cd HydraVault

# Build the control plane binary
make build

# Run the server
make run
```

---

## 📄 License
This project is licensed under the Apache 2.0 License.

# 🏛️ HydraVault — Motor de Curadoria de Dados & Active Learning para Visão Computacional

> **Curadoria, Desduplicação, Auto-Rotulagem e Controle de Versão de Datasets em Escala para Arquiteturas YOLO.**

---

## 🌟 Visão Geral

O **HydraVault** é o cérebro de curadoria e *Active Learning* do ecossistema Hydra. Ele fecha o ciclo de aprendizado contínuo conectando a ingestão de vídeo em tempo real (**HydraStream**) ao laboratório de treinamento em GPU (**HydraForge**).

---

## 🚀 Pilares Principais

1. **📥 Coleta Inteligente de Borda (Edge Harvesting):** Captura automática de frames desafiadores e com baixa confiança de inferência vindos do HydraStream.
2. **🤖 Auto-Rotulagem com Modelos de Fundação:** Pré-anotação assistida por IA usando **SAM 2** e **YOLO-World** com aprovação humana em 1 clique.
3. **🔍 Desduplicação Perceptual:** Eliminação de frames estáticos e redundantes via Perceptual Hashing e distância de embeddings.
4. **⚖️ Auditoria de Saúde do Dataset:** Diagnóstico em tempo real de desbalanceamento de classes, caixas nulas e objetos pequenos.
5. **📦 Exportação 1-Clique para o HydraForge:** Criação padronizada de splits versionados com `data.yaml`.

---

## 🛠️ Início Rápido

```bash
# Compilar o binário
make build

# Executar o servidor
make run
```

# 🔒 Diretrizes Invioláveis de Segurança & Hardening (HydraVault)

Este documento estabelece as diretrizes invioláveis de segurança, autenticação de curadoria SAM 2, isolamento de frames e anotações para o HydraVault.

---

## 1. Isolamento Multi-Tenant de Dados de Curadoria
- **Isolamento de Datasets & Inboxes:** Frames coletados via Active Learning do HydraStream são roteados para inboxes isoladas por `tenant_id`.
- **Integridade de Anotações:** Anotações e máscaras SAM 2 são vinculadas exclusivamente ao tenant dono da imagem.

---

## 2. Permissão no Servidor (Backend RBAC)
- **Controle de Curadoria e Aprovação:** Operações de auto-rotulagem em massa e exportação de `data.yaml` exigem autenticação com papel `admin` ou `annotator`.

---

## 3. Prevenção a IDOR & Exportação
- **Validação de Dataset por ID:** O download de splits versionados (`train/val/test`) valida a posse do dataset pelo token requisitante.

---

## 4. Chaves & Segredos
- **Credenciais MinIO S3 & Tokens de IA:** Segredos de acesso a buckets e chaves de modelos zero-shot são gerenciados via variáveis de ambiente.

---

## 5. Sanitização de Imagens e Arquivos
- **Validação de Mídia:** Upload e ingestão de frames validam cabeçalhos de imagem (JPEG/PNG/WebP) e impedem injeção de arquivos maliciosos ou caminhos com path traversal.

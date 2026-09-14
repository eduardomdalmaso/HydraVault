---
name: security-audit
description: Executa auditoria estrita de código em 5 categorias (Banco sem tranca, Permissão no navegador, IDOR, Chaves expostas, Inputs/XSS/LFI), gera relatório executivo em PDF com gráficos e issues completas para GitHub, e executa correções automatizadas.
---

# Security Audit Skill (Hydra Ecosystem)

Este skill define o protocolo padronizado de auditoria estrita de segurança e geração de relatórios PDF com gráficos e issues para o ecossistema Hydra.

---

## As 5 Categorias de Auditoria

1. **BANCO SEM TRANCA (Isolamento Multi-Tenant):**
   - Verificar se todas as queries SQL, buffers SHM ou coleções filtram estritamente por `tenant_id` extraído do token JWT (`r.Context()`).
   - Apontar qualquer bypass como `OR tenant_id = root` ou listagens globais sem WHERE.

2. **PERMISSÃO NO NAVEGADOR (Client-Side Enforcement vs Server RBAC):**
   - Cruzar gates de UI (`isAdmin`, `v-if`) com os endpoints correspondentes no backend.
   - Confirmar se o backend valida `RequireRole("admin")` em toda rota privilegiada de escrita ou configuração.

3. **IDOR (Insecure Direct Object Reference):**
   - Rastrear todos os endpoints com `{id}` (`DELETE`, `GET`, `PUT`).
   - Confirmar se a query/operação valida a posse do objeto pelo `tenant_id` do chamador.

4. **CHAVES EXPOSTAS & DEFAULTS INSEGUROS:**
   - Detectar fallbacks de chaves estáticas (`JWT_SECRET`, tokens S3, senhas padrão).
   - Auditar rotinas de startup que possam sobrescrever credenciais padrão.

5. **INPUTS SEM TRATAMENTO, ARBITRARY FILE READ, XSS & SSRF:**
   - No backend: Auditar `http.ServeFile` com `filepath.Clean` e restrição a diretórios seguros; auditar sondas de rede (SSRF) e injeção de comandos.
   - No frontend: Auditar `v-html`, `innerHTML`, URLs `javascript:` e dados dinâmicos.

---

## Procedimento de Execução do Relatório PDF

1. **Ambiente Isolado:** Utilizar a venv Python em `docs/security-audit/.venv`.
2. **Gerador Python:** Executar `docs/security-audit/generate_report.py`.
3. **Saída Obrigatória:**
   - Relatório PDF A4 formatado em `docs/security-audit/relatorio-auditoria-seguranca.pdf`.
   - Gráfico de rosca por severidade e gráfico de barras por categoria.
   - Seção de **ISSUES PARA O GITHUB** formatadas em blocos Markdown delimitados.

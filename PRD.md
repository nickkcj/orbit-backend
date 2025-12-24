# 🪐 Orbit - Product Requirements Document (MVP)

**Versão:** 1.0
**Data:** Dezembro 2025
**Status:** Em Desenvolvimento

---

## 1. Visão do Produto
O **Orbit** é uma plataforma *white-label* de cursos e comunidades.
Ao contrário da Udemy ou Hotmart (que funcionam como marketplaces), o Orbit atua como o **"Shopify dos Criadores"**: uma infraestrutura invisível que permite ao criador ter a sua própria escola online, com a sua marca, o seu domínio e controlo total sobre os seus alunos.

**Proposta de Valor:** Soberania para o criador + Experiência premium para o aluno.

---

## 2. Personas (Quem usa?)

### 👑 O Criador (Tenant / Admin)
* **Perfil:** Especialista que vende conhecimento (devs, designers, finanças) e quer fugir das taxas e da falta de identidade das grandes plataformas.
* **Dores:** "A Hotmart parece a Hotmart, não a minha escola", "Não tenho acesso aos dados dos meus alunos", "As taxas são altas".
* **Objetivo:** Criar um ambiente bonito, hospedar vídeos e engajar a comunidade num só lugar.

### 🎓 O Aluno (Member / End-User)
* **Perfil:** Pessoa que comprou o curso/acesso.
* **Dores:** Players de vídeo ruins, dificuldade em encontrar conteúdo, solidão (estudar sozinho).
* **Objetivo:** Assistir às aulas sem travamentos e tirar dúvidas com outros alunos e com o professor.

---

## 3. Escopo do MVP (O que entra na V1)

O sistema é dividido em **3 Grandes Módulos**:

### A. Orbit Core (Backend & Infra)
* **Multi-tenancy:** O sistema deve suportar milhares de comunidades isoladas (lógica de `tenant_id` em todas as tabelas).
* **Performance:** Backend em **Go (Golang)** para suportar milhares de conexões simultâneas (chat/vídeo) com baixo custo de infraestrutura.
* **Vídeo:** Upload direto (*Direct Upload*) para armazenamento barato (Cloudflare R2 ou AWS S3) com streaming seguro.

### B. Orbit Studio (Painel do Criador)
* **Dashboard:** Visão geral de novos membros e métricas de engajamento.
* **Course Builder:** Interface *drag-and-drop* para criar Módulos e Aulas.
* **Gestão de Uploads:** Área para subir vídeos e anexos.
* **Personalização:** Configuração de Cores (Hex), Upload de Logo e Nome da Comunidade.

### C. Orbit Classroom (Área do Aluno)
* **Player Imersivo:** Reprodução de vídeo com lista de aulas lateral (Sidebar).
* **Comunidade Contextual:** Sistema de comentários (threads/fórum) posicionado logo abaixo do vídeo para maximizar a interação.
* **Progresso:** Marcação automática de "Aula Concluída".

---

## 4. Requisitos Funcionais (Telas Chave)

| Tela / Funcionalidade | Descrição | Status |
| :--- | :--- | :--- |
| **Login/Signup** | Autenticação global. O utilizador pode ter conta em várias comunidades Orbit. | 🔄 Em andamento |
| **Tenant Dashboard** | Métricas simples e atalhos para criar conteúdo. | 📝 A Fazer |
| **Course Editor** | Fluxo: Criar Módulo -> Criar Aula -> Upload de Vídeo. | 📝 A Fazer |
| **Settings** | Upload de Logo e definição de Cor Primária (Theme). | 📝 A Fazer |
| **Student Home** | Feed de novidades e lista de cursos nos quais está matriculado. | 📝 A Fazer |
| **Video Player** | A "Joia da Coroa". Vídeo + Chat + Navegação Lateral. | 📝 A Fazer |

---

## 5. Stack Tecnológica

* **Backend:** Go (Golang) + Echo Framework.
* **Banco de Dados:** PostgreSQL (Railway) + SQLC (Type-safe queries).
* **Frontend:** Next.js 14+ (App Router).
* **Estilização:** Tailwind CSS + shadcn/ui.
* **State Management:** React Query (TanStack Query v5).
* **Infraestrutura:** Railway (Deploy) + Cloudflare R2 (Armazenamento de Vídeos - Futuro).

---

## 6. O Que NÃO É (Anti-Escopo)

* ❌ **Não é um Marketplace:** Não haverá uma página pública de "Pesquisar Cursos" global.
* ❌ **Não é uma Rede Social Aberta:** Não é um Twitter/X. O foco é estritamente conteúdo educacional e comunidade fechada.
* ❌ **Não tem Gateway de Pagamento Nativo (na V1):** Inicialmente, o criador vende externamente (Eduzz, Kiwify, Hotmart) e o Orbit apenas libera o acesso via convite ou webhook. (A integração nativa com Stripe está planeada para a V2).

---

## 7. Prioridade Atual (Roadmap Imediato) 🚨

O foco atual é conectar o Frontend (Next.js) ao Backend (Go) para validar o ciclo de vida do Tenant:

1.  Criar um Tenant (Comunidade).
2.  Criar um Utilizador (Admin).
3.  Esse utilizador conseguir fazer login e visualizar o Dashboard da sua comunidade.
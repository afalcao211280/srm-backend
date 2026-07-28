# Criptografia — Detalhes além do baseline

O **baseline** (`security-baseline.md`) já exige: TLS 1.3+, HSTS, encryption at rest com chaves em KMS/Vault + rotação, nunca roll your own crypto, bcrypt(cost≥12)/argon2id para senhas. Este arquivo detalha **escolhas de algoritmo e armadilhas de implementação**.

## Escolha de algoritmo

| Uso | Escolha | Parâmetros | Por quê |
|---|---|---|---|
| Simétrico | **AES-256-GCM** | IV/nonce de 96 bits (12 bytes) | AEAD — cifra + autentica. Nunca CBC/ECB (sem integridade; ECB vaza padrões) |
| Simétrico alt | ChaCha20-Poly1305 | nonce 96 bits | AEAD, melhor em CPUs sem AES-NI (mobile) |
| Assimétrico | **RSA-4096** (OAEP) ou ECC | padding OAEP + MGF1-SHA256 | RSA-2048 mínimo legado; OAEP evita PKCS#1 v1.5 (Bleichenbacher) |
| Derivação de chave de senha | **Argon2id** ou PBKDF2 | PBKDF2: ≥100k iterações + SHA-512 | KDF lento resiste a brute-force |

## Armadilhas de implementação (causam falha real)

- **IV/nonce único por operação** — gerar com CSPRNG (`crypto.randomBytes`/`os.urandom`) a cada `encrypt`. Reuso de nonce em GCM quebra confidencialidade e integridade.
- **Authentication tag (AEAD) é obrigatório** — extrair na cifragem, verificar na decifragem. Sem verificar a tag não há proteção contra adulteração.
- **Formato de armazenamento**: concatenar `IV || ciphertext || tag` com offsets fixos (ou campos separados). Documentar a ordem.
- **Comparação timing-safe** — usar `crypto.timingSafeEqual`/`hmac.compare_digest` para comparar hashes/tokens. `==` vaza informação por tempo.
- **Salt único por hash de senha** — embutido no output do argon2/bcrypt; nunca salt global compartilhado.
- **Envelope encryption** — cifrar dados com DEK (data key); cifrar a DEK com KEK no KMS. Rotaciona KEK sem re-cifrar dados.

## Encryption at rest — PostgreSQL

- **pgcrypto**: `pgp_sym_encrypt(data, key)` / `pgp_sym_decrypt(col, key)`; coluna `BYTEA`. Cifragem em nível de coluna — aplicar só nas colunas sensíveis (ex: CPF, cartão), deixar o resto em claro para indexação/busca.
- **Cifragem transparente via trigger** `BEFORE INSERT/UPDATE` lendo a chave de `current_setting('app.encryption_key')` (chave no contexto da sessão, não no schema).
- Alternativa: TDE no nível de storage/volume (gerenciado pelo provedor) para "todo o disco"; pgcrypto para granularidade por coluna.

## TLS/SSL hardening

- `minVersion: TLSv1.2`, `maxVersion: TLSv1.3` — exclui TLS 1.0/1.1 explicitamente.
- **Ordem de cipher suites** preferindo: `TLS_AES_256_GCM_SHA384` → `TLS_CHACHA20_POLY1305_SHA256` → `TLS_AES_128_GCM_SHA256`; `honorCipherOrder: true` (servidor decide, não o cliente).
- **mTLS**: `requestCert: true` + `rejectUnauthorized: true`; verificar o cert do peer (`getPeerCertificate`). Em microsserviços, delegar ao service mesh (Istio/Linkerd) que gerencia e rotaciona os certificados.

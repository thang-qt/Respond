CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);
CREATE UNIQUE INDEX idx_password_reset_tokens_token_hash ON password_reset_tokens (token_hash);
CREATE UNIQUE INDEX idx_email_verification_tokens_token_hash ON email_verification_tokens (token_hash);

# Backend Scripts

## generate-encryption-key.sh

Generates a secure 32-byte (256-bit) encryption key for encrypting Strava OAuth tokens.

### Usage

```bash
./generate-encryption-key.sh
```

### Output

```
Generating secure 32-byte encryption key...

Your encryption key (add to .env):
ENCRYPTION_KEY=abcd1234efgh5678ijkl9012mnop3456qrst7890uvwx==

IMPORTANT: Keep this key secure and never commit it to version control!
If you lose this key, you won't be able to decrypt existing tokens.
```

### Requirements

- OpenSSL (installed by default on most Unix systems)
- Bash shell

### What It Does

1. Uses OpenSSL to generate 32 random bytes
2. Encodes the bytes to base64 for easy storage
3. Displays the key in a format ready to copy to `.env`

### Security Notes

- **Never commit** the generated key to git
- **Use different keys** for development and production
- **Store production keys** in secure secret management systems
- **Keep a backup** of your production key in a secure location
- If you lose the key, you **cannot decrypt** existing tokens

### After Generation

1. Copy the `ENCRYPTION_KEY=...` line
2. Add it to your `.env` file
3. Restart your application
4. Do NOT commit the `.env` file

### Key Requirements

- Must be exactly 32 bytes (256 bits)
- Must be cryptographically random (don't create manually)
- Should be unique per environment
- Should be rotated periodically (see ENCRYPTION.md)

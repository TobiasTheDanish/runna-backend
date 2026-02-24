#!/bin/bash

# Generate a secure 32-byte (256-bit) encryption key
echo "Generating secure 32-byte encryption key..."
echo ""

# Generate random bytes and encode to base64
KEY=$(openssl rand -base64 32)

echo "Your encryption key (add to .env):"
echo "ENCRYPTION_KEY=$KEY"
echo ""
echo "IMPORTANT: Keep this key secure and never commit it to version control!"
echo "If you lose this key, you won't be able to decrypt existing tokens."

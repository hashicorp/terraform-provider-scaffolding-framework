#!/usr/bin/env bash
set -euo pipefail

echo "=== gpg key generation and github secrets upload ==="
echo ""

# Check if gh is installed
if ! command -v gh &> /dev/null; then
  echo "error: gh cli is not installed"
  echo "install it with: brew install gh"
  exit 1
fi
echo "✓ gh cli is installed"

# Check if authenticated
if ! gh auth status &> /dev/null; then
  echo "error: not authenticated with github"
  echo "run: gh auth login"
  exit 1
fi
echo "✓ authenticated with github"

# Get current repo
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
echo "✓ repository: $REPO"

# Get git config
GIT_NAME=$(git config --global user.name || echo "")
GIT_EMAIL=$(git config --global user.email || echo "")

if [ -z "$GIT_NAME" ] || [ -z "$GIT_EMAIL" ]; then
  echo "error: git user.name or user.email not configured"
  echo "run: git config --global user.name 'Your Name'"
  echo "run: git config --global user.email 'your@email.com'"
  exit 1
fi

echo "✓ git user: $GIT_NAME <$GIT_EMAIL>"
echo ""

# Generate random passphrase
echo "=== generating passphrase ==="
echo "generating 32 random characters..."
PASSPHRASE=$(head -c 256 /dev/urandom | LC_ALL=C tr -dc 'A-Za-z0-9' | head -c 32)
echo "writing to password.txt..."
echo -n "$PASSPHRASE" > password.txt
echo "✓ passphrase generated and saved to password.txt (${#PASSPHRASE} characters)"
echo ""

# Create temporary GPG home
TEMP_GPG_HOME=$(mktemp -d)
trap "rm -rf $TEMP_GPG_HOME" EXIT
export GNUPGHOME=$TEMP_GPG_HOME
chmod 700 "$GNUPGHOME"

# Generate GPG key
echo "=== generating gpg key ==="
echo "(this may take a moment...)"
gpg --batch --gen-key <<EOF
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: $GIT_NAME
Name-Email: $GIT_EMAIL
Name-Comment: Github Actions for terraform provider: $REPO
Expire-Date: 0
Passphrase: $PASSPHRASE
%commit
EOF

echo "✓ gpg key generated"

# Get fingerprint
GPG_FINGERPRINT=$(gpg --with-colons --list-secret-keys | grep '^fpr:' | head -1 | cut -d: -f10)
echo "✓ fingerprint: $GPG_FINGERPRINT"
echo ""

# Export keys
echo "=== exporting gpg keys ==="
gpg --batch --pinentry-mode loopback --passphrase "$PASSPHRASE" --armor --export-secret-keys "$GPG_FINGERPRINT" > secret-key.pem
echo "✓ private key exported to secret-key.pem ($(wc -c < secret-key.pem | tr -d ' ') bytes)"

gpg --batch --armor --export "$GPG_FINGERPRINT" > public-key.asc
echo "✓ public key exported to public-key.asc ($(wc -c < public-key.asc | tr -d ' ') bytes)"
echo ""

# Validate the exported key works
echo "=== validating exported key ==="
TEMP_TEST_HOME=$(mktemp -d)
trap "rm -rf $TEMP_GPG_HOME $TEMP_TEST_HOME" EXIT
GNUPGHOME=$TEMP_TEST_HOME gpg --batch --import secret-key.pem 2>&1 | grep -q "imported"
echo "test content" > test.txt
GNUPGHOME=$TEMP_TEST_HOME gpg --batch --yes --pinentry-mode loopback --passphrase "$PASSPHRASE" --local-user "$GPG_FINGERPRINT" --detach-sign test.txt
rm -f test.txt test.txt.sig
echo "✓ key and passphrase validated successfully"
echo ""

# Upload secrets to GitHub
echo "=== uploading secrets to github ==="
echo "uploading PASSPHRASE secret..."
gh secret set PASSPHRASE < password.txt
echo "✓ PASSPHRASE uploaded"

echo ""
echo "uploading GPG_PRIVATE_KEY secret..."
gh secret set GPG_PRIVATE_KEY < secret-key.pem
echo "✓ GPG_PRIVATE_KEY uploaded"

echo ""
echo "=== success ==="
echo "✓ new gpg key generated"
echo "✓ secrets uploaded to $REPO"
echo ""
echo "files created:"
echo "  - password.txt (passphrase)"
echo "  - secret-key.pem (gpg private key)"
echo "  - public-key.asc (gpg public key)"
echo ""
echo "verify github secrets:"
echo "  gh secret list"
echo ""
echo "=========================================================================="
echo "terraform registry setup"
echo "=========================================================================="
echo ""
echo "to publish this provider to terraform registry, you need to add the"
echo "public gpg key to your terraform cloud organization."
echo ""
echo "1. get your organization name from terraform cloud"
echo "2. go to: https://app.terraform.io/app/organizations/<your-org>/settings/gpg-keys"
echo "3. click 'Add a GPG key'"
echo "4. paste the content of: public-key.asc"
echo ""
echo "to view the public key now:"
echo "  cat public-key.asc"
echo ""
echo "key details:"
echo "  fingerprint: $GPG_FINGERPRINT"
echo "  name: $GIT_NAME"
echo "  email: $GIT_EMAIL"
echo ""
echo "=========================================================================="
echo "next steps"
echo "=========================================================================="
echo ""
echo "1. add public key to terraform cloud (see above)"
echo "2. create a new git tag to trigger the release workflow:"
echo "   git tag v1.0.0"
echo "   git push origin v1.0.0"
echo "3. monitor the release workflow in github actions"
echo "4. once released, the provider will be signed with your gpg key"

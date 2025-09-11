#!/bin/bash

# Script to verify certificate chain completeness
# Usage: ./verify_cert_chain.sh <cert_file_path>

if [ $# -eq 0 ]; then
    echo "Usage: $0 <cert_file_path>"
    echo "Example: $0 /ipssl/cert.pem"
    exit 1
fi

CERT_FILE="$1"

if [ ! -f "$CERT_FILE" ]; then
    echo "Error: Certificate file '$CERT_FILE' not found"
    exit 1
fi

echo "=== Certificate Chain Analysis ==="
echo "File: $CERT_FILE"
echo

# Count certificates in the file
CERT_COUNT=$(grep -c "-----BEGIN CERTIFICATE-----" "$CERT_FILE")
echo "Number of certificates found: $CERT_COUNT"

if [ "$CERT_COUNT" -eq 0 ]; then
    echo "❌ No certificates found in file"
    exit 1
elif [ "$CERT_COUNT" -eq 1 ]; then
    echo "⚠️  Only one certificate found - missing intermediate certificates"
    echo "   This may cause SSL/TLS validation issues"
elif [ "$CERT_COUNT" -eq 2 ]; then
    echo "✅ Two certificates found - likely includes intermediate certificate"
elif [ "$CERT_COUNT" -gt 2 ]; then
    echo "✅ $CERT_COUNT certificates found - complete certificate chain"
fi

echo

# Extract and analyze each certificate
echo "=== Certificate Details ==="
CERT_NUM=1
while IFS= read -r line; do
    if [[ "$line" == "-----BEGIN CERTIFICATE-----" ]]; then
        echo "Certificate #$CERT_NUM:"
        
        # Extract this certificate to a temporary file
        TEMP_CERT=$(mktemp)
        echo "$line" > "$TEMP_CERT"
        
        # Read until we find the end marker
        while IFS= read -r next_line; do
            echo "$next_line" >> "$TEMP_CERT"
            if [[ "$next_line" == "-----END CERTIFICATE-----" ]]; then
                break
            fi
        done
        
        # Analyze the certificate
        if command -v openssl >/dev/null 2>&1; then
            echo "  Subject: $(openssl x509 -in "$TEMP_CERT" -noout -subject 2>/dev/null | sed 's/subject=//')"
            echo "  Issuer:  $(openssl x509 -in "$TEMP_CERT" -noout -issuer 2>/dev/null | sed 's/issuer=//')"
            echo "  Valid:   $(openssl x509 -in "$TEMP_CERT" -noout -dates 2>/dev/null | grep notAfter | sed 's/notAfter=//')"
        else
            echo "  (OpenSSL not available for detailed analysis)"
        fi
        
        rm -f "$TEMP_CERT"
        echo
        CERT_NUM=$((CERT_NUM + 1))
    fi
done < "$CERT_FILE"

echo "=== Summary ==="
if [ "$CERT_COUNT" -ge 2 ]; then
    echo "✅ Certificate chain appears complete"
    echo "   The certificate file contains both the server certificate and intermediate certificate(s)"
else
    echo "❌ Certificate chain is incomplete"
    echo "   Missing intermediate certificates may cause SSL/TLS validation failures"
    echo "   Make sure the ZeroSSL client is downloading the full certificate bundle"
fi

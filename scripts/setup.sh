#!/bin/bash

# IPSSL Client Setup Script
# This script sets up the required directories and files for the IPSSL client

set -e

echo "Setting up IPSSL Client..."

# Create required directories
echo "Creating directories..."
mkdir -p docker/data/caddy/{data,config,logs,webroot,ipssl}
mkdir -p docker/config/caddy
mkdir -p bin

# Set proper permissions
echo "Setting permissions..."
chmod 755 docker/data/caddy/webroot
chmod 755 docker/data/caddy/ipssl
chmod 755 docker/data/caddy/logs

# Create .well-known directory for ACME validation
mkdir -p docker/data/caddy/webroot/.well-known/pki-validation
chmod 755 docker/data/caddy/webroot/.well-known/pki-validation

# Copy example files if they don't exist
if [ ! -f docker/.env ]; then
    echo "Creating .env file from example..."
    cp docker/env.example docker/.env
    echo "Please edit docker/.env file with your configuration"
fi

if [ ! -f docker/config/caddy/Caddyfile ]; then
    echo "Creating Caddyfile from example..."
    if [ -f docker/config/caddy/Caddyfile.example ]; then
        cp docker/config/caddy/Caddyfile.example docker/config/caddy/Caddyfile
        echo "Please review and customize docker/config/caddy/Caddyfile"
    else
        echo "Creating default Caddyfile..."
        cat > docker/config/caddy/Caddyfile << 'EOF'
# Default Caddyfile for IPSSL
# This will be configured automatically by the IPSSL client

:80 {
    redir https://{host}{uri} permanent
}

:443 {
    tls /ipssl/cert.pem /ipssl/key.pem
    root * /usr/share/caddy
    file_server
}
EOF
        echo "Created default Caddyfile at docker/config/caddy/Caddyfile"
    fi
fi

# Create a simple index.html for testing
if [ ! -f docker/data/caddy/webroot/index.html ]; then
    echo "Creating test index.html..."
    cat > docker/data/caddy/webroot/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>IPSSL Test Page</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .success { background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
    </style>
</head>
<body>
    <div class="container">
        <h1>IPSSL Test Page</h1>
        <div class="status success">
            <strong>SSL Certificate Status:</strong> Active
        </div>
        <p>This page is served over HTTPS using an IP-based SSL certificate managed by IPSSL Client.</p>
        <p><strong>Server IP:</strong> <span id="server-ip">Loading...</span></p>
        <p><strong>Certificate Info:</strong> <span id="cert-info">Loading...</span></p>
        
        <script>
            // Display server IP
            document.getElementById('server-ip').textContent = window.location.hostname;
            
            // Display certificate info
            if (window.location.protocol === 'https:') {
                document.getElementById('cert-info').textContent = 'Valid HTTPS Certificate';
            } else {
                document.getElementById('cert-info').textContent = 'HTTP (No Certificate)';
            }
        </script>
    </div>
</body>
</html>
EOF
fi

echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit docker/.env file with your ZeroSSL API key and IP address"
echo "2. Review and customize docker/config/caddy/Caddyfile if needed"
echo "3. Run the following commands to start the services:"
echo "   cd docker"
echo "   docker-compose up -d"
echo ""
echo "The application will be available at:"
echo "- HTTP: http://your-ip:80 (redirects to HTTPS)"
echo "- HTTPS: https://your-ip:443"
echo ""
echo "To check logs:"
echo "   cd docker"
echo "   docker-compose logs -f"

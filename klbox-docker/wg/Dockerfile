FROM ghcr.io/kloudlite/hub/wireguard:latest

# Install doas (instead of sudo)
RUN apk add --no-cache doas

# Create user 1000 and configure doas
RUN adduser -D -u 1000 user && \
    echo "permit nopass user as root" > /etc/doas.d/doas.conf && \
    chmod 600 /etc/doas.d/doas.conf

# Switch to user
USER 1000

# Set working directory
WORKDIR /home/user

# Default command
CMD ["/bin/sh"]

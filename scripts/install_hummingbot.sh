#!/bin/bash
set -eo pipefail

echo "Installing Docker and required dependencies..."
# Update package list and install required dependencies
sudo apt-get update
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# Add Docker's official GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up Docker repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine and Docker Compose
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add current user to docker group to avoid using sudo
sudo usermod -aG docker $USER

echo "Cloning Hummingbot repository..."
git clone https://github.com/hummingbot/hummingbot.git
cd hummingbot

echo "Starting Hummingbot with Docker Compose..."
docker compose up -d

echo "Installation complete! Please log out and log back in for docker group changes to take effect."
echo "Hummingbot is now running in the background." 
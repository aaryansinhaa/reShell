# Create the user-friendly sandbox demo container
FROM ubuntu:24.04

# Install basic packages required for shell environments and demo actions
RUN apt-get update && apt-get install -y \
    bash \
    zsh \
    git \
    curl \
    neovim \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Copy the built reshell binary
COPY reshell /usr/local/bin/reshell

# Create a non-root developer user with sudo privileges
RUN useradd -m -s /bin/bash developer && \
    echo "developer ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER developer
WORKDIR /home/developer

# Pre-initialize Git so git version checks and history tracking work immediately
RUN git config --global user.name "Developer Sandbox" && \
    git config --global user.email "sandbox@reshell.dev"

# Set preferred editor environment
ENV EDITOR=nvim

# Set the default entrypoint to give them a shell session directly
CMD ["/bin/bash"]

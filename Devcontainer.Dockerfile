FROM golang:alpine

RUN apk add --no-cache bash \
    curl \
    git \
    make \
    jq

# Install kubectl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
RUN chmod +x ./kubectl
RUN mv ./kubectl /usr/local/bin

# Install Flux CLI
RUN FLUX_LATEST=$(curl -s https://api.github.com/repos/fluxcd/flux2/releases/latest | jq -r .tag_name) && \
    FLUX_VERSION=${FLUX_LATEST#v} && \
    curl -L https://github.com/fluxcd/flux2/releases/download/${FLUX_LATEST}/flux_${FLUX_VERSION}_linux_amd64.tar.gz -o flux.tar.gz && \
    tar -xzf flux.tar.gz && \
    mv flux /usr/local/bin/ && \
    rm flux.tar.gz

# to make the build succeed
RUN git config --global --add safe.directory /workspaces/glab

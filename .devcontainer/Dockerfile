FROM node:26.10.0-bookworm AS node

FROM golang:1.27

COPY --from=node /usr/local/ /usr/local/

RUN apt-get update && apt-get install -y \
    curl \
    git \
    make

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID --create-home --shell /bin/bash $USERNAME

WORKDIR /workspace

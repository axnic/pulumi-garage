FROM mcr.microsoft.com/devcontainers/base:bookworm

# mise provisions every other tool (go, pulumi, golangci-lint, ...) from this
# repo's own .mise.toml / .config/mise.toml, so the devcontainer, CI, and any
# other local setup all resolve to the exact same versions - see
# postCreateCommand in devcontainer.json, which runs `mise install`.
USER vscode
RUN curl -fsSL https://mise.run | sh
ENV PATH="/home/vscode/.local/bin:/home/vscode/.local/share/mise/shims:${PATH}"
RUN echo 'eval "$(mise activate bash)"' >> /home/vscode/.bashrc

# .config/mise.toml pulls in the community vfox-pulumi plugin (to derive
# GO_VERSION_MISE/PULUMI_VERSION_MISE from go.mod/.pulumi.version). mise
# evaluates that config on every invocation once the workspace is mounted,
# including plugin-install itself, so the plugin has to go in now, from a
# directory with no mise config in scope - otherwise `mise install` fails
# with "[vfox-pulumi] plugin not installed" no matter what you run it with.
RUN cd /tmp && mise plugins install vfox-pulumi https://github.com/pulumi/vfox-pulumi

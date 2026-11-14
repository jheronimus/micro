# micro (fork)

This is a personal fork of the [Micro](https://github.com/micro-editor/micro) text editor that extends its Lua plugin runtime with two minimal primitives: a floating window overlay system (`micro.ShowOverlay`) and an in-buffer visual line transform hook (`micro.SetLineTransform`). These primitives enable rich IDE-like features—such as top-right LSP hover documentation overlays and VS Code / Obsidian-style visual Markdown styling with syntax concealment—without cluttering the editor's core codebase. Releases are automatically built and published via GitHub Actions.

# WebGAL Snapshot

WebGAL 快照小工具 - 一键打包 WebGAL 项目, 自动裁剪未使用的资源.

## :sparkles: 特性

- **开箱即用**: 无需任何依赖, 下载即运行.

- **精准裁剪**: 自动解析全部游戏场景, 仅打包实际用到的资源, 减小体积.

- **批量处理**: 支持同时传入多个项目路径, 依次打包.

## :rocket: 使用

### 直接运行

以命令行参数形式传入一个或多个 WebGAL 项目路径, 工具会在同级目录下生成对应的 `<游戏名称>.zip` 压缩包.

> [!TIP]
>
> 也可以在文件管理器中选中项目, 将其拖拽到可执行文件上.

### 自行构建

```bash
go build cmd/webgal-snapshot
```

## :package: 资源打包

### 自动打包

- 引擎资源[^1].

- 场景中引用的普通本地资源.

- 场景中引用的 Live2D, WMDL 模型及关联资源.

### 尚不支持

- 场景中引用的 Spine, JSONL 模型及关联资源.

## :page_facing_up: 许可证

Code: GPL-3.0, 2026, fltLi

[^1]: `assets/`, `icons/`, `lib/`, `index.html`, `manifest.json`, `webgal-serviceworker.js`.
